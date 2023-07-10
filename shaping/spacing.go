package shaping

import (
	"golang.org/x/image/math/fixed"
)

// addWordSpacing alters the run, adding [additionalSpacing] on each
// word separator.
// [text] is the input slice used to create the run.
//
// See also https://www.w3.org/TR/css-text-3/#word-separator
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

// addLetterSpacing alters the run, adding [additionalSpacing] between
// each Harfbuzz clusters.
// Space it NOT included before the first cluster or after the last.
//
// See also https://www.w3.org/TR/css-text-3/#letter-spacing-property
func (run *Output) addLetterSpacing(additionalSpacing fixed.Int26_6) {
	isVertical := run.Direction.IsVertical()

	halfSpacing := additionalSpacing / 2
	for startGIdx := 0; startGIdx < len(run.Glyphs); {
		startGlyph := run.Glyphs[startGIdx]
		endGIdx := startGIdx + startGlyph.GlyphCount - 1

		if startGIdx > 0 {
			if isVertical {
				run.Glyphs[startGIdx].YAdvance += halfSpacing
				run.Glyphs[startGIdx].YOffset += halfSpacing
			} else {
				run.Glyphs[startGIdx].XAdvance += halfSpacing
				run.Glyphs[startGIdx].XOffset += halfSpacing
			}
		}

		if endGIdx < len(run.Glyphs)-1 {
			if isVertical {
				run.Glyphs[endGIdx].YAdvance += halfSpacing
				run.Glyphs[endGIdx].YOffset += halfSpacing
			} else {
				run.Glyphs[endGIdx].XAdvance += halfSpacing
				run.Glyphs[endGIdx].XOffset += halfSpacing
			}
		}

		// go to next cluster
		startGIdx += startGlyph.GlyphCount
	}

	run.RecomputeAdvance()
}
