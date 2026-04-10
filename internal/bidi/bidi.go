// bidi implements the Unicode Bidi algorithm
//
// The implementation is inspired by x/text/unicode/bidi.
package bidi

import (
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

	embeddingLevel Level // default: = implicitLevel;

	// at the paragraph levels
	resultTypes  []ucd.BidiClass
	resultLevels []Level

	// Index of matching PDI for isolate initiator characters. For other
	// characters, the value of matchingPDI will be set to -1. For isolate
	// initiators with no matching PDI, matchingPDI will be set to the length of
	// the input string.
	matchingPDI []int

	// Index of matching isolate initiator for PDI characters. For other
	// characters, and for PDIs with no matching isolate initiator, the value of
	// matchingIsolateInitiator will be set to -1.
	matchingIsolateInitiator []int

	runForCharacter []int

	// exclusive indice of output runs: the run is at text[previousEnd:end]
	runsEnd []int
}

// Run is a slice of text with a constant direction.
type Run struct {
	// Start and End indicate the subslice of the input text.
	Start, End int
	Level      Level
}

func (r Run) IsLeftToRight() bool { return r.Level%2 == 0 }

type Runs struct {
	levels  []Level
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
// The returned runs are only valid until the next call to [Segment], [SegmentString] or [SegmentBytes].
//
// [defaultDirection] sets the default direction for a Paragraph. The direction is
// overridden if the text contains directional characters.
func (p *Paragraph) Segment(text []rune, defaultDirection Direction) Runs {
	p.text = append(p.text[:0], text...)
	return p.segment(defaultDirection)
}

func (p *Paragraph) SegmentString(text string, defaultDirection Direction) Runs {
	p.text = p.text[:0]
	for _, r := range text {
		p.text = append(p.text, r)
	}
	return p.segment(defaultDirection)
}

func (p *Paragraph) SegmentBytes(text []byte, defaultDirection Direction) Runs {
	p.text = p.text[:0]
	// The Go compiler should optimize this without allocating a string.
	for _, r := range string(text) {
		p.text = append(p.text, r)
	}
	return p.segment(defaultDirection)
}

func (p *Paragraph) segment(defaultDirection Direction) Runs {
	p.prepareInput() // TODO: optimize for single direction
	if len(p.initialTypes) == 0 {
		return Runs{}
	}

	lvl := implicitLevel
	switch defaultDirection {
	case LeftToRight:
		lvl = 0
	case RightToLeft:
		lvl = 1
	}

	p.embeddingLevel = lvl

	p.run()
	p.computeLevels()
	return p.buildRuns()
}

func (p *Paragraph) buildRuns() Runs {
	var isRTL bool

	// lvl = 0,2,4,...: left to right
	// lvl = 1,3,5,...: right to left
	for i, lvl := range p.resultLevels {
		curIsRTL := lvl%2 != 0
		if i == 0 {
			isRTL = curIsRTL
		} else if curIsRTL != isRTL {
			// close the current run
			p.runsEnd = append(p.runsEnd, i)
			isRTL = curIsRTL
		}
	}
	// close the last run
	p.runsEnd = append(p.runsEnd, len(p.resultLevels))

	return Runs{levels: p.resultLevels, runEnds: p.runsEnd}
}

func max(a, b Level) Level {
	if a < b {
		return b
	}
	return a
}

// A Direction indicates the overall flow of text.
type Direction uint8

const (
	Neutral Direction = iota
	LeftToRight
	RightToLeft
)

// Initialize the p.pairTypes, p.pairValues and p.types from the input previously
// set by p.SetBytes() or p.SetString(). Also limit the input up to (and including) a paragraph
// separator (bidi class B).
func (p *Paragraph) prepareInput() {
	// reset storage

	p.runsEnd = p.runsEnd[:0]

	if L := len(p.text); cap(p.pairTypes) >= L {
		p.pairTypes = p.pairTypes[:L]
		p.pairValues = p.pairValues[:L]
		p.initialTypes = p.initialTypes[:L]
		p.resultTypes = p.resultTypes[:L]
		p.resultLevels = p.resultLevels[:L]
		p.matchingPDI = p.matchingPDI[:L]
		p.matchingIsolateInitiator = p.matchingIsolateInitiator[:L]
		p.runForCharacter = p.runForCharacter[:L]
	} else {
		p.pairTypes = make([]bracketType, L)
		p.pairValues = make([]rune, L)
		p.initialTypes = make([]ucd.BidiClass, L)
		p.resultTypes = make([]ucd.BidiClass, L)
		p.resultLevels = make([]Level, L)
		p.matchingPDI = make([]int, L)
		p.matchingIsolateInitiator = make([]int, L)
		p.runForCharacter = make([]int, L)
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
			p.resultLevels = p.resultLevels[:i]
			p.matchingPDI = p.matchingPDI[:i]
			p.matchingIsolateInitiator = p.matchingIsolateInitiator[:i]
			p.runForCharacter = p.runForCharacter[:i]
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
