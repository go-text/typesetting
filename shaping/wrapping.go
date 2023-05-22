package shaping

import (
	"fmt"
	"sort"

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

// mapRuneToClusterIndex finds the lowest-index glyph for the glyph cluster contiaining the rune
// at runeIdx in the source text. It uses a binary search of the glyphs in order to achieve this.
// It is equivalent to using mapRunesToClusterIndices on only a single rune index, and is thus
// more efficient for single lookups while being less efficient for runs which require many
// lookups anyway.
func mapRuneToClusterIndex(dir di.Direction, runes Range, glyphs []Glyph, runeIdx int) glyphIndex {
	var index int
	rtl := dir.Progression() == di.TowardTopLeft
	if !rtl {
		index = sort.Search(len(glyphs), func(index int) bool {
			return glyphs[index].ClusterIndex-runes.Offset > runeIdx
		})
	} else {
		index = sort.Search(len(glyphs), func(index int) bool {
			return glyphs[index].ClusterIndex-runes.Offset < runeIdx
		})
	}
	if index < 1 {
		return 0
	}
	cluster := glyphs[index-1].ClusterIndex
	if rtl && cluster-runes.Offset > runeIdx {
		return index
	}
	for index-1 >= 0 && glyphs[index-1].ClusterIndex == cluster {
		index--
	}
	return index
}

func mapRunesToClusterIndices2(dir di.Direction, runes Range, glyphs []Glyph, buf []glyphIndex) []glyphIndex {
	if runes.Count <= 0 {
		return nil
	}
	var mapping []glyphIndex
	if cap(buf) >= runes.Count {
		mapping = buf[:runes.Count]
	} else {
		mapping = make([]glyphIndex, runes.Count)
	}

	rtl := dir.Progression() == di.TowardTopLeft
	if rtl {
		for gIdx := len(glyphs) - 1; gIdx >= 0; gIdx-- {
			cluster := glyphs[gIdx].ClusterIndex
			clusterEnd := gIdx
			for gIdx-1 >= 0 && glyphs[gIdx-1].ClusterIndex == cluster {
				gIdx--
				clusterEnd = gIdx
			}
			var nextCluster int
			if gIdx-1 >= 0 {
				nextCluster = glyphs[gIdx-1].ClusterIndex
			} else {
				nextCluster = runes.Count + runes.Offset
			}
			runesInCluster := nextCluster - cluster
			clusterOffset := cluster - runes.Offset
			for i := clusterOffset; i <= runesInCluster+clusterOffset && i < len(mapping); i++ {
				mapping[i] = clusterEnd
			}
		}
	} else {
		for gIdx := 0; gIdx < len(glyphs); gIdx++ {
			cluster := glyphs[gIdx].ClusterIndex
			clusterStart := gIdx
			for gIdx+1 < len(glyphs) && glyphs[gIdx+1].ClusterIndex == cluster {
				gIdx++
			}
			var nextCluster int
			if gIdx+1 < len(glyphs) {
				nextCluster = glyphs[gIdx+1].ClusterIndex
			} else {
				nextCluster = runes.Count + runes.Offset
			}
			runesInCluster := nextCluster - cluster
			clusterOffset := cluster - runes.Offset
			for i := clusterOffset; i <= runesInCluster+clusterOffset && i < len(mapping); i++ {
				mapping[i] = clusterStart
			}
		}
	}
	return mapping
}

func mapRunesToClusterIndices3(dir di.Direction, runes Range, glyphs []Glyph, buf []glyphIndex) []glyphIndex {
	if runes.Count <= 0 {
		return nil
	}
	var mapping []glyphIndex
	if cap(buf) >= runes.Count {
		mapping = buf[:runes.Count]
	} else {
		mapping = make([]glyphIndex, runes.Count)
	}

	rtl := dir.Progression() == di.TowardTopLeft
	if rtl {
		for gIdx := len(glyphs) - 1; gIdx >= 0; {
			glyph := &glyphs[gIdx]
			// go to the start of the cluster
			gIdx -= (glyph.GlyphCount - 1)
			clusterStart := glyph.ClusterIndex - runes.Offset // map back to [0;runes.Count[
			clusterEnd := glyph.RuneCount + clusterStart
			for i := clusterStart; i <= clusterEnd && i < len(mapping); i++ {
				mapping[i] = gIdx
			}
			// go to the next cluster
			gIdx--
		}
	} else {
		for gIdx := 0; gIdx < len(glyphs); {
			glyph := &glyphs[gIdx]
			clusterStart := glyph.ClusterIndex - runes.Offset // map back to [0;runes.Count[
			clusterEnd := glyph.RuneCount + clusterStart
			for i := clusterStart; i <= clusterEnd && i < len(mapping); i++ {
				mapping[i] = gIdx
			}
			// go to the next cluster
			gIdx += glyph.GlyphCount
		}
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

// breakOption represets a location within the rune slice at which
// it may be safe to break a line of text.
type breakOption struct {
	// breakAtRune is the index at which it is safe to break.
	breakAtRune int
}

// isValid returns whether a given option violates shaping rules (like breaking
// a shaped text cluster).
func (option breakOption) isValid(runeToGlyph []int, out Output) bool {
	breakAfter := option.breakAtRune - out.Runes.Offset
	nextRune := breakAfter + 1
	if nextRune < len(runeToGlyph) && breakAfter >= 0 {
		// Check if this break is valid.
		gIdx := runeToGlyph[breakAfter]
		g2Idx := runeToGlyph[nextRune]
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

// breaker generates line breaking candidates for a text.
type breaker struct {
	segmenter  *segmenter.LineIterator
	totalRunes int
}

// newBreaker returns a breaker initialized to break the provided text.
func newBreaker(seg *segmenter.Segmenter, text []rune) *breaker {
	seg.Init(text)
	br := &breaker{
		segmenter:  seg.LineIterator(),
		totalRunes: len(text),
	}
	return br
}

// next returns a naive break candidate which may be invalid.
func (b *breaker) next() (option breakOption, ok bool) {
	if b.segmenter.Next() {
		currentSegment := b.segmenter.Line()
		// Note : we dont use penalties for Mandatory Breaks so far,
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

// WrapConfig provides line-wrapper settings.
type WrapConfig struct {
	// TruncateAfterLines is the number of lines of text to allow before truncating
	// the text. A value of zero means no limit.
	TruncateAfterLines int
	// Truncator, if provided, will be inserted at the end of a truncated line. This
	// feature is only active if TruncateAfterLines is nonzero.
	Truncator Output
	// TextContinues indicates that the paragraph wrapped by this config is not the
	// final paragraph in the text. This alters text truncation when filling the
	// final line permitted by TruncateAfterLines. If the text of this paragraph
	// does fit entirely on TruncateAfterLines, normally the truncator symbol would
	// not be inserted. However, if the overall body of text continues beyond this
	// paragraph (indicated by TextContinues), the truncator should still be inserted
	// to indicate that further paragraphs of text were truncated. This field has
	// no effect if TruncateAfterLines is zero.
	TextContinues bool
}

// WithTruncator returns a copy of WrapConfig with the Truncator field set to the
// result of shaping input with shaper.
func (w WrapConfig) WithTruncator(shaper Shaper, input Input) WrapConfig {
	w.Truncator = shaper.Shape(input)
	return w
}

// runMapper efficiently maps a run to glyph clusters.
type runMapper struct {
	// valid indicates that the mapping field is populated.
	valid bool
	// runIdx is the index of the mapped run within glyphRuns.
	runIdx int
	// mapping holds the rune->glyph mapping for the run at index mappedRun within
	// glyphRuns.
	mapping []glyphIndex
}

// mapRun updates the mapping field to be valid for the given run. It will skip the mapping
// operation if the provided runIdx is equal to the runIdx of the previous call, as the
// current mapping value is already correct.
func (r *runMapper) mapRun(runIdx int, run Output) {
	if r.runIdx != runIdx || !r.valid {
		r.mapping = mapRunesToClusterIndices3(run.Direction, run.Runes, run.Glyphs, r.mapping)
		r.runIdx = runIdx
		r.valid = true
	}
}

// RunIterator defines a type that can incrementally provide shaped text.
type RunIterator interface {
	// Next returns the next logical run index, shaped run, and true if there are more runs,
	// then it advances the iterator to the next element.
	// Otherwise it returns an undefined index, an empty Output, and false.
	Next() (index int, run Output, isValid bool)
	// Peek returns the same thing Next() would, but does not advance the iterator (so the
	// next call to Next() will return the same thing).
	Peek() (index int, run Output, isValid bool)
	// Save marks the current iterator position such that the iterator can return to it later
	// when Restore() is called. Only one position may be saved at a time, with subsequent
	// calls to Save() overriding the current value.
	Save()
	// Restore resets the iteration state to the most recently Save()-ed position.
	Restore()
}

// runSlice is a [RunIterator] built from already-shaped text.
type runSlice struct {
	runs     []Output
	idx      int
	savedIdx int
}

var _ RunIterator = (*runSlice)(nil)

// NewSliceIterator returns a [RunIterator] backed by an already-shaped slice of [Output]s.
func NewSliceIterator(outs []Output) RunIterator {
	return &runSlice{
		runs: outs,
	}
}

// Reset configures the runSlice for reuse with the given shaped text.
func (r *runSlice) Reset(outs []Output) {
	r.runs = outs
	r.idx = 0
	r.savedIdx = 0
}

// Next implements [RunIterator.Next].
func (r *runSlice) Next() (int, Output, bool) {
	idx, run, ok := r.Peek()
	if ok {
		r.idx++
	}
	return idx, run, ok
}

// Peek implements [RunIterator.Peek].
func (r *runSlice) Peek() (int, Output, bool) {
	if r.idx < len(r.runs) {
		next := r.runs[r.idx]
		return r.idx, next, true
	}
	return r.idx, Output{}, false
}

// Save implements [RunIterator.Save].
func (r *runSlice) Save() {
	r.savedIdx = r.idx
}

// Restore implements [RunIterator.Restore].
func (r *runSlice) Restore() {
	r.idx = r.savedIdx
}

// wrapBuffer provides reusable buffers for line wrapping. When using a
// wrapBuffer, returned line wrapping results will use memory stored within
// the buffer. This means that the same buffer cannot be reused for another
// wrapping operation while the wrapped lines are still in use (unless they
// are deeply copied). If necessary, using a multiple WrapBuffers can work
// around this restriction.
type wrapBuffer struct {
	// paragraph is a buffer holding paragraph allocated (primarily) from subregions
	// of the line field.
	paragraph []Line
	// line is a large buffer of Outputs that is used to build lines.
	line []Output
	// lineUsed is the index of the first unused element in line.
	lineUsed int
	// lineExhausted indicates whether the previous shaping used all of line.
	lineExhausted bool
	// alt is a smaller temporary buffer that holds candidate lines while
	// they are being built.
	alt []Output
	// altAdvance is the sum of the advances of each run in alt.
	altAdvance fixed.Int26_6
	// best is a slice holding the best known line. When possible, it
	// is a subslice of line, but if line runs out of capacity it will
	// be heap allocated.
	best []Output
	// bestInLine tracks whether the current best is allocated from within line.
	bestInLine bool
}

// NewWrapBuffer returns a WrapBuffer with some pre-allocated storage for
// small-sized texts.
func NewWrapBuffer() *wrapBuffer {
	return &wrapBuffer{
		paragraph: make([]Line, 0, 10),
		line:      make([]Output, 0, 100),
		alt:       make([]Output, 0, 10),
	}
}

func (w *wrapBuffer) reset() {
	w.paragraph = w.paragraph[:0]
	w.alt = w.alt[:0]
	w.altAdvance = 0
	w.line = w.line[:0]
	w.lineUsed = 0
	w.best = nil
	w.bestInLine = false
	if w.lineExhausted {
		w.lineExhausted = false
		// Trigger the go slice growth heuristic by appending an element to
		// the capacity.
		w.line = append(w.line[:cap(w.line)], Output{})[:0]
	}
}

func (w *wrapBuffer) stats() string {
	return fmt.Sprintf("paragraph: %d(%d), line: %d(%d), used: %d, exhausted: %v, alt: %d(%d)", len(w.paragraph), cap(w.paragraph), len(w.line), cap(w.line), w.lineUsed, w.lineExhausted, len(w.alt), cap(w.alt))
}

// singleRunParagraph is an optimized helper for quickly constructing
// a []Line containing only a single run.
func (w *wrapBuffer) singleRunParagraph(run Output) []Line {
	w.paragraph = w.paragraph[:0]
	s := w.line[w.lineUsed : w.lineUsed+1]
	s[0] = run
	w.paragraphAppend(s)
	return w.finalParagraph()
}

func (w *wrapBuffer) paragraphAppend(line []Output) {
	w.paragraph = append(w.paragraph, line)
}

func (w *wrapBuffer) finalParagraph() []Line {
	return w.paragraph
}

func (w *wrapBuffer) startLine() {
	w.alt = w.alt[:0]
	w.altAdvance = 0
	w.best = nil
	w.bestInLine = false
}

// candidateAppend adds the given run to the current line wrapping candidate.
func (w *wrapBuffer) candidateAppend(run Output) {
	w.alt = append(w.alt, run)
	w.altAdvance = w.altAdvance + run.Advance
}

func (w *wrapBuffer) candidateAdvance() fixed.Int26_6 {
	return w.altAdvance
}

// markCandidateBest marks that the current line wrapping candidate is the best
// known line wrapping candidate with the given suffixes. Providing suffixes does
// not modify the current candidate, but does ensure that the "best" candidate ends
// with them.
func (w *wrapBuffer) markCandidateBest(suffixes ...Output) {
	neededLen := len(w.alt) + len(suffixes)
	if len(w.line[w.lineUsed:cap(w.line)]) < neededLen {
		w.lineExhausted = true
		w.best = make([]Output, neededLen)
		w.bestInLine = false
	} else {
		w.best = w.line[w.lineUsed : w.lineUsed+neededLen]
		w.bestInLine = true
	}
	n := copy(w.best, w.alt)
	copy(w.best[n:], suffixes)
}

// hasBest returns whether there is currently a known valid line wrapping candidate
// for the line.
func (w *wrapBuffer) hasBest() bool {
	return len(w.best) > 0
}

// finalizeBest commits the storage for the current best line and returns it.
func (w *wrapBuffer) finalizeBest() []Output {
	if w.bestInLine {
		w.lineUsed += len(w.best)
	}
	return w.best
}

// LineWrapper holds reusable state for a line wrapping operation. Reusing
// LineWrappers for multiple paragraphs should improve performance.
type LineWrapper struct {
	// config holds the current line wrapping settings.
	config WrapConfig
	// truncating tracks whether the wrapper should be performing truncation.
	truncating bool
	// seg is an internal storage used to initiate the breaker iterator.
	seg segmenter.Segmenter

	// breaker provides line-breaking candidates.
	breaker *breaker

	// scratch holds wrapping algorithm storage buffers for reuse.
	scratch *wrapBuffer

	// mapper tracks rune->glyphCluster mappings.
	mapper runMapper
	// unusedBreak is a break requested from the breaker in a previous iteration
	// but which was not chosen as the line ending. Subsequent invocations of
	// WrapLine should start with this break.
	unusedBreak breakOption
	// isUnused indicates that the unusedBreak field is valid.
	isUnused bool
	// glyphRuns holds the runs of shaped text being wrapped.
	glyphRuns RunIterator
	// lineStartRune is the rune index of the first rune on the next line to
	// be shaped.
	lineStartRune int
	// more indicates that the iteration API has more data to return.
	more bool
}

// ensureScratch makes sure that there is an allocated scratch buffer ready for
// use by this wrapper.
func (l *LineWrapper) ensureScratch() {
	if l.scratch == nil {
		l.scratch = NewWrapBuffer()
	} else {
		l.scratch.reset()
	}
}

// Prepare initializes the LineWrapper for the given paragraph and shaped text.
// It must be called prior to invoking WrapNextLine. Prepare invalidates any
// lines previously returned by this wrapper.
func (l *LineWrapper) Prepare(config WrapConfig, paragraph []rune, runs RunIterator) {
	l.config = config
	l.truncating = l.config.TruncateAfterLines > 0
	l.breaker = newBreaker(&l.seg, paragraph)
	l.glyphRuns = runs
	l.isUnused = false
	l.lineStartRune = 0
	l.more = true
	l.mapper.valid = false
	l.ensureScratch()
}

// WrapParagraph wraps the paragraph's shaped glyphs to a constant maxWidth.
// It is equivalent to iteratively invoking WrapLine with a constant maxWidth.
// If the config has a non-zero TruncateAfterLines, WrapParagraph will return at most
// that many lines. The truncated return value is the count of runes truncated from
// the end of the text. The returned lines are only valid until the next call to
// [*LineWrapper.WrapParagraph] or [*LineWrapper.Prepare].
func (l *LineWrapper) WrapParagraph(config WrapConfig, maxWidth int, paragraph []rune, runs RunIterator) (_ []Line, truncated int) {
	l.ensureScratch()
	// Check whether we can skip line wrapping altogether for the simple single-run-that-fits case.
	runs.Save()
	_, firstRun, firstOk := runs.Next()
	if _, _, ok := runs.Peek(); firstOk && !ok && firstRun.Advance.Ceil() < maxWidth && !(config.TextContinues && config.TruncateAfterLines == 1) {
		return l.scratch.singleRunParagraph(firstRun), 0
	}
	runs.Restore()

	l.Prepare(config, paragraph, runs)
	var done bool
	for !done {
		var line Line
		line, truncated, done = l.WrapNextLine(maxWidth)
		if line != nil {
			l.scratch.paragraphAppend(line)
		}
	}
	return l.scratch.finalParagraph(), truncated
}

// nextBreakOption returns the next rune offset at which the line can be broken,
// if any. If it returns false, there are no more candidates.
func (l *LineWrapper) nextBreakOption() (breakOption, bool) {
	var option breakOption
	if l.isUnused {
		option = l.unusedBreak
		l.isUnused = false
	} else {
		var breakOk bool
		option, breakOk = l.breaker.next()
		if !breakOk {
			return option, false
		}
		l.unusedBreak = option
	}
	return option, true
}

type fillResult uint8

const (
	// noCandidate indicates that it is not possible to compose a new line candidate using the provided
	// breakOption, so the best known line should be used instead.
	noCandidate fillResult = iota
	// noRunWithBreak indicates that none of the runs available to the line wrapper contain the break
	// option, so the returned candidate is the best option.
	noRunWithBreak
	// newCandidate indicates that the returned line candidate is valid.
	newCandidate
)

// fillUntil tries to fill the provided line candidate slice with runs until it reaches a run containing the
// provided break option. It returns the index of the run containing the option, the new width of the candidate
// line, the contents of the new candidate line, and a result indicating how to proceed.
func (l *LineWrapper) fillUntil(runs RunIterator, option breakOption, scratch *wrapBuffer) (status fillResult) {
	currRunIndex, run, more := runs.Peek()
	for more && option.breakAtRune >= run.Runes.Count+run.Runes.Offset {
		if l.lineStartRune >= run.Runes.Offset+run.Runes.Count {
			_, _, isMore := runs.Next()
			if !isMore {
				return noCandidate
			}
			currRunIndex, run, more = runs.Peek()
			continue
		} else if l.lineStartRune > run.Runes.Offset {
			// If part of this run has already been used on a previous line, trim
			// the runes corresponding to those glyphs off.
			l.mapper.mapRun(currRunIndex, run)
			run = cutRun(run, l.mapper.mapping, l.lineStartRune, run.Runes.Count+run.Runes.Offset)
		}
		// While the run being processed doesn't contain the current line breaking
		// candidate, just append it to the candidate line.
		scratch.candidateAppend(run)
		_, _, isMore := runs.Next()
		if !isMore {
			return noRunWithBreak
		}
		currRunIndex, run, more = runs.Peek()
	}
	return newCandidate
}

// lineConfig tracks settings for line wrapping a single line of text.
type lineConfig struct {
	// truncating indicates whether this line is being truncated (if sufficiently long).
	truncating bool
	// maxWidth is the maximum space a line can occupy.
	maxWidth int
	// truncatedMaxWidth holds the maximum width of the line available for text if the truncator
	// is occupying part of the line.
	truncatedMaxWidth int
}

// WrapNextLine wraps the shaped glyphs of a paragraph to a particular max width.
// It is meant to be called iteratively to wrap each line, allowing lines to
// be wrapped to different widths within the same paragraph. When done is true,
// subsequent calls to WrapNextLine (without calling Prepare) will return a nil line.
// The truncated return value is the count of runes truncated from the end of the line,
// if this line was truncated. The returned line is only valid until the next call to
// [*LineWrapper.Prepare] or [*LineWrapper.WrapParagraph].
func (l *LineWrapper) WrapNextLine(maxWidth int) (finalLine Line, truncated int, done bool) {
	// If we've already finished the paragraph, don't do any more work.
	if !l.more {
		return nil, 0, true
	}
	defer func() {
		// Update the start position of the next line.
		if len(finalLine) > 0 {
			finalRun := finalLine[len(finalLine)-1]
			l.lineStartRune = finalRun.Runes.Count + finalRun.Runes.Offset
		}
		// Check whether we've exhausted the text.
		done = done || l.lineStartRune >= l.breaker.totalRunes
		// Implement truncation if needed.
		if l.truncating {
			l.config.TruncateAfterLines--
			insertTruncator := false
			if l.config.TruncateAfterLines == 0 {
				done = true
				truncated = l.breaker.totalRunes - l.lineStartRune
				insertTruncator = truncated > 0 || l.config.TextContinues
			}
			if insertTruncator {
				finalLine = append(finalLine, l.config.Truncator)
			}
		}
		// Mark the paragraph as complete if needed.
		if done {
			l.more = false
		}
	}()
	// Save current iterator state so we can peek ahead.
	l.glyphRuns.Save()
	// If the iterator is empty, return early.
	_, nextRun, more := l.glyphRuns.Next()
	if !more {
		return nil, 0, true
	}
	// If the iterator contains only one run, and that run fits, take the fast path.
	l.scratch.startLine()
	truncating := l.config.TruncateAfterLines == 1
	_, _, moreRuns := l.glyphRuns.Peek()
	emptyLine := len(nextRun.Glyphs) == 0
	if emptyLine ||
		(!moreRuns && // This is the final run in the iterator.
			nextRun.Runes.Offset == l.lineStartRune && // We have not already used part of the run.
			nextRun.Advance.Ceil() < maxWidth && // The run fits in the line.
			!(l.config.TextContinues && truncating)) {
		if emptyLine {
			// Pass empty lines through as empty.
			nextRun.Runes = Range{Count: l.breaker.totalRunes}
		}
		l.scratch.candidateAppend(nextRun)
		l.scratch.markCandidateBest()
		return l.scratch.finalizeBest(), 0, true
	}
	// Restore iterator state in preparation for real line wrapping algorithm.
	l.glyphRuns.Restore()

	config := lineConfig{
		truncating:        truncating,
		maxWidth:          maxWidth,
		truncatedMaxWidth: maxWidth - l.config.Truncator.Advance.Ceil(),
	}
	done = l.wrapNextLine(config)
	finalLine = l.scratch.finalizeBest()
	return finalLine, 0, done
}

// wrapNextLine iteratively processes line breaking candidates, building a line within the
// wrapper's scratch [WrapBuffer]. It returns whether the paragraph is finished once it has
// successfully built a line.
func (l *LineWrapper) wrapNextLine(config lineConfig) (done bool) {
	for option, ok := l.nextBreakOption(); ok; option, ok = l.nextBreakOption() {
		switch l.processBreakOption(option, config) {
		case breakInvalid, fits:
			continue
		case endLine:
			return true
		case newLine:
			return false
		case cannotFit:
			if config.truncating {
				// Any candidate that we have accumulated is going to use space
				// that isn't valid, so we force the use of the (empty) best result
				// here.
				// TODO(whereswaldon): when implementing grapheme cluster boundary
				// wrapping, drop this logic.
				return true
			} else {
				return false
			}
		}
		// TODO(whereswaldon): segment using UAX#29 grapheme clustering here and try
		// breaking again using only those boundaries to find a viable break in cases
		// where no UAX#14 breaks were viable above.
	}
	return true
}

type processBreakResult uint8

const (
	// breakInvalid indicates that the provided break candidate is incompatible with the shaped
	// text and should be skipped.
	breakInvalid processBreakResult = iota
	// endLine indicates that the returned best []Output contains a full line terminating the text.
	endLine
	// newLine indicates that the returned best []Output contains a full line that does not terminate the text.
	newLine
	// fits indicates that the text up to the break option fit within the line and that another break
	// option can be attempted.
	fits
	// cannotFit indicates that this is the first break option on the line, and that even this option cannot
	// fit within the space available.
	cannotFit
)

// processBreakOption evaluates whether the provided breakOption can fit onto the current line wrapping line.
func (l *LineWrapper) processBreakOption(option breakOption, config lineConfig) processBreakResult {
	l.glyphRuns.Save()
	result := l.fillUntil(l.glyphRuns, option, l.scratch)

	if result == noCandidate {
		return endLine
	} else if result == noRunWithBreak {
		l.scratch.markCandidateBest()
		return endLine
	}

	currRunIndex, run, _ := l.glyphRuns.Peek()
	l.mapper.mapRun(currRunIndex, run)
	if !option.isValid(l.mapper.mapping, run) {
		// Reject invalid line break candidate and acquire a new one.
		return breakInvalid
	}
	candidateRun := cutRun(run, l.mapper.mapping, l.lineStartRune, option.breakAtRune)
	candidateLineWidth := (candidateRun.Advance + l.scratch.candidateAdvance()).Ceil()
	if candidateLineWidth > config.maxWidth {
		// The run doesn't fit on the line.
		if !l.scratch.hasBest() {
			if config.truncating {
				return endLine
			}
			// There is no existing candidate that fits, and we have just hit the
			// first line breaking candidate. Commit this break position as the
			// best available, even though it doesn't fit.
			l.scratch.markCandidateBest(candidateRun)
			return cannotFit
		} else {
			// The line is a valid, shorter wrapping. Return it and mark that
			// we should reuse the current line break candidate on the next
			// line.
			l.isUnused = true
			l.glyphRuns.Restore()
			return newLine
		}
	} else if config.truncating && candidateLineWidth > config.truncatedMaxWidth {
		// The run would not fit if truncated.
		finalRunRune := candidateRun.Runes.Count + candidateRun.Runes.Offset
		if finalRunRune == l.breaker.totalRunes && !l.config.TextContinues {
			// The run contains the entire end of the text, so no truncation is
			// necessary.
			l.scratch.markCandidateBest(candidateRun)
		}
		// We must truncate the line in order to show it.
		return endLine
	} else {
		// The run does fit on the line. Commit this line as the best known
		// line, but keep lineCandidate unmodified so that later break
		// options can be attempted to see if a more optimal solution is
		// available.
		l.scratch.markCandidateBest(candidateRun)
		return fits
	}
}
