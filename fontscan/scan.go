package fontscan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/bitmap"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fonts/type1"
)

// DefaultFontDirs return the OS-dependent usual directories for
// fonts, or an error if no one exists.
func DefaultFontDirs() ([]string, error) {
	var dirs []string
	switch runtime.GOOS {
	case "windows":
		sysRoot := os.Getenv("SYSTEMROOT")
		if sysRoot == "" {
			sysRoot = os.Getenv("SYSTEMDRIVE")
		}
		if sysRoot == "" { // try with the common C:
			sysRoot = "C:"
		}
		dir := filepath.Join(filepath.VolumeName(sysRoot), `\Windows`, "Fonts")
		dirs = []string{
			dir,
			filepath.Join(os.Getenv("windir"), "Fonts"),
			filepath.Join(os.Getenv("localappdata"), "Microsoft", "Windows", "Fonts"),
		}
	case "darwin":
		dirs = []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
			"~/Library/Fonts",
			"/Network/Library/Fonts",
			"/System/Library/Assets/com_apple_MobileAsset_Font3",
			"/System/Library/Assets/com_apple_MobileAsset_Font4",
			"/System/Library/Assets/com_apple_MobileAsset_Font5",
		}
	case "linux":
		dirs = []string{
			"/usr/share/fonts",
			"/usr/share/texmf/fonts/opentype/public",
		}

		if dataPath := os.Getenv("XDG_DATA_HOME"); dataPath != "" {
			dirs = append(dirs, "~/.fonts/", filepath.Join(dataPath, "fonts"))
		} else {
			dirs = append(dirs, "~/.fonts/", "~/.local/share/fonts/")
		}

		if dataPaths := os.Getenv("XDG_DATA_DIRS"); dataPaths != "" {
			for _, dataPath := range filepath.SplitList(dataPaths) {
				dirs = append(dirs, filepath.Join(dataPath, "fonts"))
			}
		}
	case "android":
		dirs = []string{
			"/system/fonts",
			"/system/font",
			"/data/fonts",
		}
	case "ios":
		dirs = []string{
			"/System/Library/Fonts",
			"/System/Library/Fonts/Cache",
		}
	default:
		return nil, fmt.Errorf("unsupported plaform %s", runtime.GOOS)
	}

	var validDirs []string
	for _, dir := range dirs {
		dir = expandUser(dir)

		info, err := os.Stat(dir)
		if err != nil { // ignore the non existent directory
			continue
		}
		if !info.IsDir() {
			log.Println("font dir is not a directory", dir)
			continue
		}
		validDirs = append(validDirs, dir)
	}
	if len(validDirs) == 0 {
		return nil, errors.New("no font directory found")
	}

	return validDirs, nil
}

func expandUser(path string) (expandedPath string) {
	if strings.HasPrefix(path, "~") {
		if u, err := user.Current(); err == nil {
			return strings.Replace(path, "~", u.HomeDir, -1)
		}
	}
	return path
}

// rejects several extensions which are for sure not supported font files
// return `true` is the file should be ignored
func ignoreFontFile(name string) bool {
	// ignore hidden file
	if name == "" || name[0] == '.' {
		return true
	} else if strings.HasSuffix(name, ".enc.gz") || // encodings
		strings.HasSuffix(name, ".afm") || // metrics (ascii)
		strings.HasSuffix(name, ".pfm") || // metrics (binary)
		strings.HasSuffix(name, ".dir") || // summary
		strings.HasSuffix(name, ".scale") ||
		strings.HasSuffix(name, ".alias") {
		return true
	}

	return false
}

type fontFileHandler interface {
	consume(path string, info fs.FileInfo) error
}

// recursively walk through the given directory, scanning font files and calling dst.consume
// for each valid file found.
func scanDirectory(dir string, visited map[string]bool, dst fontFileHandler) error {
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking font directories: %s", err)
		}

		if d.IsDir() { // keep going
			return nil
		}

		if visited[path] {
			return nil // skip the path
		}
		visited[path] = true

		// load the information, following potential symoblic links
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		// always ignore files which should never be font files
		if ignoreFontFile(info.Name()) {
			return nil
		}

		err = dst.consume(path, info)

		return err
	}

	err := filepath.WalkDir(dir, walkFn)

	return err
}

// --------------------- footprint mode -----------------------

// try the different supported loader and returns the list of the fonts
// contained in `file`, with their format.
func getFontDescriptors(file fonts.Resource) ([]fonts.FontDescriptor, Format) {
	out, err := truetype.ScanFont(file)
	if err == nil {
		return out, OpenType
	}
	out, err = type1.ScanFont(file)
	if err == nil {
		return out, Type1
	}
	out, err = bitmap.ScanFont(file)
	if err == nil {
		return out, PCF
	}
	return nil, 0
}

