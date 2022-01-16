package fontscan

import (
	"errors"
	"fmt"
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
	consume([]fonts.FontDescriptor, Format, string)
}

// recursively walk through the given directory, scanning font files and calling dst.consume
// for each valid file found.
func scanDirectory(dir string, seen map[string]bool, dst descriptorAccumulator) error {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("invalid font location: %s", err)
		}

		if info.IsDir() { // keep going
			if seen[path] {
				return filepath.SkipDir
			}
			seen[path] = true
			return nil
		}

		if ignoreFontFile(info.Name()) {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		fds, format := getFontDescriptors(file)

		// note that consume may read from the file,
		// so that we should not close it before calling it.
		dst.consume(fds, format, path)

		file.Close()

		return nil
	}

	err := filepath.Walk(dir, walkFn)

	return err
}

type familyAccumulator []string

func (fa *familyAccumulator) consume(fds []fonts.FontDescriptor, _ Format, _ string) {
	for _, fd := range fds {
		*fa = append(*fa, fd.Family())
	}
}

// ScanFamilies walk through the given directories
// and scan each font file to extract the font family.
// An error is returned if the directory traversal fails, not for invalid font files,
// which are simply ignored.
// TODO: return only one matching a given criterion
func ScanFamilies(dirs ...string) ([]string, error) {
	seen := make(map[string]bool) // keep track of visited dirs to avoid double inclusions
	var (
		accu familyAccumulator
		err  error
	)
	for _, dir := range dirs {
		err = scanDirectory(dir, seen, &accu)
		if err != nil {
			return nil, err
		}
	}
	return accu, nil
}

type footprintAccumulator []Footprint

func (fa *footprintAccumulator) consume(fds []fonts.FontDescriptor, format Format, path string) {
	for i, fd := range fds {
		footprint, err := newFootprintFromDescriptor(fd, format)
		// the font won't be usable, just warn and ignore it
		if err != nil {
			log.Println("unsupported font file", path, ":", err)
			continue
		}

		footprint.Location.Index = uint16(i)

		*fa = append(*fa, footprint)
	}
}

// ScanFonts walk through the given directories
// and scan each font file to extract its footprint.
// An error is returned if the directory traversal fails, not for invalid font files,
// which are simply ignored.
// TODO: handle Location and tree structure
func ScanFonts(dirs ...string) ([]Footprint, error) {
	seen := make(map[string]bool) // keep track of visited dirs to avoid double inclusions
	var (
		accu footprintAccumulator
		err  error
	)
	for _, dir := range dirs {
		err = scanDirectory(dir, seen, &accu)
		if err != nil {
			return nil, err
		}
	}
	return accu, nil
}
