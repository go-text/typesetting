package shaping

import (
	"github.com/gioui/uax/segment"
	"github.com/gioui/uax/uax14"
)

// glyphIndex is the index in a Glyph slice
type glyphIndex = int

// length is the unit used to measure a width
type length = int

// mapRunesToClusterIndices returns a slice. Each index within that slice corresponds
// to an index within the runes input slice. The value stored at that index is the
// index of the glyph at the start of the corresponding glyph cluster shaped by
// harfbuzz.
func mapRunesToClusterIndices(runes []rune, glyphs []Glyph) []glyphIndex {
	mapping := make([]glyphIndex, len(runes))
	glyphCursor := 0
	if len(runes) == 0 {
		return nil
	}
	// If the final cluster values are lower than the starting ones,
	// the text is RTL.
	rtl := len(glyphs) > 0 && glyphs[len(glyphs)-1].ClusterIndex < glyphs[0].ClusterIndex
	if rtl {
		glyphCursor = len(glyphs) - 1
	}
	for i := range runes {
		for glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			((rtl && glyphs[glyphCursor].ClusterIndex <= i) ||
				(!rtl && glyphs[glyphCursor].ClusterIndex < i)) {
			if rtl {
				glyphCursor--
			} else {
				glyphCursor++
			}
		}
		if rtl {
			glyphCursor++
		} else if (glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			glyphs[glyphCursor].ClusterIndex > i) ||
			(glyphCursor == len(glyphs) && len(glyphs) > 1) {
			glyphCursor--
			targetClusterIndex := glyphs[glyphCursor].ClusterIndex
			for glyphCursor-1 >= 0 && glyphs[glyphCursor-1].ClusterIndex == targetClusterIndex {
				glyphCursor--
			}
		}
		if glyphCursor < 0 {
			glyphCursor = 0
		} else if glyphCursor >= len(glyphs) {
			glyphCursor = len(glyphs) - 1
		}
		mapping[i] = glyphCursor
	}
	return mapping
}

// inclusiveGlyphRange returns the inclusive range of runes and glyphs matching
// the provided start and breakAfter rune positions.
// runeToGlyph must be a valid mapping from the rune representation to the
// glyph reprsentation produced by mapRunesToClusterIndices.
// numGlyphs is the number of glyphs in the output representing the runes
// under consideration.
func inclusiveGlyphRange(start, breakAfter int, runeToGlyph []int, numGlyphs int) (glyphStart, glyphEnd glyphIndex) {
	rtl := runeToGlyph[len(runeToGlyph)-1] < runeToGlyph[0]
	if rtl {
		glyphStart = runeToGlyph[breakAfter]
		if start-1 >= 0 {
			glyphEnd = runeToGlyph[start-1] - 1
		} else {
			glyphEnd = numGlyphs - 1
		}
	} else {
		glyphStart = runeToGlyph[start]
		if breakAfter+1 < len(runeToGlyph) {
			glyphEnd = runeToGlyph[breakAfter+1] - 1
		} else {
			glyphEnd = numGlyphs - 1
		}
	}
	return
}

// breakOption represets a location within the rune slice at which
// it may be safe to break a line of text.
type breakOption struct {
	// breakAtRune is the index at which it is safe to break.
	breakAtRune int
	// penalty is the cost of breaking at this index. Negative
	// penalties mean that the break is beneficial, and a penalty
	// of uax14.PenaltyForMustBreak means a required break.
	penalty int
}

// getBreakOptions returns a slice of line break candidates for the
// text in the provided slice.
func getBreakOptions(text []rune) []breakOption {
	// Collect options for breaking the lines in a slice.
	var options []breakOption
	const adjust = -1
	breaker := uax14.NewLineWrap()
	segmenter := segment.NewSegmenter(breaker)
	segmenter.InitFromSlice(text)
	runeOffset := 0
	brokeAtEnd := false
	for segmenter.Next() {
		penalty, _ := segmenter.Penalties()
		// Determine the indices of the breaking runes in the runes
		// slice. Would be nice if the API provided this.
		currentSegment := segmenter.Runes()
		runeOffset += len(currentSegment)

		// Collect all break options.
		options = append(options, breakOption{
			penalty:     penalty,
			breakAtRune: runeOffset + adjust,
		})
		if options[len(options)-1].breakAtRune == len(text)-1 {
			brokeAtEnd = true
		}
	}
	if len(text) > 0 && !brokeAtEnd {
		options = append(options, breakOption{
			penalty:     uax14.PenaltyForMustBreak,
			breakAtRune: len(text) - 1,
		})
	}
	return options
}

// Range indicates the location of a sequence of elements within a longer slice.
type Range struct {
	Offset int
	Count  int
}

