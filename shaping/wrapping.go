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
func (sp LineWrapper) shouldKeepSegmentOnLine(run Output, mapping []int, lineStartRune int, b breakOption, curLineWidth, curLineUsed, nextLineWidth length) (candidateLine Output, fits, incomplete, consumed bool) {
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
		incomplete = true
	}
	glyphStart, glyphEnd := inclusiveGlyphRange(run.Direction, runeStart, runeEnd, mapping, len(run.Glyphs))

	// Construct a line out of the inclusive glyph range.
	candidateLine = run
	candidateLine.Glyphs = candidateLine.Glyphs[glyphStart : glyphEnd+1]
	candidateLine.RecomputeAdvance()
	candidateLine.Runes.Offset = run.Runes.Offset + runeStart
	candidateLine.Runes.Count = runeEnd - runeStart + 1
	candidateAdvance := candidateLine.Advance.Ceil()

	candidateConsumed := candidateLine.Runes.Offset+candidateLine.Runes.Count == run.Runes.Offset+run.Runes.Count
	if candidateAdvance > curLineWidth && candidateAdvance-curLineUsed <= nextLineWidth {
		// If it fits on the next line, put it there.
		return candidateLine, false, incomplete, candidateConsumed
	}

	return candidateLine, true, incomplete, candidateConsumed
}

// nextValidBreak returns the next line-breaking candidate position if there is one.
// If ok is false, there are no more candidates.
func (sp LineWrapper) nextValidBreak(run Output, mapping []int) (_ breakOption, ok bool) {
	return sp.wordBreaker.nextValid(mapping, run)
}

// BreakState holds intermediate line-wrapping state for breaking runs of text
// across multiple lines.
type BreakState struct {
	breaker *breaker
	// unusedBreak is a break requested from the breaker in a previous iteration
	// but which was not chosen as the line ending. Subsequent invocations of
	// WrapLine should start with this break.
	unusedBreak breakOption
	// isUnused indicates that the unusedBreak field is valid.
	isUnused bool
	// glyphRuns holds the runs of shaped text being wrapped.
	glyphRuns []Output
	// currentRun holds the index in use within glyphRuns.
	currentRun int
	// lineStartRune is the rune index of the first rune on the next line to
	// be shaped.
	lineStartRune int
}

// NewBreakState initializes a BreakState for the given paragraph of text.
func NewBreakState(paragraph []rune, shapedRuns ...Output) BreakState {
	return BreakState{
		breaker:   newBreaker(paragraph),
		glyphRuns: shapedRuns,
	}
}

// WrapParagraph2 wraps the paragraph's shaped glyphs to a constant maxWidth.
// It is equivalent to iteratively invoking WrapLine with a constant maxWidth.
func WrapParagraph2(maxWidth int, paragraph []rune, shapedRuns ...Output) []Line {
	state := NewBreakState(paragraph, shapedRuns...)
	var lines []Line
	var done bool
	for !done {
		var line Line
		line, state, done = WrapLine(maxWidth, state)
		lines = append(lines, line)
	}
	return lines
}

// cutRun returns the sub-run of run containing glyphs corresponding to the provided
// _inclusive_ rune range.
func cutRun(run Output, mapping []glyphIndex, startRune, endRune int) Output {
	// Convert the rune range of interest into an inclusive range within the
	// current run's runes.
	runeStart := startRune - run.Runes.Offset
	runeEnd := endRune - run.Runes.Offset
	if runeStart < 0 {
		// If the start location is prior to the run of shaped text under consideration,
		// just work from the beginning of this run.
		runeStart = 0
	}
	if runeEnd >= len(mapping) {
		// If the break location is after the entire run of shaped text,
		// keep through the end of the run.
		runeEnd = len(mapping) - 1
	}
	glyphStart, glyphEnd := inclusiveGlyphRange(run.Direction, runeStart, runeEnd, mapping, len(run.Glyphs))

	// Construct a run out of the inclusive glyph range.
	run.Glyphs = run.Glyphs[glyphStart : glyphEnd+1]
	run.RecomputeAdvance()
	run.Runes.Offset = run.Runes.Offset + runeStart
	run.Runes.Count = runeEnd - runeStart + 1
	return run
}

