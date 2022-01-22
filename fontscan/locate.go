package fontscan

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
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
	inFamily = ignoreBlanksAndCase(inFamily)
	crible := applySubstitutions(inFamily)

	var matches scoredPaths
	for _, filePath := range paths {
		filename, _ := splitAtDot(filePath)
		filename = ignoreBlanksAndCase(filename)
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
// a regular style, or `ErrFontNotFound` if not found
func selectRegular(paths []string) (font.Face, error) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening font file %s: %s", path, err)
		}

		descriptors, format := getFontDescriptors(file)

		for index, descriptor := range descriptors {
			aspect := newAspectFromDescriptor(descriptor)
			if aspect.Style == fonts.StyleNormal {
				// found it: load the the face
				faces, err := format.Loader()(file)
				if err != nil {
					// if an error occur (for instance for unsupported cmaps)
					// try the next file path
					break
				}

				file.Close()

				return faces[index], nil
			}
		}

		file.Close()
	}

	return nil, ErrFontNotFound
}

// FindFont look for a regular font matching `family` in the
// standard font folders.
// If `family` is not found, suitable substitutions are tried
// to find a close font.
// In the (unlikely) case where no font is found,
// ErrFontNotFound is returned.
func FindFont(family string) (font.Face, error) {
	directories, err := DefaultFontDirectories()
	if err != nil {
		return nil, err
	}

	paths, err := scanFontFiles(directories...)
	if err != nil {
		return nil, err
	}

	paths = selectFileByFamily(family, paths)

	face, err := selectRegular(paths)
	if err != nil {
		return nil, err
	}

	return face, nil
}
