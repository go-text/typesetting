package shaping

import (
	"golang.org/x/image/math/fixed"
)

// addWordSpacing alter the run, adding [additionalSpacing] on each
// word separator.
// [text] is the origin input slice used to create the run.
// See https://www.w3.org/TR/css-text-3/#word-separator
func (run *Output) addWordSpacing(text []rune, additionalSpacing fixed.Int26_6) {
	isVertical := run.Direction.IsVertical()
	for i, g := range run.Glyphs {
		// find the corresponding runes :
		// to simplify, we assume the words separators are not produced by ligatures
		// so that the cluster "rune-length" is 1
		if g.RuneCount != 1 {
			continue
		}
		r := text[g.ClusterIndex]
		switch r {
		case '\u0020', // space
			'\u00A0',                   // no-break space
			'\u1361',                   // Ethiopic word space
			'\U00010100', '\U00010101', // Aegean word separators
			'\U0001039F', // Ugaritic word divider
			'\U0001091F': // Phoenician word separator
		default:
			continue
		}
		// we have a word separator: add space
		// we do it by enlarging the separator glyph advance
		// and distributing space around the glyph content
		if isVertical {
			run.Glyphs[i].YAdvance += additionalSpacing
			run.Glyphs[i].YOffset += additionalSpacing / 2
		} else {
			run.Glyphs[i].XAdvance += additionalSpacing
			run.Glyphs[i].XOffset += additionalSpacing / 2 // distribute space around the char
		}
	}
	run.RecomputeAdvance()
}