// WrapLine wraps the shaped glyphs of a paragraph to a particular max width.
// It is meant to be called iteratively to wrap each line, allowing lines to
// be wrapped to different widths within the same paragraph. The returned
// BreakState should always be passed as input to the next call until the
// returned done boolean is true. Subsequent invocations with the returned
// BreakState are invalid.
func WrapLine(maxWidth int, state BreakState) (_ Line, _ BreakState, done bool) {
	if len(state.glyphRuns) == 0 {
		return nil, state, true
	} else if len(state.glyphRuns[0].Glyphs) == 0 {
		// Pass empty lines through as empty.
		state.glyphRuns[0].Runes = Range{Count: state.breaker.totalRunes}
		return Line([]Output{state.glyphRuns[0]}), state, true
	}

	lineCandidate, bestCandidate := []Output{}, []Output{}
	candidateWidth := fixed.I(0)

	for {
		run := state.glyphRuns[state.currentRun]
		var option breakOption
		if state.isUnused {
			option = state.unusedBreak
			state.isUnused = false
		} else {
			var breakOk bool
			option, breakOk = state.breaker.next()
			if !breakOk {
				return bestCandidate, state, true
			}
			state.unusedBreak = option
		}
		for option.breakAtRune >= run.Runes.Count+run.Runes.Offset {
			if state.lineStartRune > run.Runes.Offset {
				// If part of this run has already been used on a previous line, trim
				// the runes corresponding to those glyphs off.
				mapping := mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
				run = cutRun(run, mapping, state.lineStartRune, run.Runes.Count+run.Runes.Offset)
			}
			// While the run being processed doesn't contain the current line breaking
			// candidate, just append it to the candidate line.
			lineCandidate = append(lineCandidate, run)
			candidateWidth += run.Advance
			state.currentRun++
			run = state.glyphRuns[state.currentRun]
		}
		mapping := mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
		if !state.breaker.isValid(option, mapping, run) {
			// Reject invalid line break candidate and acquire a new one.
			continue
		}
		candidateRun := cutRun(run, mapping, state.lineStartRune, option.breakAtRune)
		if (candidateRun.Advance + candidateWidth).Ceil() > maxWidth {
			// The run doesn't fit on the line.
			if len(bestCandidate) < 1 {
				// There is no existing candidate that fits, and we have just hit the
				// first line breaking canddiate. Commit this break position as the
				// best available, even though it doesn't fit.
				lineCandidate = append(lineCandidate, candidateRun)
				state.lineStartRune = candidateRun.Runes.Offset + candidateRun.Runes.Count
				return lineCandidate, state, false
			} else {
				// The line is a valid, shorter wrapping. Return it and mark that
				// we should reuse the current line break candidate on the next
				// line.
				state.isUnused = true
				finalRunRunes := bestCandidate[len(bestCandidate)-1].Runes
				state.lineStartRune = finalRunRunes.Count + finalRunRunes.Offset
				return bestCandidate, state, false
			}
		} else {
			// The run does fit on the line. Commit this line as the best known
			// line, but keep lineCandidate unmodified so that later break
			// options can be attempted to see if a more optimal solution is
			// available.
			if target := len(lineCandidate) + 1; cap(bestCandidate) < target {
				bestCandidate = make([]Output, target-1, target)
			} else if len(bestCandidate) < target {
				bestCandidate = bestCandidate[:target-1]
			}
			bestCandidate = bestCandidate[:copy(bestCandidate, lineCandidate)]
			bestCandidate = append(bestCandidate, candidateRun)
		}
	}
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
		goodLineRunIndex := runIdx
		// Always keep the first segment on a line.
		good, _, more, consumed := sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, b, maxWidth, 0, maxWidth)
		goodLine = append(goodLine, good)
		goodLineWidth += good.Advance
		for more {
			goodLineRunIndex++
			run = sp.glyphRuns[goodLineRunIndex]
			runeToGlyph = mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
			good, _, more, consumed = sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, b, maxWidth-goodLineWidth.Ceil(), 0, maxWidth)
			goodLine = append(goodLine, good)
			goodLineWidth += good.Advance
		}
		if consumed {
			goodLineRunIndex++
		}
		end := b.breakAtRune

		// Search through break candidates looking for candidates that can fit on the current line.
		for {
			candidateLineRunIndex := goodLineRunIndex
			if candidateLineRunIndex != runIdx {
				// If we've already traversed forward through the runs, reset to the beginning
				// of the line.
				run = sp.glyphRuns[runIdx]
				runeToGlyph = mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
				candidateLineRunIndex = runIdx
			}
			var candidateLine []Output
			var candidateLineWidth fixed.Int26_6
			bb, ok := sp.nextValidBreak(run, runeToGlyph)
			if !ok {
				// There are no line breaking candidates remaining.
				breakOk = false
				break
			}
			candidate, ok, more, candidateConsumed := sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, bb, maxWidth, goodLineWidth.Ceil(), maxWidth)
			candidateLine = append(candidateLine, candidate)
			candidateLineWidth += candidate.Advance
			for more && ok {
				candidateLineRunIndex++
				run = sp.glyphRuns[candidateLineRunIndex]
				runeToGlyph = mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs)
				candidate, ok, more, candidateConsumed = sp.shouldKeepSegmentOnLine(run, runeToGlyph, start, bb, maxWidth-candidateLineWidth.Ceil(), 0, maxWidth)
				candidateLine = append(candidateLine, candidate)
				candidateLineWidth += candidate.Advance
			}
			if candidateConsumed {
				candidateLineRunIndex++
			}
			if ok {
				// The break described by bb fits on this line. Use this new, longer segment.
				goodLine = candidateLine
				goodLineWidth = candidateLineWidth
				goodLineRunIndex = candidateLineRunIndex
				end = bb.breakAtRune
			} else {
				// The break described by bb will not fit on this line, commit whatever the last good
				// break was and then start a new line considering this break candidate.
				b = bb
				break
			}
		}
		runIdx = goodLineRunIndex
		outputs = append(outputs, goodLine)
		start = end + 1
	}
	return outputs
}
