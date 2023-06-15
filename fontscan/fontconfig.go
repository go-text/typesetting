// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package fontscan

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Support for a (very limited) subset of the Linux fontconfig config file format
// See https://www.freedesktop.org/wiki/Software/fontconfig/ for reference

const fcRootConfig = "/etc/fonts/fonts.conf"

const (
	_ = iota
	fcDir
	fcInclude
)

// fcDirective is either a <dir> or a <include> element,
// as indicated by [kind]
type fcDirective struct {
	dir struct {
		Dir    string `xml:",chardata"`
		Prefix string `xml:"prefix,attr"`
	}
	include struct {
		Include       string `xml:",chardata"`
		IgnoreMissing string `xml:"ignore_missing,attr"`
		Prefix        string `xml:"prefix,attr"`
	}
	kind uint8
}

func (directive *fcDirective) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case "dir":
		directive.kind = fcDir
		return d.DecodeElement(&directive.dir, &start)
	case "include":
		directive.kind = fcInclude
		return d.DecodeElement(&directive.include, &start)
	default:
		// ignore the element
		return d.Skip()
	}
}

// parseFcFile opens and process a FontConfig config file,
// returning the font directories to scan and the (optionnal)
// supplementary config files (or directories) to include.
func parseFcFile(file, currentWorkingDir string) (fontDirs, includes []string, _ error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, fmt.Errorf("opening fontconfig config file: %s", err)
	}
	defer f.Close()

	var config struct {
		Fontconfig []fcDirective `xml:",any"`
	}
	err = xml.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing fontconfig config file: %s", err)
	}

	// post-process : handle "prefix" attr and use absolute path
	xdg := os.Getenv("XDG_DATA_HOME")
	for _, item := range config.Fontconfig {
		switch item.kind {
		case fcDir:
			dir := item.dir.Dir
			switch item.dir.Prefix {
			case "default", "cwd":
				dir = filepath.Join(currentWorkingDir, dir)
			case "relative":
				dir = filepath.Join(filepath.Dir(file), dir)
			case "xdg":
				dir = filepath.Join(xdg, dir)
			}
			fontDirs = append(fontDirs, dir)
		case fcInclude:
			include := item.include.Include
			if item.include.Prefix == "xdg" {
				include = filepath.Join(xdg, include)
			} else {
				// handle implicit relative dirs
				if strings.HasPrefix(include, "~") {
					include = expandUser(include)
				} else if !filepath.IsAbs(include) {
					include = filepath.Join(filepath.Dir(file), include)
				}
			}
			includes = append(includes, include)
		}
	}
	return
}

// parseFcDir processes all the files in [dir] matching the [09]*.conf pattern
// seen is updated with the processed fontconfig files
func parseFcDir(dir, currentWorkingDir string, seen map[string]bool) (fontDirs, includes []string, _ error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading fontconfig config directory: %s", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if name := entry.Name(); strings.HasSuffix(name, ".conf") {
			c := name[0]
			if '0' <= c && c <= '9' {
				file := filepath.Join(dir, name)
				seen[file] = true
				fds, incs, err := parseFcFile(file, currentWorkingDir)
				if err != nil {
					return nil, nil, err
				}
				fontDirs = append(fontDirs, fds...)
				includes = append(includes, incs...)
			}
		}
	}

	return
}

// parseFcConfig recursively parses the fontconfig config file at [rootConfig]
// and its includes, returning the font directories to scan
func parseFcConfig(rootConfig string) ([]string, error) {
	seen := map[string]bool{rootConfig: true}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("processing fontconfig config file: %s", err)
	}

	// includes is a queue
	dirs, includes, err := parseFcFile(rootConfig, cwd)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(includes); i++ {
		include := includes[i]
		if seen[include] {
			continue
		}
		seen[include] = true

		fi, err := os.Stat(include)
		if err != nil { // gracefully ignore broken includes
			log.Printf("missing fontconfig include %s: skipping", include)
			continue
		}

		var newDirs, newIncludes []string
		if fi.IsDir() {
			newDirs, newIncludes, err = parseFcDir(include, cwd, seen)
		} else {
			newDirs, newIncludes, err = parseFcFile(include, cwd)
		}
		if err != nil {
			return nil, err
		}

		dirs = append(dirs, newDirs...)
		includes = append(includes, newIncludes...)
	}

	return dirs, nil
}
