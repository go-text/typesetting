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

// breaker generates line breaking candidates for a text.
type breaker struct {
	segmenter  *segment.Segmenter
	runeOffset int
	brokeAtEnd bool
	totalRunes int
}

// newBreaker returns a breaker initialized to break the provided text.
func newBreaker(text []rune) *breaker {
	segmenter := segment.NewSegmenter(uax14.NewLineWrap())
	segmenter.InitFromSlice(text)
	return &breaker{
		segmenter:  segmenter,
		totalRunes: len(text),
	}
}

// isValid returns whether a given option violates shaping rules (like breaking
// a shaped text cluster).
func (b *breaker) isValid(option breakOption, runeToGlyph []int, out Output) bool {
	if option.breakAtRune+1 < len(runeToGlyph) {
		// Check if this break is valid.
		gIdx := runeToGlyph[option.breakAtRune]
		g2Idx := runeToGlyph[option.breakAtRune+1]
		cIdx := out.Glyphs[gIdx].ClusterIndex
		c2Idx := out.Glyphs[g2Idx].ClusterIndex
		if cIdx == c2Idx {
			// This break is within a harfbuzz cluster, and is
			// therefore invalid.
			return false
		}
	}
	return true
}

// nextValid returns the next valid break candidate, if any. If ok is false, there are no candidates.
func (b *breaker) nextValid(currentRuneToGlyph []int, currentOutput Output) (option breakOption, ok bool) {
	option, ok = b.next()
	for ok && !b.isValid(option, currentRuneToGlyph, currentOutput) {
		option, ok = b.next()
	}
	return
}

// next returns a naive break candidate which may be invalid.
func (b *breaker) next() (option breakOption, ok bool) {
	if b.segmenter.Next() {
		penalty, _ := b.segmenter.Penalties()
		// Determine the indices of the breaking runes in the runes
		// slice. Would be nice if the API provided this.
		currentSegment := b.segmenter.Runes()
		b.runeOffset += len(currentSegment)

		// Collect all break options.
		option := breakOption{
			penalty:     penalty,
			breakAtRune: b.runeOffset - 1,
		}
		if option.breakAtRune == b.totalRunes-1 {
			b.brokeAtEnd = true
		}
		return option, true
	} else if b.totalRunes > 0 && !b.brokeAtEnd {
		return breakOption{
			penalty:     uax14.PenaltyForMustBreak,
			breakAtRune: b.totalRunes - 1,
		}, true
	}
	return breakOption{}, false
}

// Range indicates the location of a sequence of elements within a longer slice.
type Range struct {
	Offset int
	Count  int
}

// LineWrapper holds a one-dimentional, shaped text,
// to be wrapped into lines
type LineWrapper struct {
	// out - the Output that is being line-broken.
	// text is the original input text
	text []rune
	// glyphRuns is the result of the Harfbuzz shaping
	glyphRuns []Output
	// wordBreaker generates line break candidates between words.
	wordBreaker *breaker
}

// NewLineWrapper creates a line wrapper prepared to convert the given text and glyph
// data into a multi-line paragraph. The provided glyphs should be the output of
// invoking Shape on the given text.
func NewLineWrapper(text []rune, glyphRuns ...Output) LineWrapper {
	return LineWrapper{
		text:        text,
		glyphRuns:   glyphRuns,
		wordBreaker: newBreaker(text),
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
func (sp LineWrapper) shouldKeepSegmentOnLine(run Output, mapping []int, lineStartRune int, b breakOption, curLineWidth, curLineUsed, nextLineWidth length) (candidateLine Output, keep bool) {
	// Convert the break target to an inclusive index.
	glyphStart, glyphEnd := inclusiveGlyphRange(lineStartRune, b.breakAtRune, mapping, len(run.Glyphs))

	// Construct a line out of the inclusive glyph range.
	candidateLine = run
	candidateLine.Glyphs = candidateLine.Glyphs[glyphStart : glyphEnd+1]
	candidateLine.RecomputeAdvance()
	candidateAdvance := candidateLine.Advance.Ceil()
	if candidateAdvance > curLineWidth && candidateAdvance-curLineUsed <= nextLineWidth {
		// If it fits on the next line, put it there.
		return candidateLine, false
	}

	return candidateLine, true
}

// nextValidBreak returns the next line-breaking candidate position if there is one.
// If ok is false, there are no more candidates.
func (sp LineWrapper) nextValidBreak(run Output, mapping []int) (_ breakOption, ok bool) {
	return sp.wordBreaker.nextValid(mapping, run)
}

// WrapParagraph wraps the shaped glyphs of a paragraph to a particular max width.
func (sp LineWrapper) WrapParagraph(maxWidth int) []Output {
	if len(sp.glyphRuns) == 0 {
		return nil
	} else if len(sp.glyphRuns[0].Glyphs) == 0 {
		// Pass empty lines through as empty.
		sp.glyphRuns[0].Runes = Range{Count: len(sp.text)}
		return []Output{sp.glyphRuns[0]}
	}

	var outputs []Output
	start := 0
	runesProcessedCount := 0
	for _, run := range sp.glyphRuns {
		runeToGlyph := mapRunesToClusterIndices(sp.text, run.Glyphs)
		b, breakOk := sp.nextValidBreak(run, runeToGlyph)
		for breakOk {
			// Always keep the first segment on a line.
			good, _ := sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, b, maxWidth, 0, maxWidth)
			end := b.breakAtRune

			// Search through break candidates looking for candidates that can fit on the current line.
			for {
				bb, ok := sp.nextValidBreak(run, runeToGlyph)
				if !ok {
					// There are no line breaking candidates remaining.
					breakOk = false
					break
				}
				candidate, ok := sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, bb, maxWidth, good.Advance.Ceil(), maxWidth)
				if ok {
					// The break described by bb fits on this line. Use this new, longer segment.
					good = candidate
					end = bb.breakAtRune
				} else {
					// The break described by bb will not fit on this line, commit whatever the last good
					// break was and then start a new line considering this break candidate.
					b = bb
					break
				}
			}

			lineRuneCount := end - start + 1
			good.Runes = Range{
				Count:  lineRuneCount,
				Offset: runesProcessedCount,
			}
			outputs = append(outputs, good)
			runesProcessedCount += lineRuneCount
			start = end + 1
		}
	}
	return outputs
}
