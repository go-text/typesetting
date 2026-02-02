// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package segmenter

import (
	"unicode"

	ucd "github.com/go-text/typesetting/unicodedata"
)

// Apply the Line Breaking Rules and returns the computed break opportunity
// See https://unicode.org/reports/tr14/#BreakingRules
func (cr *cursor) applyLineBoundaryRules() breakOpportunity {
	// start by attributing the break class for the current rune
	cr.ruleLB1()

	triggerNumSequence := cr.updateNumSequence()

	// add the line break rules in reverse order to override
	// the lower priority rules.
	breakOp := breakEmpty

	cr.ruleLB30(&breakOp)
	cr.ruleLB30ab(&breakOp)
	cr.ruleLB29To26(&breakOp)
	cr.ruleLB25(&breakOp, triggerNumSequence)
	cr.ruleLB24To22(&breakOp)
	cr.ruleLB21To8(&breakOp)
	cr.ruleLB7To4(&breakOp)

	return breakOp
}

// breakOpportunity is a convenient enum,
// mapped to the LineBreak and MandatoryBreak properties,
// avoiding too many bit operations
type breakOpportunity uint8

const (
	breakEmpty      breakOpportunity = iota // not specified
	breakProhibited                         // no break
	breakAllowed                            // direct break (can always break here)
	breakMandatory                          // break is mandatory (implies breakAllowed)
)

