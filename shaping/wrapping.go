package shaping

import (
	"github.com/gioui/uax/segment"
	"github.com/gioui/uax/uax14"
	"github.com/go-text/typesetting/di"
	"golang.org/x/image/math/fixed"
)

// glyphIndex is the index in a Glyph slice
type glyphIndex = int

// length is the unit used to measure a width
type length = int

// mapRunesToClusterIndices
// returns a slice that maps rune indicies in the text to the index of the
// first glyph in the glyph cluster containing that rune in the shaped text.
// The indicies are relative to the region of runes covered by the input run.
// To translate an absolute rune index in text into a rune index into the returned
// mapping, subtract run.Runes.Offset first.
func mapRunesToClusterIndices(dir di.Direction, runes Range, glyphs []Glyph) []glyphIndex {
	if runes.Count <= 0 {
		return nil
	}
	mapping := make([]glyphIndex, runes.Count)
	glyphCursor := 0
	rtl := dir.Progression() == di.TowardTopLeft
	if rtl {
		glyphCursor = len(glyphs) - 1
	}
	// off tracks the offset position of the glyphs from the first rune of the
	// shaped text. This must be subtracted from all cluster indicies in order to
	// normalize them into the range [0,runes.Count).
	off := runes.Offset
	for i := 0; i < runes.Count; i++ {
		for glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			((rtl && glyphs[glyphCursor].ClusterIndex-off <= i) ||
				(!rtl && glyphs[glyphCursor].ClusterIndex-off < i)) {
			if rtl {
				glyphCursor--
			} else {
				glyphCursor++
			}
		}
		if rtl {
			glyphCursor++
		} else if (glyphCursor >= 0 && glyphCursor < len(glyphs) &&
			glyphs[glyphCursor].ClusterIndex-off > i) ||
			(glyphCursor == len(glyphs) && len(glyphs) > 1) {
			glyphCursor--
			targetClusterIndex := glyphs[glyphCursor].ClusterIndex - off
			for glyphCursor-1 >= 0 && glyphs[glyphCursor-1].ClusterIndex-off == targetClusterIndex {
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
func inclusiveGlyphRange(dir di.Direction, start, breakAfter int, runeToGlyph []int, numGlyphs int) (glyphStart, glyphEnd glyphIndex) {
	rtl := dir.Progression() == di.TowardTopLeft
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

// Line holds runs of shaped text wrapped onto a single line. All the contained
// Output should be displayed sequentially on one line.
type Line []Output

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
func (sp LineWrapper) shouldKeepSegmentOnLine(run Output, mapping []int, lineStartRune int, b breakOption, curLineWidth, curLineUsed, nextLineWidth length) (candidateLine Output, keep, more bool) {
	// Convert the break target to an inclusive index.
	runeStart := lineStartRune - run.Runes.Offset
	runeEnd := b.breakAtRune - run.Runes.Offset
	if runeStart < 0 {
		// If the start location is prior to the run of shaped text under consideration,
		// just work from the beginning of this run.
		runeStart = 0
	}
	if runeEnd >= len(mapping) {
		// If the break location is after the entire run of shaped text,
		// keep through the end of the run.
		runeEnd = len(mapping) - 1
		// Set the more return value to true, indicating that subsequent runs should be
		// appended to the line candidate.
		more = true
	}
	glyphStart, glyphEnd := inclusiveGlyphRange(run.Direction, runeStart, runeEnd, mapping, len(run.Glyphs))

	// Construct a line out of the inclusive glyph range.
	candidateLine = run
	candidateLine.Glyphs = candidateLine.Glyphs[glyphStart : glyphEnd+1]
	candidateLine.RecomputeAdvance()
	candidateLine.Runes.Offset = run.Runes.Offset + runeStart
	candidateLine.Runes.Count = runeEnd - runeStart + 1
	candidateAdvance := candidateLine.Advance.Ceil()
	if candidateAdvance > curLineWidth && candidateAdvance-curLineUsed <= nextLineWidth {
		// If it fits on the next line, put it there.
		return candidateLine, false, more
	}

	return candidateLine, true, more
}

// nextValidBreak returns the next line-breaking candidate position if there is one.
// If ok is false, there are no more candidates.
func (sp LineWrapper) nextValidBreak(run Output, mapping []int) (_ breakOption, ok bool) {
	return sp.wordBreaker.nextValid(mapping, run)
}

// WrapParagraph wraps the shaped glyphs of a paragraph to a particular max width.
func (sp LineWrapper) WrapParagraph(maxWidth int) []Line {
	if len(sp.glyphRuns) == 0 {
		return nil
	} else if len(sp.glyphRuns[0].Glyphs) == 0 {
		// Pass empty lines through as empty.
		sp.glyphRuns[0].Runes = Range{Count: len(sp.text)}
		return []Line{Line([]Output{sp.glyphRuns[0]})}
	}

	var outputs []Line
	start := 0

	runIdx := 0
	run := sp.glyphRuns[runIdx]
	runeToGlyph := mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)

	b, breakOk := sp.nextValidBreak(run, runeToGlyph)
	for breakOk {
		var goodLine []Output
		var goodLineWidth fixed.Int26_6
		lineRunIndex := runIdx
		// Always keep the first segment on a line.
		good, _, more := sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, b, maxWidth, 0, maxWidth)
		goodLine = append(goodLine, good)
		goodLineWidth += good.Advance
		for more {
			lineRunIndex++
			run = sp.glyphRuns[lineRunIndex]
			runeToGlyph = mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
			good, _, more = sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, b, maxWidth, goodLineWidth.Ceil(), maxWidth)
			goodLine = append(goodLine, good)
			goodLineWidth += good.Advance
		}
		end := b.breakAtRune

		// Search through break candidates looking for candidates that can fit on the current line.
		for {
			if lineRunIndex != runIdx {
				// If we've already traversed forward through the runs, reset to the beginning
				// of the line.
				run = sp.glyphRuns[runIdx]
				runeToGlyph = mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
				lineRunIndex = runIdx
			}
			var candidateLine []Output
			var candidateLineWidth fixed.Int26_6
			bb, ok := sp.nextValidBreak(run, runeToGlyph)
			if !ok {
				// There are no line breaking candidates remaining.
				breakOk = false
				break
			}
			candidate, ok, more := sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, bb, maxWidth, goodLineWidth.Ceil(), maxWidth)
			candidateLine = append(candidateLine, candidate)
			candidateLineWidth += candidate.Advance
			for more && ok {
				lineRunIndex++
				run = sp.glyphRuns[lineRunIndex]
				runeToGlyph = mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
				candidate, ok, more = sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, bb, maxWidth, candidateLineWidth.Ceil(), maxWidth)
				candidateLine = append(candidateLine, candidate)
				candidateLineWidth += candidate.Advance
			}
			if ok {
				// The break described by bb fits on this line. Use this new, longer segment.
				goodLine = candidateLine
				goodLineWidth = candidateLineWidth
				end = bb.breakAtRune
			} else {
				// The break described by bb will not fit on this line, commit whatever the last good
				// break was and then start a new line considering this break candidate.
				b = bb
				break
			}
		}
		outputs = append(outputs, goodLine)
		start = end + 1
	}
	return outputs
}
