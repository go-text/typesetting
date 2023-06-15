//go:build !go1.16
// +build !go1.16

package fontscan

import (
	"fmt"
	"os"
	"path/filepath"
)

// recursively walk through the given directory, scanning font files and calling dst.consume
// for each valid file found.
func scanDirectory(dir string, visited map[string]bool, dst fontFileHandler) error {
	walkFn := func(path string, d os.FileInfo, err error) error {
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

	err := filepath.Walk(dir, walkFn)

	return err
}

// readDir re-implements os.ReadDir (Go 1.16+) using only Go 1.14's stdlib.
func readDir(name string) ([]os.DirEntry, error) {
	d, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	entries, err := d.ReadDir(0)
	if err != nil {
		return nil, err
	}
	return entries, nil
}
