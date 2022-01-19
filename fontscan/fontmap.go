package fontscan

import (
	"sort"
)

// the match algorithm is inspired from fontconfig and Rust font-kit library

// FontSet stores the fonts available for text shaping.
// It supports both system and user added fonts.
type FontSet struct {
	systemFootprint []Footprint
	userFootprint   []Footprint
}

// stores the possible matches with their score:
// lower is better
type familyCrible map[string]int

func newFamilyCrible(family string) familyCrible {
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

// loop through `footprints` and stores the matching fonts into `dst`
func (fc familyCrible) selectFrom(footprints []Footprint, dst *scoredFootprints) {
	for _, footprint := range footprints {
		if score := fc.matches(footprint.Family); score != -1 {
			dst.footprints = append(dst.footprints, footprint)
			dst.scores = append(dst.scores, score)
		}
	}
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
// the returned slice may be empty if no font matches the given `family`
func (fm *FontSet) selectByFamily(family string) []Footprint {
	// build the crible, handling recursive substitutions
	crible := newFamilyCrible(family)

	var matches scoredFootprints

	// select the matching fonts
	crible.selectFrom(fm.systemFootprint, &matches)
	crible.selectFrom(fm.userFootprint, &matches)

	// sort the matched font by score (lower is better)
	sort.Stable(matches)

	return matches.footprints
}
