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

	// o            Ordering
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

	IsRTL bool

	level level
}

type Runs struct {
	levels []level
}

// NumRuns returns the number of runs.
func (o *Runs) NumRuns() int {
	return 0 // FIXME
}

// Run returns the ith run within the ordering.
func (o *Runs) Run(i int) Run {
	return Run{} // FIXME
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

func (b *Paragraph) SegmentBytes(text string, defaultDirection DefaultDirection) Runs {
	b.text = b.text[:0]
	// The Go compiler should optimize this without allocating a string.
	for _, r := range string(text) {
		b.text = append(b.text, r)
	}
	return b.segment(defaultDirection)
}

func (b *Paragraph) segment(defaultDirection DefaultDirection) Runs {
	b.prepareInput()
	levels := b.Order(defaultDirection)
	// TODO : runs from levels
	return Runs{levels: levels}
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

// // IsLeftToRight reports whether the principle direction of rendering for this
// // paragraphs is left-to-right. If this returns false, the principle direction
// // of rendering is right-to-left.
// func (p *Paragraph) IsLeftToRight() bool {
// 	return p.Direction() == LeftToRight
// }

// // Direction returns the direction of the text of this paragraph.
// //
// // The direction may be LeftToRight, RightToLeft, Mixed, or Neutral.
// func (p *Paragraph) Direction() Direction {
// 	return p.o.Direction()
// }

// // TODO: what happens if the position is > len(input)? This should return an error.

// // RunAt reports the Run at the given position of the input text.
// //
// // This method can be used for computing line breaks on paragraphs.
// func (p *Paragraph) RunAt(pos int) Run {
// 	c := 0
// 	runNumber := 0
// 	for i, r := range p.o.runes {
// 		c += len(r)
// 		if pos < c {
// 			runNumber = i
// 		}
// 	}
// 	return p.o.Run(runNumber)
// }

// func calculateOrdering(levels []level, runes []rune) Ordering {
// 	var curDir Direction

// 	prevDir := Neutral
// 	prevI := 0

// 	o := Ordering{}
// 	// lvl = 0,2,4,...: left to right
// 	// lvl = 1,3,5,...: right to left
// 	for i, lvl := range levels {
// 		if lvl%2 == 0 {
// 			curDir = LeftToRight
// 		} else {
// 			curDir = RightToLeft
// 		}
// 		if curDir != prevDir {
// 			if i > 0 {
// 				o.runes = append(o.runes, runes[prevI:i])
// 				o.directions = append(o.directions, prevDir)
// 				o.startpos = append(o.startpos, prevI)
// 			}
// 			prevI = i
// 			prevDir = curDir
// 		}
// 	}
// 	o.runes = append(o.runes, runes[prevI:])
// 	o.directions = append(o.directions, prevDir)
// 	o.startpos = append(o.startpos, prevI)
// 	return o
// }

// Order computes the visual ordering of all the runs in a Paragraph.
func (p *Paragraph) Order(defaultDirection DefaultDirection) []level {
	if len(p.initialTypes) == 0 {
		return nil
	}

	fmt.Println(p.initialTypes)
	fmt.Println(p.pairTypes)

	lvl := level(-1)
	if defaultDirection == LeftToRight {
		lvl = 0
	} else if defaultDirection == RightToLeft {
		lvl = 1
	}

	// if err := validateTypes(p.initialTypes); err != nil {
	// 	return Ordering{}, err
	// }
	// if err := validatePbTypes(p.pairTypes); err != nil {
	// 	return Ordering{}, err
	// }
	// if err := validatePbValues(p.pairValues, p.pairTypes); err != nil {
	// 	return Ordering{}, err
	// }
	// if err := validateParagraphEmbeddingLevel(lvl); err != nil {
	// 	return Ordering{}, err
	// }

	p.embeddingLevel = lvl

	p.run()

	return p.getLevels()

	// p.o = calculateOrdering(levels, p.runes)
	// return p.o, nil
}

// // Line computes the visual ordering of runs for a single line starting and
// // ending at the given positions in the original text.
// func (p *Paragraph) Line(start, end int) (Ordering, error) {
// 	lineTypes := p.types[start:end]
// 	para, err := newParagraph(lineTypes, p.pairTypes[start:end], p.pairValues[start:end], -1)
// 	if err != nil {
// 		return Ordering{}, err
// 	}
// 	levels := para.getLevels([]int{len(lineTypes)})
// 	o := calculateOrdering(levels, p.runes[start:end])
// 	return o, nil
// }