// timeStamp is the unix modification time of a font file,
// used to trigger or not the scan of a font file
type timeStamp int64

func newTimeStamp(file os.FileInfo) timeStamp { return timeStamp(file.ModTime().UnixNano()) }

func (fh timeStamp) serialize() []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(fh))
	return buf[:]
}

// assume len(src) >= 8
func (fh *timeStamp) deserialize(src []byte) {
	*fh = timeStamp(binary.BigEndian.Uint64(src))
}

// systemFontsIndex stores the footprint comming from the file system
type systemFontsIndex []fileFootprints

func (sfi systemFontsIndex) flatten() []Footprint {
	var out []Footprint
	for _, file := range sfi {
		out = append(out, file.footprints...)
	}
	return out
}

// groups the footprints by origin file
type fileFootprints struct {
	path string // file path

	footprints []Footprint // font content for the path

	// modification time for the file
	modTime timeStamp
}

type footprintScanner struct {
	previousIndex map[string]fileFootprints // reference index, to be updated

	dst systemFontsIndex // accumulated footprints
}

func newFootprintAccumulator(currentIndex systemFontsIndex) footprintScanner {
	// map font files to their footprints
	out := footprintScanner{previousIndex: make(map[string]fileFootprints, len(currentIndex))}
	for _, fp := range currentIndex {
		out.previousIndex[fp.path] = fp
	}
	return out
}

func (fa *footprintScanner) consume(path string, info fs.FileInfo) error {
	modTime := newTimeStamp(info)

	// try to avoid scanning the file
	if indexedFile, has := fa.previousIndex[path]; has && indexedFile.modTime == modTime {
		// we already have an up to date scan of the file:
		// skip the scan and add the current footprints
		fa.dst = append(fa.dst, indexedFile)
		return nil
	}

	// do the actual scan

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	fontDescriptors, format := getFontDescriptors(file)
	ff := fileFootprints{
		path:    path,
		modTime: modTime,
	}

	for i, fd := range fontDescriptors {
		footprint, err := newFootprintFromDescriptor(fd, format)
		// the font won't be usable, just ignore it
		if err != nil {
			continue
		}

		footprint.Location.File = path
		footprint.Location.Index = uint16(i)
		// TODO: for now, we do not handle variable fonts

		ff.footprints = append(ff.footprints, footprint)
	}

	// note that newFootprintFromDescriptor may read from the file,
	// so that we should not close it earlier.
	file.Close()

	fa.dst = append(fa.dst, ff)

	return nil
}

// scanFontFootprints walk through the given directories
// and scan each font file to extract its footprint.
// An error is returned if the directory traversal fails, not for invalid font files,
// which are simply ignored.
// `currentIndex` may be passed to avoid scanning font files that are
// already present in `currentIndex` and up to date, and directly duplicating
// the footprint in `currentIndex`
func scanFontFootprints(currentIndex systemFontsIndex, dirs ...string) (systemFontsIndex, error) {
	// keep track of visited dirs to avoid double inclusions,
	// for instance with symbolic links
	visited := make(map[string]bool)

	accu := newFootprintAccumulator(currentIndex)
	for _, dir := range dirs {
		err := scanDirectory(dir, visited, &accu)
		if err != nil {
			return nil, err
		}
	}
	return accu.dst, nil
}

// --------------------- File name mode ------------------------------

type fileNameScanner []string // list of paths

// return the lower filename and ext
func splitAtDot(filePath string) (name, ext string) {
	filePath = filepath.Base(strings.ToLower(filePath))
	if i := strings.IndexByte(filePath, '.'); i != -1 {
		return filePath[:i], filePath[i:]
	}
	return name, ""
}

func isFontFile(fileName string) bool {
	_, ext := splitAtDot(fileName)
	switch ext {
	case ".ttf", ".ttc", ".otf", ".otc", ".woff", // Opentype
		".t1", ".pfb", // Type1
		".pcf.gz", ".pcf": // Bitmap
		return true
	default:
		return false
	}
}

func (fns *fileNameScanner) consume(path string, _ fs.FileInfo) error {
	if isFontFile(path) {
		*fns = append(*fns, path)
	}
	return nil
}

// returns a list of file path looking like font files
// no font loading is performed by this function
func scanFontFiles(dirs ...string) ([]string, error) {
	visited := make(map[string]bool)

	var dst fileNameScanner
	for _, dir := range dirs {
		err := scanDirectory(dir, visited, &dst)
		if err != nil {
			return nil, err
		}
	}

	return dst, nil
}
