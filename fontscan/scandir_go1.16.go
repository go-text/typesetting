//go:build go1.16
// +build go1.16

package fontscan

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// recursively walk through the given directory, scanning font files and calling dst.consume
// for each valid file found.
func scanDirectory(dir string, visited map[string]bool, dst fontFileHandler) error {
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("error walking font directory %q: %v", path, err)
			return filepath.SkipDir
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

type DirEntry = fs.DirEntry

func readDir(name string) ([]DirEntry, error) {
	return os.ReadDir(name)
}