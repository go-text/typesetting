package fontscan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-text/typesetting/font"
)

// The family substitution algorithm is copied from fontconfig
// and the match algorithm is inspired from Rust font-kit library

// FontSet stores the list of fonts available for text shaping.
// It is usually build from a system font index or by manually appending
// fonts.
type FontSet []Footprint

// stores the possible matches with their score:
// lower is better
type familyCrible map[string]int

// applySubstitutions starts from `family` (ignoring blank and case)
// and applies all the substitutions coded in the package
// to add substitutes values
func applySubstitutions(family string) familyCrible {
	family = ignoreBlanksAndCase(family)

	fl := newFamilyList([]string{family})
	for _, subs := range familySubstitution {
		fl.execute(subs)
	}

	return fl.compile()
}

// returns -1 if no match
func (fc familyCrible) matches(family string) int {
	if score, has := fc[ignoreBlanksAndCase(family)]; has {
		return score
	}
	return -1
}

type scoredFootprints struct {
	footprints []Footprint
	scores     []int
}

// Len is the number of elements in the collection.
func (sf scoredFootprints) Len() int { return len(sf.footprints) }

func (sf scoredFootprints) Less(i int, j int) bool { return sf.scores[i] < sf.scores[j] }

// Swap swaps the elements with indexes i and j.
func (sf scoredFootprints) Swap(i int, j int) {
	sf.footprints[i], sf.footprints[j] = sf.footprints[j], sf.footprints[i]
	sf.scores[i], sf.scores[j] = sf.scores[j], sf.scores[i]
}

// selectByFamily returns all the fonts in the fontmap matching
// the given `family`, with the best matches coming first.
// `family` is expanded to known similar families, and may also
// be one the following generic family : "serif", "sans-serif", "monospace", "cursive", "fantasy"
// The returned slice may be empty if no font matches the given `family`
func (fm FontSet) selectByFamily(family string) []Footprint {
	// build the crible, handling substitutions
	crible := applySubstitutions(family)

	var matches scoredFootprints

	// select the matching fonts:
	// loop through `footprints` and stores the matching fonts into `dst`
	for _, footprint := range fm {
		if score := crible.matches(footprint.Family); score != -1 {
			matches.footprints = append(matches.footprints, footprint)
			matches.scores = append(matches.scores, score)
		}
	}

	// sort the matched font by score (lower is better)
	sort.Stable(matches)

	return matches.footprints
}

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
// finaly, load the first valid font file.
func selectFileByFamily(inFamily string, paths []string) []font.Face {
	crible := applySubstitutions(inFamily)

	var matches scoredPaths
	for _, filePath := range paths {
		filename, _ := splitAtDot(filePath)
		filename = ignoreBlanksAndCase(filename)
		for family, score := range crible {
			if strings.Contains(filename, family) {
				matches.paths = append(matches.paths, filePath)
				matches.scores = append(matches.scores, score)
			}
		}
	}

	sort.Stable(matches)
	fmt.Println(matches.paths)
	return nil
}
