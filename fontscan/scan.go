package fontscan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
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
		dirs = []string{dir}
	case "darwin":
		dirs = []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
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
		info, err := os.Stat(dir)
		if err != nil {
			log.Println("invalid font dir", dir, err)
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

type descriptorAccumulator interface {
	// it true, the font file at `path` wont be scanned
	skipFile(path string, modTime timeStamp) bool

	consume([]fonts.FontDescriptor, Format, string, timeStamp)
}

// recursively walk through the given directory, scanning font files and calling dst.consume
// for each valid file found.
func scanDirectory(dir string, visited map[string]bool, dst descriptorAccumulator) error {
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking font directories: %s", err)
		}

		if d.IsDir() { // keep going
			return nil
		}

		// evaluate symlinks before consulting seen,
		// since a symlink may point towards a directory already
		// included in the search directories
		if d.Type()&os.ModeSymlink != 0 {
			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
		}

		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		if visited[path] {
			if info.IsDir() { // optimize by entirely skipping the directory
				return filepath.SkipDir
			}
			return nil // just skip the file
		}
		visited[path] = true

		modTime := newTimeStamp(info)

		// always ignore files which should never be font files
		if ignoreFontFile(info.Name()) {
			return nil
		}

		// try to avoid scanning the file
		if dst.skipFile(path, modTime) {
			// keep going without scanning the file
			return nil
		}

		// do the actual scan

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		fontDescriptors, format := getFontDescriptors(file)

		dst.consume(fontDescriptors, format, path, modTime)

		// note that consume may read from the file,
		// so that we should not close it earlier.
		file.Close()

		return nil
	}

	err := filepath.WalkDir(dir, walkFn)

	return err
}

// groups the footprints by origin file
type fileFootprints struct {
	footprints []Footprint
	modTime    timeStamp
}

type footprintAccumulator struct {
	previousIndex map[string]fileFootprints

	dst []Footprint // accumulated footprints
}

func newFootprintAccumulator(currentIndex []Footprint) footprintAccumulator {
	// map font files to their modification time and footprints
	out := footprintAccumulator{previousIndex: make(map[string]fileFootprints)}
	for _, fp := range currentIndex {
		file := out.previousIndex[fp.Location.File]
		file.modTime = fp.modTime
		file.footprints = append(file.footprints, fp)
		out.previousIndex[fp.Location.File] = file
	}
	return out
}

func (fa *footprintAccumulator) skipFile(path string, modTime timeStamp) bool {
	if indexedFile, has := fa.previousIndex[path]; has && indexedFile.modTime == modTime {
		// we already have an up to date scan of the file:
		// skip the scan and add the current footprints
		fa.dst = append(fa.dst, indexedFile.footprints...)
		return true
	}

	// trigger the scan
	return false
}

func (fa *footprintAccumulator) consume(fds []fonts.FontDescriptor, format Format, path string, mod timeStamp) {
	for i, fd := range fds {
		footprint, err := newFootprintFromDescriptor(fd, format)
		// the font won't be usable, just ignore it
		if err != nil {
			continue
		}

		footprint.Location.File = path
		footprint.Location.Index = uint16(i)
		// TODO: for now, we do not handle variable fonts

		footprint.modTime = mod

		fa.dst = append(fa.dst, footprint)
	}
}

// ScanFonts walk through the given directories
// and scan each font file to extract its footprint.
// An error is returned if the directory traversal fails, not for invalid font files,
// which are simply ignored.
// `currentIndex` may be passed to avoid scanning font files that are
// already present in `currentIndex` and up to date, and directly duplicating
// the footprint in `currentIndex`
func ScanFonts(currentIndex []Footprint, dirs ...string) ([]Footprint, error) {
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
