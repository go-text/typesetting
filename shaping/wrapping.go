package shaping

import (
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/segmenter"
	"golang.org/x/image/math/fixed"
)

// glyphIndex is the index in a Glyph slice
type glyphIndex = int

// mapRunesToClusterIndices
// returns a slice that maps rune indicies in the text to the index of the
// first glyph in the glyph cluster containing that rune in the shaped text.
// The indicies are relative to the region of runes covered by the input run.
// To translate an absolute rune index in text into a rune index into the returned
// mapping, subtract run.Runes.Offset first. If the provided buf is large enough to
// hold the return value, it will be used instead of allocating a new slice.
func mapRunesToClusterIndices(dir di.Direction, runes Range, glyphs []Glyph, buf []glyphIndex) []glyphIndex {
	if runes.Count <= 0 {
		return nil
	}
	var mapping []glyphIndex
	if cap(buf) >= runes.Count {
		mapping = buf[:runes.Count]
	} else {
		mapping = make([]glyphIndex, runes.Count)
	}
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
	segmenter  *segmenter.LineIterator
	totalRunes int
}

// newBreaker returns a breaker initialized to break the provided text.
func newBreaker(text []rune) *breaker {
	var seg segmenter.Segmenter // Note : we should cache this segmenter to reuse internal storage
	seg.Init(text)
	br := &breaker{
		segmenter:  seg.LineIterator(),
		totalRunes: len(text),
	}
	return br
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
		currentSegment := b.segmenter.Line()
		// We dont use penalties for Mandatory Breaks so far,
		// we could add it with currentSegment.IsMandatoryBreak
		option := breakOption{
			breakAtRune: currentSegment.Offset + len(currentSegment.Text) - 1,
		}
		return option, true
	}
	// Unicode rules impose to always break at the end
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

// WrapParagraph wraps the paragraph's shaped glyphs to a constant maxWidth.
// It is equivalent to iteratively invoking WrapLine with a constant maxWidth.
func (l *LineWrapper) WrapParagraph(maxWidth int, paragraph []rune, shapedRuns ...Output) []Line {
	if len(shapedRuns) == 1 && shapedRuns[0].Advance.Ceil() < maxWidth {
		return []Line{shapedRuns}
	}
	state := NewBreakState(paragraph, shapedRuns...)
	var lines []Line
	var done bool
	for !done {
		var line Line
		line, state, done = l.WrapLine(maxWidth, state)
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

// LineWrapper holds reusable state for a line wrapping operation. Reusing
// LineWrappers for multiple paragraphs should improve performance.
type LineWrapper struct {
	mapping []glyphIndex
}

// WrapLine wraps the shaped glyphs of a paragraph to a particular max width.
// It is meant to be called iteratively to wrap each line, allowing lines to
// be wrapped to different widths within the same paragraph. The returned
// BreakState should always be passed as input to the next call until the
// returned done boolean is true. Subsequent invocations with the returned
// BreakState are invalid.
func (l *LineWrapper) WrapLine(maxWidth int, state BreakState) (_ Line, _ BreakState, done bool) {
	if len(state.glyphRuns) == 0 {
		return nil, state, true
	} else if len(state.glyphRuns[0].Glyphs) == 0 {
		// Pass empty lines through as empty.
		state.glyphRuns[0].Runes = Range{Count: state.breaker.totalRunes}
		return Line([]Output{state.glyphRuns[0]}), state, true
	} else if len(state.glyphRuns) == 1 && state.glyphRuns[0].Advance.Ceil() < maxWidth {
		return Line(state.glyphRuns), state, true
	}

	lineCandidate, bestCandidate := []Output{}, []Output{}
	candidateWidth := fixed.I(0)

	// mappedRun tracks which run (if any) we have already generated rune->glyph
	// mappings for. This allows us to skip regenerating them if we need them
	// again.
	mappedRun := -1
	// mapRun performs a rune->glyph mapping for the given run, using the provided
	// run index to skip the work if that run was already mapped.
	mapRun := func(runIdx int, run Output) {
		if mappedRun != runIdx {
			l.mapping = mapRunesToClusterIndices(run.Direction, run.Runes, run.Glyphs, l.mapping)
			mappedRun = runIdx
		}
	}

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
				mapRun(state.currentRun, run)
				run = cutRun(run, l.mapping, state.lineStartRune, run.Runes.Count+run.Runes.Offset)
			}
			// While the run being processed doesn't contain the current line breaking
			// candidate, just append it to the candidate line.
			lineCandidate = append(lineCandidate, run)
			candidateWidth += run.Advance
			state.currentRun++
			run = state.glyphRuns[state.currentRun]
		}
		mapRun(state.currentRun, run)
		if !state.breaker.isValid(option, l.mapping, run) {
			// Reject invalid line break candidate and acquire a new one.
			continue
		}
		candidateRun := cutRun(run, l.mapping, state.lineStartRune, option.breakAtRune)
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