func (cr *cursor) ruleLB30(breakOp *breakOpportunity) {
	// (AL | HL | NU) × [OP-[\p{ea=F}\p{ea=W}\p{ea=H}]]
	if (cr.prevLine == ucd.BreakAL || cr.prevLine == ucd.BreakHL || cr.prevLine == ucd.BreakNU) &&
		cr.line == ucd.BreakOP && !unicode.Is(ucd.LargeEastAsian, cr.r) {
		*breakOp = breakProhibited
	}
	// [CP-[\p{ea=F}\p{ea=W}\p{ea=H}]] × (AL | HL | NU)
	if cr.prevLine == ucd.BreakCP && !unicode.Is(ucd.LargeEastAsian, cr.prev) &&
		(cr.line == ucd.BreakAL || cr.line == ucd.BreakHL || cr.line == ucd.BreakNU) {
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB30ab(breakOp *breakOpportunity) {
	// (RI RI)* RI × RI
	if cr.isPrevLinebreakRIOdd && cr.line == ucd.BreakRI { // LB30a
		*breakOp = breakProhibited
	}

	// LB30b
	// EB × EM
	if cr.prevLine == ucd.BreakEB && cr.line == ucd.BreakEM {
		*breakOp = breakProhibited
	}
	// [\p{Extended_Pictographic}&\p{Cn}] × EM
	if cr.isPrevNonAssignedExtendedPic && cr.line == ucd.BreakEM {
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB29To26(breakOp *breakOpportunity) {
	bm1, b0, b1, b2 := cr.prevPrevLine, cr.prevLine, cr.line, cr.nextLine
	// LB29 : IS × (AL | HL)
	if b0 == ucd.BreakIS && (b1 == ucd.BreakAL || b1 == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	// LB28a Do not break inside the orthographic syllables of Brahmic scripts.
	// AP × (AK | [◌] | AS)
	if b0 == ucd.BreakAP && (b1 == ucd.BreakAK || cr.r == 0x25CC || b1 == ucd.BreakAS) ||
		// (AK | [◌] | AS) × (VF | VI)
		(b0 == ucd.BreakAK || cr.isPrevDottedCircle || b0 == ucd.BreakAS) && (b1 == ucd.BreakVF || b1 == ucd.BreakVI) ||
		// (AK | [◌] | AS) VI × (AK | [◌])
		(bm1 == ucd.BreakAK || cr.isPrevPrevDottedCircle || bm1 == ucd.BreakAS) && b0 == ucd.BreakVI && (b1 == ucd.BreakAK || cr.r == 0x25CC) ||
		// (AK | [◌] | AS) × (AK | [◌] | AS) VF
		(b0 == ucd.BreakAK || cr.isPrevDottedCircle || b0 == ucd.BreakAS) && (b1 == ucd.BreakAK || cr.r == 0x25CC || b1 == ucd.BreakAS) && b2 == ucd.BreakVF {
		*breakOp = breakProhibited
	}

	// LB28 : (AL | HL) × (AL | HL)
	if (b0 == ucd.BreakAL || b0 == ucd.BreakHL) && (b1 == ucd.BreakAL || b1 == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	// LB27
	// (JL | JV | JT | H2 | H3) × PO
	if (b0 == ucd.BreakJL || b0 == ucd.BreakJV || b0 == ucd.BreakJT || b0 == ucd.BreakH2 || b0 == ucd.BreakH3) &&
		b1 == ucd.BreakPO {
		*breakOp = breakProhibited
	}
	// PR × (JL | JV | JT | H2 | H3)
	if b0 == ucd.BreakPR &&
		(b1 == ucd.BreakJL || b1 == ucd.BreakJV || b1 == ucd.BreakJT || b1 == ucd.BreakH2 || b1 == ucd.BreakH3) {
		*breakOp = breakProhibited
	}
	// LB26
	// JL × (JL | JV | H2 | H3)
	if b0 == ucd.BreakJL &&
		(b1 == ucd.BreakJL || b1 == ucd.BreakJV || b1 == ucd.BreakH2 || b1 == ucd.BreakH3) {
		*breakOp = breakProhibited
	}
	// (JV | H2) × (JV | JT)
	if (b0 == ucd.BreakJV || b0 == ucd.BreakH2) && (b1 == ucd.BreakJV || b1 == ucd.BreakJT) {
		*breakOp = breakProhibited
	}
	// (JT | H3) × JT
	if (b0 == ucd.BreakJT || b0 == ucd.BreakH3) && b1 == ucd.BreakJT {
		*breakOp = breakProhibited
	}
}

// we follow other implementations by using the tailoring described
// in Example 7
func (cr *cursor) ruleLB25(breakOp *breakOpportunity, triggerNumSequence bool) {
	br0, br1 := cr.prevLine, cr.line
	// (PR | PO) × ( OP | HY )? NU
	if (br0 == ucd.BreakPR || br0 == ucd.BreakPO) && br1 == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	if (br0 == ucd.BreakPR || br0 == ucd.BreakPO) &&
		(br1 == ucd.BreakOP || br1 == ucd.BreakHY) &&
		cr.nextLine == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	// ( OP | HY | IS ) × NU
	if (br0 == ucd.BreakOP || br0 == ucd.BreakHY || br0 == ucd.BreakIS) && br1 == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	// NU × (NU | SY | IS)
	if br0 == ucd.BreakNU && (br1 == ucd.BreakNU || br1 == ucd.BreakSY || br1 == ucd.BreakIS) {
		*breakOp = breakProhibited
	}
	// NU (NU | SY | IS)* × (NU | SY | IS | CL | CP )
	if triggerNumSequence {
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB24To22(breakOp *breakOpportunity) {
	br0, br1 := cr.prevLine, cr.line
	// LB24
	// (PR | PO) × (AL | HL)
	if (br0 == ucd.BreakPR || br0 == ucd.BreakPO) && (br1 == ucd.BreakAL || br1 == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	// (AL | HL) × (PR | PO)
	if (br0 == ucd.BreakAL || br0 == ucd.BreakHL) && (br1 == ucd.BreakPR || br1 == ucd.BreakPO) {
		*breakOp = breakProhibited
	}
	// LB23
	// (AL | HL) × NU
	if (br0 == ucd.BreakAL || br0 == ucd.BreakHL) && br1 == ucd.BreakNU {
		*breakOp = breakProhibited
	}
	// NU × (AL | HL)
	if br0 == ucd.BreakNU && (br1 == ucd.BreakAL || br1 == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	// LB23a
	// PR × (ID | EB | EM)
	if br0 == ucd.BreakPR && (br1 == ucd.BreakID || br1 == ucd.BreakEB || br1 == ucd.BreakEM) {
		*breakOp = breakProhibited
	}
	// (ID | EB | EM) × PO
	if (br0 == ucd.BreakID || br0 == ucd.BreakEB || br0 == ucd.BreakEM) && br1 == ucd.BreakPO {
		*breakOp = breakProhibited
	}

	// LB22 : × IN
	if br1 == ucd.BreakIN {
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB21To8(breakOp *breakOpportunity) {
	brm1, br0, br1 := cr.prevPrevLine, cr.prevLine, cr.line
	// LB21
	// × BA
	// × HH
	// × HY
	// × NS
	// BB ×
	if br1 == ucd.BreakBA || br1 == ucd.BreakHH || br1 == ucd.BreakHY || br1 == ucd.BreakNS || br0 == ucd.BreakBB {
		*breakOp = breakProhibited
	}
	// LB21a : HL (HY | HH) × [^HL]
	if cr.prevPrevLine == ucd.BreakHL &&
		(br0 == ucd.BreakHY || br0 == ucd.BreakHH) && br1 != ucd.BreakHL {
		*breakOp = breakProhibited
	}
	// LB21b : SY × HL
	if br0 == ucd.BreakSY && br1 == ucd.BreakHL {
		*breakOp = breakProhibited
	}
	// LB20a Do not break after a word-initial hyphen.
	// ( sot | BK | CR | LF | NL | SP | ZW | CB | GL ) ( HY | HH ) × ( AL | HL )
	if (cr.isPreviousSot || brm1 == ucd.BreakBK || brm1 == ucd.BreakCR || brm1 == ucd.BreakLF || brm1 == ucd.BreakNL || brm1 == ucd.BreakSP || brm1 == ucd.BreakZW || brm1 == ucd.BreakCB || brm1 == ucd.BreakGL) &&
		(br0 == ucd.BreakHY || br0 == ucd.BreakHH) && (br1 == ucd.BreakAL || br1 == ucd.BreakHL) {
		*breakOp = breakProhibited
	}
	// LB20
	// ÷ CB
	// CB ÷
	if br0 == ucd.BreakCB || br1 == ucd.BreakCB {
		*breakOp = breakAllowed
	}
	// LB19
	// × [ QU - \p{Pi} ]
	// [ QU - \p{Pf} ] ×
	if (br1 == ucd.BreakQU && ucd.LookupType(cr.r) != ucd.Pi) || (br0 == ucd.BreakQU && ucd.LookupType(cr.prev) != ucd.Pf) {
		*breakOp = breakProhibited
	}
	// LB 19a
	// [^$EastAsian] × QU
	// × QU ( [^$EastAsian] | eot )
	// QU × [^$EastAsian]
	// ( sot | [^$EastAsian] ) QU ×
	if (br1 == ucd.BreakQU && !unicode.Is(ucd.LargeEastAsian, cr.prev)) ||
		(br1 == ucd.BreakQU && (cr.index == cr.len-1 || !unicode.Is(ucd.LargeEastAsian, cr.next))) ||
		(br0 == ucd.BreakQU && !unicode.Is(ucd.LargeEastAsian, cr.r)) ||
		((cr.isPreviousSot || !unicode.Is(ucd.LargeEastAsian, cr.prevPrev)) && br0 == ucd.BreakQU) {
		*breakOp = breakProhibited
	}

	// LB18 : SP ÷
	if br0 == ucd.BreakSP {
		*breakOp = breakAllowed
	}
	// LB17 : B2 SP* × B2
	spaceM1 := ucd.LookupLineBreakClass(cr.beforeSpaces)
	if spaceM1 == ucd.BreakB2 && br1 == ucd.BreakB2 {
		*breakOp = breakProhibited
	}
	// LB16 : (CL | CP) SP* × NS
	if (spaceM1 == ucd.BreakCL || spaceM1 == ucd.BreakCP) && br1 == ucd.BreakNS {
		*breakOp = breakProhibited
	}
	// LB15a Do not break after an unresolved initial punctuation that lies at the start of the line, after a space, after opening punctuation, or after an unresolved quotation mark, even after spaces.
	// (sot | BK | CR | LF | NL | OP | QU | GL | SP | ZW) [\p{Pi}&QU] SP* ×
	spaceM2 := ucd.LookupLineBreakClass(cr.prevBeforeSpaces)
	if (cr.beforeSpacesIndex == 0 || spaceM2 == ucd.BreakBK || spaceM2 == ucd.BreakCR || spaceM2 == ucd.BreakLF || spaceM2 == ucd.BreakNL || spaceM2 == ucd.BreakOP || spaceM2 == ucd.BreakQU || spaceM2 == ucd.BreakGL || spaceM2 == ucd.BreakSP || spaceM2 == ucd.BreakZW) &&
		(ucd.LookupType(cr.beforeSpaces) == ucd.Pi && spaceM1 == ucd.BreakQU) {
		*breakOp = breakProhibited
	}
	// LB15b Do not break before an unresolved final punctuation that lies at the end of the line, before a space, before a prohibited break, or before an unresolved quotation mark, even after spaces.
	// × [\p{Pf}&QU] ( SP | GL | WJ | CL | QU | CP | EX | IS | SY | BK | CR | LF | NL | ZW | eot)
	br2 := cr.nextLine
	if (ucd.LookupType(cr.r) == ucd.Pf && br1 == ucd.BreakQU) && (br2 == ucd.BreakSP || br2 == ucd.BreakGL || br2 == ucd.BreakWJ || br2 == ucd.BreakCL || br2 == ucd.BreakQU || br2 == ucd.BreakCP || br2 == ucd.BreakEX || br2 == ucd.BreakIS || br2 == ucd.BreakSY || br2 == ucd.BreakBK || br2 == ucd.BreakCR || br2 == ucd.BreakLF || br2 == ucd.BreakNL || br2 == ucd.BreakZW || cr.index == cr.len-1) {
		*breakOp = breakProhibited
	}
	if br0 == ucd.BreakSP && br1 == ucd.BreakIS && br2 == ucd.BreakNU {
		// LB15c Break before a decimal mark that follows a space, for instance, in ‘subtract .5’.
		// SP ÷ IS NU
		*breakOp = breakAllowed
	} else if br1 == ucd.BreakIS {
		// LB15d Otherwise, do not break before ‘;’, ‘,’, or ‘.’, even after spaces.
		// × IS
		*breakOp = breakProhibited
	}

	// LB14 : OP SP* ×
	if spaceM1 == ucd.BreakOP {
		*breakOp = breakProhibited
	}

	// rule LB13
	// × CL
	// × CP
	// × EX
	// × SY
	if br1 == ucd.BreakCL || br1 == ucd.BreakCP || br1 == ucd.BreakEX || br1 == ucd.BreakSY {
		*breakOp = breakProhibited
	}
	// LB12 : GL ×
	if br0 == ucd.BreakGL {
		*breakOp = breakProhibited
	}
	// LB12a : [^SP BA HY HH] × GL
	if (br0 != ucd.BreakSP && br0 != ucd.BreakBA && br0 != ucd.BreakHY && br0 != ucd.BreakHH) &&
		br1 == ucd.BreakGL {
		*breakOp = breakProhibited
	}
	// LB11
	// × WJ
	// WJ ×
	if br0 == ucd.BreakWJ || br1 == ucd.BreakWJ {
		*breakOp = breakProhibited
	}

	// rule LB9 : "Do not break a combining character sequence"
	// where X is any line break class except BK, CR, LF, NL, SP, or ZW.
	// see also [endIteration]
	if br1 == ucd.BreakCM || br1 == ucd.BreakZWJ {
		if !(br0 == ucd.BreakBK || br0 == ucd.BreakCR || br0 == ucd.BreakLF ||
			br0 == ucd.BreakNL || br0 == ucd.BreakSP || br0 == ucd.BreakZW) {
			*breakOp = breakProhibited
		}
	}

	// there is a catch here : prevLine or beforeSpace are not always
	// computed at index i-1, because of rules LB9 and LB10
	// however, rule LB8 and LB8a applies before LB9 and LB10, meaning
	// we need to use the real class

	if unicode.Is(ucd.BreakZW, cr.beforeSpaceRaw) { // rule LB8 : ZW SP* ÷
		*breakOp = breakAllowed
	} else if unicode.Is(ucd.BreakZWJ, cr.prev) { // rule LB8a : ZWJ ×
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB7To4(breakOp *breakOpportunity) {
	// LB7
	// × SP
	// × ZW
	if cr.line == ucd.BreakSP || cr.line == ucd.BreakZW {
		*breakOp = breakProhibited
	}
	// LB6 : × ( BK | CR | LF | NL )
	if cr.line == ucd.BreakBK || cr.line == ucd.BreakCR || cr.line == ucd.BreakLF || cr.line == ucd.BreakNL {
		*breakOp = breakProhibited
	}

	// LB4 and LB5
	// BK !
	// CR !
	// LF !
	// NL !
	// (CR × LF is actually handled in rule LB6)
	if cr.prevLine == ucd.BreakBK || (cr.prevLine == ucd.BreakCR && cr.r != '\n') ||
		cr.prevLine == ucd.BreakLF || cr.prevLine == ucd.BreakNL {
		*breakOp = breakMandatory
	}
}

// apply rule LB1 to resolve break classses AI, SG, XX, SA and CJ.
// We use the default values specified in https://unicode.org/reports/tr14/#BreakingRules.
func (cr *cursor) ruleLB1() {
	switch cr.line {
	case ucd.BreakAI, ucd.BreakSG, ucd.BreakXX:
		cr.line = ucd.BreakAL
	case ucd.BreakSA:
		if unicode.Is(ucd.Mn, cr.r) || unicode.Is(ucd.Mc, cr.r) {
			cr.line = ucd.BreakCM
		} else {
			cr.line = ucd.BreakAL
		}
	case ucd.BreakCJ:
		cr.line = ucd.BreakNS
	}
}

type numSequenceState uint8

const (
	noNumSequence numSequenceState = iota // we are not in a sequence
	inNumSequence                         // we are in NU (NU | SY | IS)*
	seenCloseNum                          // we are at NU (NU | SY | IS)* (CL | CP)?
)

// update the `numSequence` state used for rule LB25
// and returns true if we matched one
func (cr *cursor) updateNumSequence() bool {
	// note that rule LB9 also apply : (CM|ZWJ) do not change
	// the flag
	if cr.line == ucd.BreakCM || cr.line == ucd.BreakZWJ {
		return false
	}

	switch cr.numSequence {
	case noNumSequence:
		if cr.line == ucd.BreakNU { // start a sequence
			cr.numSequence = inNumSequence
		}
		return false
	case inNumSequence:
		switch cr.line {
		case ucd.BreakNU, ucd.BreakSY, ucd.BreakIS:
			// NU (NU | SY | IS)* × (NU | SY | IS) : the sequence continue
			return true
		case ucd.BreakCL, ucd.BreakCP:
			// NU (NU | SY | IS)* × (CL | CP)
			cr.numSequence = seenCloseNum
			return true
		case ucd.BreakPO, ucd.BreakPR:
			// NU (NU | SY | IS)* × (PO | PR) : close the sequence
			cr.numSequence = noNumSequence
			return true
		default:
			cr.numSequence = noNumSequence
			return false
		}
	case seenCloseNum:
		cr.numSequence = noNumSequence // close the sequence anyway
		if cr.line == ucd.BreakPO || cr.line == ucd.BreakPR {
			// NU (NU | SY | IS)* (CL | CP) × (PO | PR)
			return true
		}
		return false
	default:
		panic("exhaustive switch")
	}
}

// startIteration updates the cursor properties, setting the current
// rune to text[i].
// Some properties depending on the context are rather
// updated in the previous `endIteration` call.
func (cr *cursor) startIteration(text []rune, i int) {
	cr.index = i

	cr.prevPrev = cr.prev
	cr.prev = cr.r
	if i < len(text) {
		cr.r = text[i]
	} else {
		cr.r = paragraphSeparator
	}
	if i == len(text) {
		cr.next = 0
	} else if i == len(text)-1 {
		// we fill in the last element of `attrs` by assuming
		// there's a paragraph separators off the end of text
		cr.next = paragraphSeparator
	} else {
		cr.next = text[i+1]
	}

	// query general unicode properties for the current rune

	cr.isExtentedPic = unicode.Is(ucd.Extended_Pictographic, cr.r)
	cr.indicConjunctBreak = ucd.LookupIndicConjunctBreak(cr.r)

	cr.prevGrapheme = cr.grapheme
	cr.grapheme = ucd.LookupGraphemeBreakClass(cr.r)

	if cr.word != ucd.WordBreakExtendFormat {
		cr.prevPrevWord = cr.prevWord
		cr.prevWord = cr.word
		cr.prevWordNoExtend = i - 1
	}
	cr.word = ucd.LookupWordBreakClass(cr.r)

	// prevPrevLine and prevLine are handled in endIteration
	cr.line = cr.nextLine // avoid calling LookupLineBreakClass twice
	cr.nextLine = ucd.LookupLineBreakClass(cr.next)
}

// end the current iteration, computing some of the properties
// required for the next rune and respecting rule LB9 and LB10
func (cr *cursor) endIteration() {
	// start by handling rule LB9 and LB10
	if cr.line == ucd.BreakCM || cr.line == ucd.BreakZWJ {
		isLB10 := cr.prevLine == ucd.BreakBK ||
			cr.prevLine == ucd.BreakCR ||
			cr.prevLine == ucd.BreakLF ||
			cr.prevLine == ucd.BreakNL ||
			cr.prevLine == ucd.BreakSP ||
			cr.prevLine == ucd.BreakZW
		if cr.index == 0 || isLB10 { // Rule LB10
			cr.prevLine = ucd.BreakAL
		} // else rule LB9 : ignore the rune for prevLine and prevPrevLine

	} else { // regular update
		cr.isPreviousSot = cr.index == 0

		cr.prevPrevLine = cr.prevLine
		cr.prevLine = cr.line

		cr.isPrevPrevDottedCircle = cr.isPrevDottedCircle
		cr.isPrevDottedCircle = cr.r == 0x25CC

		cr.isPrevNonAssignedExtendedPic = cr.isExtentedPic && ucd.LookupType(cr.r) == nil

		// keep track of the rune before the spaces
		if cr.prevLine != ucd.BreakSP {
			cr.beforeSpaces = cr.r
			cr.prevBeforeSpaces = cr.prev
			cr.beforeSpacesIndex = cr.index
		}
	}

	// keep track of the rune before the spaces
	if cr.prevLine != ucd.BreakSP {
		cr.beforeSpaceRaw = cr.r
	}

	// update RegionalIndicator parity used for LB30a
	if cr.line == ucd.BreakRI {
		cr.isPrevLinebreakRIOdd = !cr.isPrevLinebreakRIOdd
	} else if !(cr.line == ucd.BreakCM || cr.line == ucd.BreakZWJ) { // beware of the rule LB9: (CM|ZWJ) ignore the update
		cr.isPrevLinebreakRIOdd = false
	}
}
