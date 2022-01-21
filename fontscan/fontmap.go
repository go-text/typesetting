package fontscan

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/go-text/typesetting/font"
)

// The family substitution algorithm is copied from fontconfig
// and the match algorithm is inspired from Rust font-kit library

// FontMap provides a mechanism to select a font.Face from a font description.
// It supports system and user-provided fonts, and implements the CSS font substitutions
// rules.
// It is designed to work with an index built by scanning the system fonts,
// which is a costly operation (see XXX for more details).
// A lightweight alternative is provided by the FindFont function, which only use
// file paths to select a font.
type FontMap struct {
	// the database to query, either loaded from an index
	// or populated with the XXX method
	database FontSet

	// TODO: // an internal buffer used when matching fonts
	// candidates FontSet

	// the current query, which influences ResolveFace output
	query FontQuery
}

// SetQuery set the families and aspect required, influencing subsequent
// `ResolveFace` calls.
func (fm *FontMap) SetQuery(query FontQuery) {
	// TODO: caching layer, since many runes will be looked for
	// the same query
	fm.query = query
}

// ResolveFace select a face based on the current query (see SetQuery),
// applying CSS font selection rules.
// The function will return nil if the underlying font database is empty,
// or if the file system is broken; otherwise the returned font.Face is always valid.
func (fm *FontMap) ResolveFace(r rune) font.Face {
	// TODO: caching layer
	selectFace := func(substitute bool) font.Face {
		for _, family := range fm.query.Families {
			candidates := fm.database.selectByFamily(family, substitute)
			if len(candidates) == 0 {
				continue
			}

			// select the correct aspect
			fp := candidates.selectBestMatch(fm.query.Aspect)

			// check the coverage
			if fp.Runes.Contains(r) {
				// try to use the font
				face, err := fm.loadFace(fp)
				if err != nil { // very unlikely; try an other family
					log.Println(err)
					continue
				}
				log.Println("found", fp.Location.File)
				return face
			}
		}
		return nil
	}

	// we first look up for an exact family match, without substitutions
	face := selectFace(false)
	if face != nil {
		return face
	}

	// if no family has matched so far, try again with system fallback
	face = selectFace(true)
	if face != nil {
		return face
	}

	// FIXME:
	fmt.Println("BAD")

	// this is very very unlikely, since the substitution
	// always add a default generic family
	for _, fp := range fm.database {
		face, err := fm.loadFace(fp)
		if err != nil { // very unlikely
			continue
		}
		return face
	}

	// arg, we have a very very serious issue here
	return nil
}

func (fm *FontMap) loadFace(fp Footprint) (font.Face, error) {
	// TODO: handle user font and caching
	file, err := os.Open(fp.Location.File)
	if err != nil {
		return nil, err
	}

	faces, err := fp.Format.Loader()(file)
	if err != nil {
		return nil, err
	}

	return faces[fp.Location.Index], nil
}

// func (fm *FontMap) resetBuffer() {
// 	if cap(fm.candidates) < len(fm.database) { // grow the buffer
// 		fm.candidates = make(FontSet, 0, len(fm.database))
// 	}
// 	fm.candidates = fm.candidates[0:len(fm.database)]
// 	copy(fm.candidates, fm.database)
// }

// FontQuery exposes the intention of an author about the
// font to use to shape and render text.
type FontQuery struct {
	// Families is a list of required families,
	// the first having the highest priority.
	// Each of them is tried until a suitable match is found.
	Families []string

	// Aspect selects which particular face to use among
	// the font matching the family criteria.
	Aspect Aspect
}

// FontSet stores the list of fonts available for text shaping.
// It is usually build from a system font index or by manually appending
// fonts.
type FontSet []Footprint

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
func (fm FontSet) selectByFamily(family string, substitute bool) FontSet {
	// build the crible, handling substitutions
	crible := newFamilyCrible(family, substitute)

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

// matchStretch look for the given stretch in the font set,
// or, if not found, the closest stretch
// if always return a valid value (contained in `fs`) if `fs` is not empty
func (fs FontSet) matchStretch(query Stretch) Stretch {
	// narrower and wider than the query
	var narrower, wider Stretch

	for _, fp := range fs {
		stretch := fp.Aspect.Stretch
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
func (fs FontSet) matchStyle(query Style) Style {
	var crible [fonts.StyleOblique + 1]bool

	for _, fp := range fs {
		crible[fp.Aspect.Style] = true
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
func (fs FontSet) matchWeight(query Weight) Weight {
	var fatter, thinner Weight // approximate match
	for _, fp := range fs {
		weight := fp.Aspect.Weight
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

// filter in place
func (fs *FontSet) filterByStretch(stretch Stretch) {
	n := 0
	for _, fp := range *fs {
		if fp.Aspect.Stretch == stretch {
			(*fs)[n] = fp
			n++
		}
	}
	*fs = (*fs)[:n]
}

// filter in place
func (fs *FontSet) filterByStyle(style Style) {
	n := 0
	for _, fp := range *fs {
		if fp.Aspect.Style == style {
			(*fs)[n] = fp
			n++
		}
	}
	*fs = (*fs)[:n]
}

// filter in place
func (fs *FontSet) filterByWeight(weight Weight) {
	n := 0
	for _, fp := range *fs {
		if fp.Aspect.Weight == weight {
			(*fs)[n] = fp
			n++
		}
	}
	*fs = (*fs)[:n]
}

// selectBestMatch returns the closest footprint to `query`, according to the CSS font rules
// note that this method mutate `fs`
// the function will panic if `fs` is empty
func (fs FontSet) selectBestMatch(query Aspect) Footprint {
	// this follows CSS Fonts Level 3 § 5.2 [1].
	// https://drafts.csswg.org/css-fonts-3/#font-style-matching

	query.setDefaults()

	// First step: font-stretch
	matchingStretch := fs.matchStretch(query.Stretch)
	fs.filterByStretch(matchingStretch) // only retain matching stretch

	// Second step : font-style
	matchingStyle := fs.matchStyle(query.Style)
	fs.filterByStyle(matchingStyle)

	// Third step : font-weight
	matchingWeight := fs.matchWeight(query.Weight)
	fs.filterByWeight(matchingWeight)

	return fs[0]
}
