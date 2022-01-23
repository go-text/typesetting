package fontscan

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/go-text/typesetting/font"
)

// this file implement a file path based font lookup
// it is designed to be fast, no requiring any database
// as a consequence it may be a bit less accurate than the FontSet API
// also, it cannot handle coverage based substitution

type scoredPaths struct {
	paths  []string
	scores []int
}

// Len is the number of elements in the collection.
func (sf scoredPaths) Len() int { return len(sf.paths) }

func (sf scoredPaths) Less(i int, j int) bool { return sf.scores[i] < sf.scores[j] }

// Swap swaps the elements with indexes i and j.
func (sf scoredPaths) Swap(i int, j int) {
	sf.paths[i], sf.paths[j] = sf.paths[j], sf.paths[i]
	sf.scores[i], sf.scores[j] = sf.scores[j], sf.scores[i]
}

// expand `family` with the standard substitutions,
// then loop through the given file names looking for the best match
func selectFileByFamily(inFamily string, paths []string) []string {
	inFamily = normalizeFamily(inFamily)
	crible := make(familyCrible)
	crible.fillWithSubstitutions(inFamily)

	var matches scoredPaths
	for _, filePath := range paths {
		filename, _ := splitAtDot(filePath)
		filename = normalizeFamily(filename)
		for family, score := range crible {
			// approximate search
			if strings.Contains(filename, family) {
				matches.paths = append(matches.paths, filePath)
				matches.scores = append(matches.scores, score)
			}
		}
	}

	sort.Stable(matches)

	return matches.paths
}

var ErrFontNotFound = errors.New("font not found")

// loop through `paths` and select the first face with
// a matching style.
// Is no exact match is found, the CSS rules for approximate match are applied
// the method panic if `paths` is empty
func selectByAspect(paths []string, aspect Aspect) (font.Face, Location, error) {
	// try for an exact match and build the fontset for approximate match
	var fs fontSet

	aspect.setDefaults()

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, Location{}, fmt.Errorf("opening font file %s: %s", path, err)
		}

		descriptors, format := getFontDescriptors(file)

		for index, descriptor := range descriptors {
			fontAspect := newAspectFromDescriptor(descriptor)

			loc := Location{
				File:  path,
				Index: uint16(index),
			}
			fs = append(fs, footprint{
				Aspect:   fontAspect,
				Location: loc,
			})

			if fontAspect == aspect { // exact match, return early
				faces, err := format.Loader()(file)
				if err != nil {
					// if an error occur (for instance for unsupported cmaps)
					// try the next file path
					break
				}

				file.Close()
				return faces[index], loc, nil
			}
		}

		file.Close()
	}

	if len(fs) == 0 { // unlikely, may happen if all the paths are invalid font files
		return nil, Location{}, ErrFontNotFound
	}

	// no exact match
	matches := fs.retainsBestMatches(allIndices(fs), aspect)

	footprint := fs[matches[0]]
	face, err := footprint.loadFromDisk()

	return face, footprint.Location, err
}

// FindFont looks for a font matching `family` and `aspect` in the
// standard font folders.
// If `family` is not found, suitable substitutions are tried
// to find a close font.
// If no exact match for `aspect` is found, the closer font is returned.
// If `aspect` is empty, it is replaced by a regular style.
//
// In the (unlikely) case where no font is found, ErrFontNotFound is returned.
func FindFont(family string, aspect Aspect) (font.Face, Location, error) {
	directories, err := DefaultFontDirectories()
	if err != nil {
		return nil, Location{}, err
	}

	paths, err := scanFontFiles(directories...)
	if err != nil {
		return nil, Location{}, err
	}

	paths = selectFileByFamily(family, paths)
	if len(paths) == 0 {
		return nil, Location{}, ErrFontNotFound
	}

	return selectByAspect(paths, aspect)
}

func allIndices(fs fontSet) []int {
	out := make([]int, len(fs))
	for i := range fs {
		out[i] = i
	}
	return out
}
