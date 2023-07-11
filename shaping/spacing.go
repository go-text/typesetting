package shaping

import (
	"github.com/go-text/typesetting/di"
	"golang.org/x/image/math/fixed"
)

// AddWordSpacing alters the run, adding [additionalSpacing] on each
// word separator.
// [text] is the input slice used to create the run.
//
// See also https://www.w3.org/TR/css-text-3/#word-separator
func (run *Output) AddWordSpacing(text []rune, additionalSpacing fixed.Int26_6) {
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

// AddLetterSpacing alters the run, adding [additionalSpacing] between
// each Harfbuzz clusters.
//
// Space it also included before the first cluster and after the last cluster.
//
// See [Line.TrimLetterSpacing] to trim unwanted space at line boundaries.
//
// See also https://www.w3.org/TR/css-text-3/#letter-spacing-property
func (run *Output) AddLetterSpacing(additionalSpacing fixed.Int26_6) {
	isVertical := run.Direction.IsVertical()

	halfSpacing := additionalSpacing / 2
	for startGIdx := 0; startGIdx < len(run.Glyphs); {
		startGlyph := run.Glyphs[startGIdx]
		endGIdx := startGIdx + startGlyph.GlyphCount - 1

		if isVertical {
			run.Glyphs[startGIdx].YAdvance += halfSpacing
			run.Glyphs[startGIdx].YOffset += halfSpacing
		} else {
			run.Glyphs[startGIdx].XAdvance += halfSpacing
			run.Glyphs[startGIdx].XOffset += halfSpacing
		}

		if isVertical {
			run.Glyphs[endGIdx].YAdvance += halfSpacing
		} else {
			run.Glyphs[endGIdx].XAdvance += halfSpacing
		}

		// go to next cluster
		startGIdx += startGlyph.GlyphCount
	}

	run.RecomputeAdvance()
}

// TrimLetterSpacing post-processes the line by removing the given [letterSpacing]
// at the start and at the end of the line.
//
// This method should be used after wrapping runs altered by [Output.AddLetterSpacing],
// with the same [letterSpacing] argument.
func (line Line) TrimLetterSpacing(letterSpacing fixed.Int26_6) {
	if len(line) == 0 {
		return
	}

	halfSpacing := letterSpacing / 2
	firstRun, lastRun := &line[0], &line[len(line)-1]
	if firstRun.Direction.Axis() == di.Horizontal {
		firstRun.Glyphs[0].XOffset -= halfSpacing
		firstRun.Glyphs[0].XAdvance -= halfSpacing

		L := len(lastRun.Glyphs)
		firstRun.Glyphs[L-1].XAdvance -= halfSpacing
	} else {
		firstRun.Glyphs[0].YOffset -= halfSpacing
		firstRun.Glyphs[0].YAdvance -= halfSpacing

		L := len(lastRun.Glyphs)
		firstRun.Glyphs[L-1].YAdvance -= halfSpacing
	}

	firstRun.RecomputeAdvance()
	lastRun.RecomputeAdvance()
}
