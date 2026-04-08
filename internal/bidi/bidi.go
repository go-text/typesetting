// bidi implements the Unicode Bidi algorithm
//
// The implementation is inspired by x/text/unicode/bidi.
package bidi

import (
	"fmt"

	"github.com/go-text/typesetting/internal/unicodedata"
	ucd "github.com/go-text/typesetting/internal/unicodedata"
)

// Paragraph is the main entry point of the package.
//
// It holds a single text for Bidi processing,
// stores internal data required to segment a string,
// and should be reused to reduce allocations.
type Paragraph struct {
	text []rune // input values

	initialTypes []ucd.BidiClass
	pairTypes    []bracketType
	pairValues   []rune

	embeddingLevel level // default: = implicitLevel;

	// at the paragraph levels
	resultTypes  []ucd.BidiClass
	resultLevels []level

	// TODO: holds enough for run computation

	// Index of matching PDI for isolate initiator characters. For other
	// characters, the value of matchingPDI will be set to -1. For isolate
	// initiators with no matching PDI, matchingPDI will be set to the length of
	// the input string.
	matchingPDI []int

	// Index of matching isolate initiator for PDI characters. For other
	// characters, and for PDIs with no matching isolate initiator, the value of
	// matchingIsolateInitiator will be set to -1.
	matchingIsolateInitiator []int
}

// Run is a slice of text with a constant direction.
type Run struct {
	// Start and End indicate the subslice of the input text.
	Start, End int
	level level
}

func (r Run) IsLeftToRight() bool { return r.level % 2 != 0 }

type Runs struct {
	levels []level
	runEnds []int
}

// NumRuns returns the number of runs.
func (r *Runs) NumRuns() int { return len(r.runEnds) }

// Run returns the ith run of segmented text.
// This method panics if [i] is not in the range [0,NumRuns()[
func (r *Runs) Run(i int) Run {
	start := 0
	if i != 0 {
		start = r.runEnds[i-1]
	}
	end := r.runEnds[i]
	level := r.levels[start]
	return Run{start, end, level} 
}

// Segment applies the Bidi algorithm.
// The returned iterator is only valid until the next call to [Segment].
//
// [defaultDirection] sets the default direction for a Paragraph. The direction is
// overridden if the text contains directional characters.
func (b *Paragraph) Segment(text []rune, defaultDirection DefaultDirection) Runs {
	b.text = append(b.text[:0], text...)
	return b.segment(defaultDirection)
}

func (b *Paragraph) SegmentString(text string, defaultDirection DefaultDirection) Runs {
	b.text = b.text[:0]
	for _, r := range text {
		b.text = append(b.text, r)
	}
	return b.segment(defaultDirection)
}

func (b *Paragraph) SegmentBytes(text []byte, defaultDirection DefaultDirection) Runs {
	b.text = b.text[:0]
	// The Go compiler should optimize this without allocating a string.
	for _, r := range string(text) {
		b.text = append(b.text, r)
	}
	return b.segment(defaultDirection)
}

func (b *Paragraph) segment(defaultDirection DefaultDirection) Runs {
	b.prepareInput() // TODO: optimize for single direction 

	lvl := level(-1)
	if defaultDirection == LeftToRight {
		lvl = 0
	} else if defaultDirection == RightToLeft {
		lvl = 1
	}

	p.embeddingLevel = lvl

	p.run()
	levels := p.getLevels()
	return p.buildRuns(levels)
}

func (b *Paragraph) buildRuns(levels []level) Runs {
	var (
		isRTL bool
		// TODO: allocate only once in Paragrpah
		runEnds []int // exclusive indice : the run is at text[previousEnd:end] 
	)

	// lvl = 0,2,4,...: left to right
	// lvl = 1,3,5,...: right to left
	for i, lvl := range levels {
		curIsRTL := lvl%2 != 0
		if i == 0 {
			isRTL = curIsRTL
		} else if curIsRTL != isRTL {
			// close the current run 
			runEnds = append(runEnds, i)
			isRTL = curIsRTL
		}
	}
	// close the last run 
	runEnds = append(runEnds, len(levels))

	return Runs{levels:levels, runEnds:runEnds}
}


type charType = unicodedata.BidiClass

type Level = level

type ParType = charType

func max(a, b level) level {
	if a < b {
		return b
	}
	return a
}

// A Direction indicates the overall flow of text.
type DefaultDirection uint8

const (
	Neutral DefaultDirection = iota
	LeftToRight
	RightToLeft
)

// Initialize the p.pairTypes, p.pairValues and p.types from the input previously
// set by p.SetBytes() or p.SetString(). Also limit the input up to (and including) a paragraph
// separator (bidi class B).
//
// The function p.Order() needs these values to be set, so this preparation could be postponed.
// But since the SetBytes and SetStrings functions return the length of the input up to the paragraph
// separator, the whole input needs to be processed anyway and should not be done twice.
//
// The function has the same return values as SetBytes() / SetString()
func (p *Paragraph) prepareInput() {
	// clear slices from previous SetString or SetBytes
	if L := len(p.text); cap(p.pairTypes) < L {
		p.pairTypes = make([]bracketType, L)
		p.pairValues = make([]rune, L)
		p.initialTypes = make([]ucd.BidiClass, L)
		p.resultTypes = make([]ucd.BidiClass, L)
	} else {
		p.pairTypes = p.pairTypes[:L]
		p.pairValues = p.pairValues[:L]
		p.initialTypes = p.initialTypes[:L]
		p.resultTypes = p.resultTypes[:L]
	}

	for i, r := range p.text {
		cls, bracket := ucd.LookupBidiClass(r)
		if cls == ucd.BD_B {
			// Unlikely, but trim the arrays and exit
			p.text = p.text[:i]
			p.pairTypes = p.pairTypes[:i]
			p.pairValues = p.pairValues[:i]
			p.initialTypes = p.initialTypes[:i]
			p.resultTypes = p.resultTypes[:i]
			return
		}
		p.initialTypes[i] = cls
		p.resultTypes[i] = cls
		if bracket.IsOpening() {
			p.pairTypes[i] = bpOpen
			p.pairValues[i] = r
		} else if bracket.IsBracket() {
			// this must be a closing bracket,
			// since IsOpeningBracket is not true
			p.pairTypes[i] = bpClose
			p.pairValues[i] = bracket.Reverse(r)
		} else {
			p.pairTypes[i] = bpNone
			p.pairValues[i] = 0
		}
	}
}
