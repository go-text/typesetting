package shaping

import (
	"github.com/go-text/typesetting/di"
	"golang.org/x/image/math/fixed"
)

// AddWordSpacing alters the run, adding [additionalSpacing] on each
// word separator.
// [text] is the input slice used to create the run.
// Note that space is always added, even on boundaries.
//
// See also the convenience function [AddSpacing] to handle a slice of runs.
//
// See also https://www.w3.org/TR/css-text-3/#word-separator
func (run *Output) AddWordSpacing(text []rune, additionalSpacing fixed.Int26_6) {
	isVertical := run.Direction.IsVertical()
	for i, g := range run.Glyphs {
		// find the corresponding runes :
		// to simplify, we assume a simple one to one rune/glyph mapping
		// which should be common in practice for word separators
		if !(g.RuneCount == 1 && g.GlyphCount == 1) {
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
		run.Glyphs[i].Advance += additionalSpacing
		if isVertical {
			run.Glyphs[i].YAdvance += additionalSpacing
			run.Glyphs[i].YOffset += additionalSpacing / 2
		} else {
			run.Glyphs[i].XAdvance += additionalSpacing
			run.Glyphs[i].XOffset += additionalSpacing / 2
		}
	}
	run.RecomputeAdvance()
}

// AddLetterSpacing alters the run, adding [additionalSpacing] between
// each Harfbuzz clusters.
//
// Space is added at the boundaries if and only if there is an adjacent run, as specified by [isStartRun] and [isEndRun].
//
// See also the convenience function [AddSpacing] to handle a slice of runs.
//
// See also https://www.w3.org/TR/css-text-3/#letter-spacing-property
func (run *Output) AddLetterSpacing(additionalSpacing fixed.Int26_6, isStartRun, isEndRun bool) {
	isVertical := run.Direction.IsVertical()
	// glyphs are in visual order, so with a TowardTopLeft progression the
	// visually leading glyph is the logical end of the run.
	reversed := run.Direction.Progression() == di.TowardTopLeft
	// run boundaries which must stay bare, seen from the visual side
	bareLeading, bareTrailing := isStartRun, isEndRun
	if reversed {
		bareLeading, bareTrailing = isEndRun, isStartRun
	}

	halfSpacing := additionalSpacing / 2
	for startGIdx := 0; startGIdx < len(run.Glyphs); {
		startGlyph := run.Glyphs[startGIdx]
		endGIdx := startGIdx + startGlyph.GlyphCount - 1
		isFirstCluster := startGIdx == 0
		isLastCluster := startGIdx+startGlyph.GlyphCount >= len(run.Glyphs)

		// visually leading side. Shift the glyph content and enlarge the advance.
		if !isFirstCluster || !bareLeading {
			g := &run.Glyphs[startGIdx]
			g.Advance += halfSpacing
			if isVertical {
				g.YAdvance += halfSpacing
				g.YOffset += halfSpacing
			} else {
				g.XAdvance += halfSpacing
				g.XOffset += halfSpacing
			}
			if reversed {
				g.endLetterSpacing += halfSpacing
			} else {
				g.startLetterSpacing += halfSpacing
			}
		}

		// visually trailing side. Only the advance grows.
		if !isLastCluster || !bareTrailing {
			g := &run.Glyphs[endGIdx]
			g.Advance += halfSpacing
			if isVertical {
				g.YAdvance += halfSpacing
			} else {
				g.XAdvance += halfSpacing
			}
			if reversed {
				g.startLetterSpacing += halfSpacing
			} else {
				g.endLetterSpacing += halfSpacing
			}
		}

		// go to next cluster
		startGIdx += startGlyph.GlyphCount
	}

	run.RecomputeAdvance()
}

// trimStartLetterSpacing removes the letter spacing added before the logical
// first glyph. It does not run RecomputeAdvance.
func (run *Output) trimStartLetterSpacing() {
	if len(run.Glyphs) == 0 {
		return
	}
	reversed := run.Direction.Progression() == di.TowardTopLeft
	firstG := &run.Glyphs[0]
	if reversed {
		firstG = &run.Glyphs[len(run.Glyphs)-1]
	}
	halfSpacing := firstG.startLetterSpacing
	firstG.Advance -= halfSpacing
	if run.Direction.IsVertical() {
		firstG.YAdvance -= halfSpacing
	} else {
		firstG.XAdvance -= halfSpacing
	}
	if !reversed { // the spacing was applied as an offset on the leading side
		if run.Direction.IsVertical() {
			firstG.YOffset -= halfSpacing
		} else {
			firstG.XOffset -= halfSpacing
		}
	}
	firstG.startLetterSpacing = 0
}

// AddSpacing adds additionnal spacing between words and letters, mutating the given [runs].
// [text] is the input slice the [runs] refer to.
//
// See the method [Output.AddWordSpacing] and [Output.AddLetterSpacing] for details
// about what spacing actually is.
func AddSpacing(runs []Output, text []rune, wordSpacing, letterSpacing fixed.Int26_6) {
	for i := range runs {
		isStartRun, isEndRun := i == 0, i == len(runs)-1
		if wordSpacing != 0 {
			runs[i].AddWordSpacing(text, wordSpacing)
		}
		if letterSpacing != 0 {
			runs[i].AddLetterSpacing(letterSpacing, isStartRun, isEndRun)
		}
	}
}