// shapedPararaph holds a one-dimentional, shaped text,
// to be wrapped into lines
type shapedPararaph struct {
	// out - the Output that is being line-broken.
	// text is the original input text
	text []rune
	// glyphs is the result of the Harfbuzz shaping
	glyphs Output
	// breaks is the slice of the break options collected for the text
	breaks []breakOption
	// mapping is a mapping where accessing the slice at the index of a rune
	// will yield the index of the first glyph corresponding to that rune.
	mapping []glyphIndex
}

func newShapedParagraph(text []rune, glyphs Output) shapedPararaph {
	return shapedPararaph{
		text:    text,
		glyphs:  glyphs,
		mapping: mapRunesToClusterIndices(text, glyphs.Glyphs),
		breaks:  getBreakOptions(text),
	}
}

// shouldKeepSegmentOnLine decides whether the segment of text from the current
// end of the line to the provided breakOption should be kept on the current
// line. It should be called successively with each available breakOption,
// and the line should be broken (without keeping the current segment)
// whenever it returns false.
//
// The parameters require some explanation:
// lineStartRune - the index of the first rune in the line.
// b - the line break candidate under consideration.
// curLineWidth - the amount of space total in the current line.
// curLineUsed - the amount of space in the current line that is already used.
// nextLineWidth - the amount of space available on the next line.
//
// This function returns both a valid Output broken at b and a boolean
// indicating whether the returned output should be used.
func (sp shapedPararaph) shouldKeepSegmentOnLine(lineStartRune int, b breakOption, curLineWidth, curLineUsed, nextLineWidth length) (candidateLine Output, keep bool) {
	// Convert the break target to an inclusive index.
	glyphStart, glyphEnd := inclusiveGlyphRange(lineStartRune, b.breakAtRune, sp.mapping, len(sp.glyphs.Glyphs))

	// Construct a line out of the inclusive glyph range.
	candidateLine = sp.glyphs
	candidateLine.Glyphs = candidateLine.Glyphs[glyphStart : glyphEnd+1]
	candidateLine.RecomputeAdvance()
	candidateAdvance := candidateLine.Advance.Ceil()
	if candidateAdvance > curLineWidth && candidateAdvance-curLineUsed <= nextLineWidth {
		// If it fits on the next line, put it there.
		return candidateLine, false
	}

	return candidateLine, true
}

// sanitizeBreaks remove break options not compatible
// with harbuzz shaping
func (sp *shapedPararaph) sanitizeBreaks() {
	for i := 0; i < len(sp.breaks); i++ {
		b := sp.breaks[i]
		if b.breakAtRune+1 < len(sp.mapping) {
			// Check if this break is valid.
			gIdx := sp.mapping[b.breakAtRune]
			g2Idx := sp.mapping[b.breakAtRune+1]
			cIdx := sp.glyphs.Glyphs[gIdx].ClusterIndex
			c2Idx := sp.glyphs.Glyphs[g2Idx].ClusterIndex
			if cIdx == c2Idx {
				// This break is within a harfbuzz cluster, and is
				// therefore invalid.
				copy(sp.breaks[i:], sp.breaks[i+1:])
				sp.breaks = sp.breaks[:len(sp.breaks)-1]
				i--
			}
		}
	}
}

// lineWrap wraps the shaped glyphs of a paragraph to a particular max width.
func (sp shapedPararaph) lineWrap(maxWidth int) []output {
	if len(sp.breaks) == 0 {
		// Pass empty lines through as empty.
		return []output{{
			Shaped: sp.glyphs,
			RuneRange: Range{
				Count: len(sp.text),
			},
		}}
	}

	sp.sanitizeBreaks()

	var outputs []output
	start := 0
	runesProcessedCount := 0
	for i := 0; i < len(sp.breaks); i++ {
		b := sp.breaks[i]
		// Always keep the first segment on a line.
		good, _ := sp.shouldKeepSegmentOnLine(start, b, maxWidth, 0, maxWidth)
		end := b.breakAtRune

		// Search through break candidates looking for candidates that can fit on the current line.
		for k := i + 1; k < len(sp.breaks); k++ {
			bb := sp.breaks[k]
			candidate, ok := sp.shouldKeepSegmentOnLine(start, bb, maxWidth, good.Advance.Ceil(), maxWidth)
			if ok {
				// The break described by bb fits on this line. Use this new, longer segment.
				good = candidate
				end = bb.breakAtRune
				i++
			} else {
				// The break described by bb will not fit on this line, commit whatever the last good
				// break was and then start a new line.
				break
			}
		}

		lineRuneCount := end - start + 1
		outputs = append(outputs, output{
			Shaped: good,
			RuneRange: Range{
				Count:  lineRuneCount,
				Offset: runesProcessedCount,
			},
		})
		runesProcessedCount += lineRuneCount
		start = end + 1
	}
	return outputs
}

// output is a run of shaped text with metadata about its position
// within a text document.
type output struct {
	Shaped    Output
	RuneRange Range
}
