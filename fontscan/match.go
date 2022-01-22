package fontscan

import (
	"sort"

	"github.com/benoitkugler/textlayout/fonts"
)

// Query exposes the intention of an author about the
// font to use to shape and render text.
type Query struct {
	// Families is a list of required families,
	// the first having the highest priority.
	// Each of them is tried until a suitable match is found.
	Families []string

	// Aspect selects which particular face to use among
	// the font matching the family criteria.
	Aspect Aspect
}

// fontSet stores the list of fonts available for text shaping.
// It is usually build from a system font index or by manually appending
// fonts.
type fontSet []footprint

// stores the possible matches with their score:
// lower is better
type familyCrible map[string]int

func newFamilyCrible(family string, substitute bool) familyCrible {
	family = ignoreBlanksAndCase(family)

	// always substitute generic families
	if substitute || isGenericFamily(family) {
		return applySubstitutions(family)
	}

	return familyCrible{family: 0}
}

// applySubstitutions starts from `family` (ignoring blank and case)
// and applies all the substitutions coded in the package
// to add substitutes values
func applySubstitutions(family string) familyCrible {
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
	footprints []int
	scores     []int
}

// keep the underlying storage
func (sf *scoredFootprints) reset() {
	sf.footprints = sf.footprints[:0]
	sf.scores = sf.scores[:0]
}

// Len is the number of elements in the collection.
func (sf scoredFootprints) Len() int { return len(sf.footprints) }

func (sf scoredFootprints) Less(i int, j int) bool { return sf.scores[i] < sf.scores[j] }

// Swap swaps the elements with indexes i and j.
func (sf scoredFootprints) Swap(i int, j int) {
	sf.footprints[i], sf.footprints[j] = sf.footprints[j], sf.footprints[i]
	sf.scores[i], sf.scores[j] = sf.scores[j], sf.scores[i]
}

func isGenericFamily(family string) bool {
	switch family {
	case "serif", "sans-serif", "monospace", "cursive", "fantasy":
		return true
	default:
		return false
	}
}

// selectByFamily returns all the fonts in the fontmap matching
// the given `family`, with the best matches coming first.
// `substitute` controls whether or not system substitutions are applied.
// The following generic family : "serif", "sans-serif", "monospace", "cursive", "fantasy"
// are always expanded to concrete families.
// The returned slice may be empty if no font matches the given `family`.
// buffer is used to reduce allocations
func (fm fontSet) selectByFamily(family string, substitute bool, buffer *scoredFootprints) []int {
	// build the crible, handling substitutions
	crible := newFamilyCrible(family, substitute)

	buffer.reset()

	// select the matching fonts:
	// loop through `footprints` and stores the matching fonts into `dst`
	for index, footprint := range fm {
		if score := crible.matches(footprint.Family); score != -1 {
			buffer.footprints = append(buffer.footprints, index)
			buffer.scores = append(buffer.scores, score)
		}
	}

	// sort the matched font by score (lower is better)
	sort.Stable(*buffer)

	return buffer.footprints
}

// matchStretch look for the given stretch in the font set,
// or, if not found, the closest stretch
// if always return a valid value (contained in `candidates`) if `candidates` is not empty
func (fs fontSet) matchStretch(candidates []int, query Stretch) Stretch {
	// narrower and wider than the query
	var narrower, wider Stretch

	for _, index := range candidates {
		stretch := fs[index].Aspect.Stretch
		if stretch > query { // wider candidate
			if wider == 0 || stretch-query < wider-query { // closer
				wider = stretch
			}
		} else if stretch < query { // narrower candidate
			// if narrower == 0, it is always more distant to queryStretch than stretch
			if query-stretch < query-narrower { // closer
				narrower = stretch
			}
		} else {
			// found an exact match, just return it
			return query
		}
	}

	// default to closest
	if query <= fonts.StretchNormal { // narrow first
		if narrower != 0 {
			return narrower
		}
		return wider
	} else { // wide first
		if wider != 0 {
			return wider
		}
		return narrower
	}
}

// matchStyle look for the given style in the font set,
// or, if not found, the closest style
// if always return a valid value (contained in `fs`) if `fs` is not empty
func (fs fontSet) matchStyle(candidates []int, query Style) Style {
	var crible [fonts.StyleOblique + 1]bool

	for _, index := range candidates {
		crible[fs[index].Aspect.Style] = true
	}

	switch query {
	case fonts.StyleNormal: // StyleNormal, StyleOblique, StyleItalic
		if crible[fonts.StyleNormal] {
			return fonts.StyleNormal
		} else if crible[fonts.StyleOblique] {
			return fonts.StyleOblique
		} else {
			return fonts.StyleItalic
		}
	case fonts.StyleItalic: // StyleItalic, StyleOblique, StyleNormal
		if crible[fonts.StyleItalic] {
			return fonts.StyleItalic
		} else if crible[fonts.StyleOblique] {
			return fonts.StyleOblique
		} else {
			return fonts.StyleNormal
		}
	case fonts.StyleOblique: // StyleOblique, StyleItalic, StyleNormal
		if crible[fonts.StyleOblique] {
			return fonts.StyleOblique
		} else if crible[fonts.StyleItalic] {
			return fonts.StyleItalic
		} else {
			return fonts.StyleNormal
		}
	}

	panic("should not happen") // query.Style is sanitized by setDefaults
}

// matchWeight look for the given weight in the font set,
// or, if not found, the closest weight
// if always return a valid value (contained in `fs`) if `fs` is not empty
// we follow https://drafts.csswg.org/css-fonts/#font-style-matching
func (fs fontSet) matchWeight(candidates []int, query Weight) Weight {
	var fatter, thinner Weight // approximate match
	for _, index := range candidates {
		weight := fs[index].Aspect.Weight
		if weight > query { // fatter candidate
			if fatter == 0 || weight-query < fatter-query { // weight is closer to query
				fatter = weight
			}
		} else if weight < query {
			if query-weight < query-thinner { // weight is closer to query
				thinner = weight
			}
		} else {
			// found an exact match, just return it
			return query
		}
	}

	// approximate match
	if 400 <= query && query <= 500 { // fatter until 500, then thinner then fatter
		if fatter != 0 && fatter <= 500 {
			return fatter
		} else if thinner != 0 {
			return thinner
		}
		return fatter
	} else if query < 400 { // thinner then fatter
		if thinner != 0 {
			return thinner
		}
		return fatter
	} else { // fatter then thinner
		if fatter != 0 {
			return fatter
		}
		return thinner
	}
}

// filter `candidates` in place and returns the updated slice
func (fs fontSet) filterByStretch(candidates []int, stretch Stretch) []int {
	n := 0
	for _, index := range candidates {
		if fs[index].Aspect.Stretch == stretch {
			candidates[n] = index
			n++
		}
	}
	candidates = candidates[:n]
	return candidates
}

// filter `candidates` in place and returns the updated slice
func (fs fontSet) filterByStyle(candidates []int, style Style) []int {
	n := 0
	for _, index := range candidates {
		if fs[index].Aspect.Style == style {
			candidates[n] = index
			n++
		}
	}
	candidates = candidates[:n]
	return candidates
}

// filter `candidates` in place and returns the updated slice
func (fs fontSet) filterByWeight(candidates []int, weight Weight) []int {
	n := 0
	for _, index := range candidates {
		if fs[index].Aspect.Weight == weight {
			candidates[n] = index
			n++
		}
	}
	candidates = candidates[:n]
	return candidates
}

// retainsBestMatches narrows `candidates` to the closest footprints to `query`, according to the CSS font rules
// note that this method mutate `candidates`
func (fs fontSet) retainsBestMatches(candidates []int, query Aspect) []int {
	// this follows CSS Fonts Level 3 § 5.2 [1].
	// https://drafts.csswg.org/css-fonts-3/#font-style-matching

	query.setDefaults()

	// First step: font-stretch
	matchingStretch := fs.matchStretch(candidates, query.Stretch)
	candidates = fs.filterByStretch(candidates, matchingStretch) // only retain matching stretch

	// Second step : font-style
	matchingStyle := fs.matchStyle(candidates, query.Style)
	candidates = fs.filterByStyle(candidates, matchingStyle)

	// Third step : font-weight
	matchingWeight := fs.matchWeight(candidates, query.Weight)
	candidates = fs.filterByWeight(candidates, matchingWeight)

	return candidates
}
