// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package segmenter

import (
	ucd "github.com/go-text/typesetting/internal/unicodedata"
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
	if (cr.prevLine == ucd.LB_AL || cr.prevLine == ucd.LB_HL || cr.prevLine == ucd.LB_NU) &&
		cr.line == ucd.LB_OP && !ucd.IsLargeEastAsian(cr.r) {
		*breakOp = breakProhibited
	}
	// [CP-[\p{ea=F}\p{ea=W}\p{ea=H}]] × (AL | HL | NU)
	if cr.prevLine == ucd.LB_CP && !ucd.IsLargeEastAsian(cr.prev) &&
		(cr.line == ucd.LB_AL || cr.line == ucd.LB_HL || cr.line == ucd.LB_NU) {
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB30ab(breakOp *breakOpportunity) {
	// (RI RI)* RI × RI
	if cr.isPrevLinebreakRIOdd && cr.line == ucd.LB_RI { // LB30a
		*breakOp = breakProhibited
	}

	// LB30b
	// EB × EM
	if cr.prevLine == ucd.LB_EB && cr.line == ucd.LB_EM {
		*breakOp = breakProhibited
	}
	// [\p{Extended_Pictographic}&\p{Cn}] × EM
	if cr.isPrevNonAssignedExtendedPic && cr.line == ucd.LB_EM {
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB29To26(breakOp *breakOpportunity) {
	bm1, b0, b1, b2 := cr.prevPrevLine, cr.prevLine, cr.line, cr.nextLine
	// LB29 : IS × (AL | HL)
	if b0 == ucd.LB_IS && (b1 == ucd.LB_AL || b1 == ucd.LB_HL) {
		*breakOp = breakProhibited
	}
	// LB28a Do not break inside the orthographic syllables of Brahmic scripts.
	// AP × (AK | [◌] | AS)
	if b0 == ucd.LB_AP && (b1 == ucd.LB_AK || cr.r == 0x25CC || b1 == ucd.LB_AS) ||
		// (AK | [◌] | AS) × (VF | VI)
		(b0 == ucd.LB_AK || cr.isPrevDottedCircle || b0 == ucd.LB_AS) && (b1 == ucd.LB_VF || b1 == ucd.LB_VI) ||
		// (AK | [◌] | AS) VI × (AK | [◌])
		(bm1 == ucd.LB_AK || cr.isPrevPrevDottedCircle || bm1 == ucd.LB_AS) && b0 == ucd.LB_VI && (b1 == ucd.LB_AK || cr.r == 0x25CC) ||
		// (AK | [◌] | AS) × (AK | [◌] | AS) VF
		(b0 == ucd.LB_AK || cr.isPrevDottedCircle || b0 == ucd.LB_AS) && (b1 == ucd.LB_AK || cr.r == 0x25CC || b1 == ucd.LB_AS) && b2 == ucd.LB_VF {
		*breakOp = breakProhibited
	}

	// LB28 : (AL | HL) × (AL | HL)
	if (b0 == ucd.LB_AL || b0 == ucd.LB_HL) && (b1 == ucd.LB_AL || b1 == ucd.LB_HL) {
		*breakOp = breakProhibited
	}
	// LB27
	// (JL | JV | JT | H2 | H3) × PO
	if (b0 == ucd.LB_JL || b0 == ucd.LB_JV || b0 == ucd.LB_JT || b0 == ucd.LB_H2 || b0 == ucd.LB_H3) &&
		b1 == ucd.LB_PO {
		*breakOp = breakProhibited
	}
	// PR × (JL | JV | JT | H2 | H3)
	if b0 == ucd.LB_PR &&
		(b1 == ucd.LB_JL || b1 == ucd.LB_JV || b1 == ucd.LB_JT || b1 == ucd.LB_H2 || b1 == ucd.LB_H3) {
		*breakOp = breakProhibited
	}
	// LB26
	// JL × (JL | JV | H2 | H3)
	if b0 == ucd.LB_JL &&
		(b1 == ucd.LB_JL || b1 == ucd.LB_JV || b1 == ucd.LB_H2 || b1 == ucd.LB_H3) {
		*breakOp = breakProhibited
	}
	// (JV | H2) × (JV | JT)
	if (b0 == ucd.LB_JV || b0 == ucd.LB_H2) && (b1 == ucd.LB_JV || b1 == ucd.LB_JT) {
		*breakOp = breakProhibited
	}
	// (JT | H3) × JT
	if (b0 == ucd.LB_JT || b0 == ucd.LB_H3) && b1 == ucd.LB_JT {
		*breakOp = breakProhibited
	}
}

// we follow other implementations by using the tailoring described
// in Example 7
func (cr *cursor) ruleLB25(breakOp *breakOpportunity, triggerNumSequence bool) {
	br0, br1 := cr.prevLine, cr.line
	// (PR | PO) × ( OP | HY )? NU
	if (br0 == ucd.LB_PR || br0 == ucd.LB_PO) && br1 == ucd.LB_NU {
		*breakOp = breakProhibited
	}
	if (br0 == ucd.LB_PR || br0 == ucd.LB_PO) &&
		(br1 == ucd.LB_OP || br1 == ucd.LB_HY) &&
		cr.nextLine == ucd.LB_NU {
		*breakOp = breakProhibited
	}
	// ( OP | HY | IS ) × NU
	if (br0 == ucd.LB_OP || br0 == ucd.LB_HY || br0 == ucd.LB_IS) && br1 == ucd.LB_NU {
		*breakOp = breakProhibited
	}
	// NU × (NU | SY | IS)
	if br0 == ucd.LB_NU && (br1 == ucd.LB_NU || br1 == ucd.LB_SY || br1 == ucd.LB_IS) {
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
	if (br0 == ucd.LB_PR || br0 == ucd.LB_PO) && (br1 == ucd.LB_AL || br1 == ucd.LB_HL) {
		*breakOp = breakProhibited
	}
	// (AL | HL) × (PR | PO)
	if (br0 == ucd.LB_AL || br0 == ucd.LB_HL) && (br1 == ucd.LB_PR || br1 == ucd.LB_PO) {
		*breakOp = breakProhibited
	}
	// LB23
	// (AL | HL) × NU
	if (br0 == ucd.LB_AL || br0 == ucd.LB_HL) && br1 == ucd.LB_NU {
		*breakOp = breakProhibited
	}
	// NU × (AL | HL)
	if br0 == ucd.LB_NU && (br1 == ucd.LB_AL || br1 == ucd.LB_HL) {
		*breakOp = breakProhibited
	}
	// LB23a
	// PR × (ID | EB | EM)
	if br0 == ucd.LB_PR && (br1 == ucd.LB_ID || br1 == ucd.LB_EB || br1 == ucd.LB_EM) {
		*breakOp = breakProhibited
	}
	// (ID | EB | EM) × PO
	if (br0 == ucd.LB_ID || br0 == ucd.LB_EB || br0 == ucd.LB_EM) && br1 == ucd.LB_PO {
		*breakOp = breakProhibited
	}

	// LB22 : × IN
	if br1 == ucd.LB_IN {
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
	if br1 == ucd.LB_BA || br1 == ucd.LB_HH || br1 == ucd.LB_HY || br1 == ucd.LB_NS || br0 == ucd.LB_BB {
		*breakOp = breakProhibited
	}
	// LB21a : HL (HY | HH) × [^HL]
	if cr.prevPrevLine == ucd.LB_HL &&
		(br0 == ucd.LB_HY || br0 == ucd.LB_HH) && br1 != ucd.LB_HL {
		*breakOp = breakProhibited
	}
	// LB21b : SY × HL
	if br0 == ucd.LB_SY && br1 == ucd.LB_HL {
		*breakOp = breakProhibited
	}
	// LB20a Do not break after a word-initial hyphen.
	// ( sot | BK | CR | LF | NL | SP | ZW | CB | GL ) ( HY | HH ) × ( AL | HL )
	if (cr.isPreviousSot || brm1 == ucd.LB_BK || brm1 == ucd.LB_CR || brm1 == ucd.LB_LF || brm1 == ucd.LB_NL || brm1 == ucd.LB_SP || brm1 == ucd.LB_ZW || brm1 == ucd.LB_CB || brm1 == ucd.LB_GL) &&
		(br0 == ucd.LB_HY || br0 == ucd.LB_HH) && (br1 == ucd.LB_AL || br1 == ucd.LB_HL) {
		*breakOp = breakProhibited
	}
	// LB20
	// ÷ CB
	// CB ÷
	if br0 == ucd.LB_CB || br1 == ucd.LB_CB {
		*breakOp = breakAllowed
	}
	// LB19
	// × [ QU - \p{Pi} ]
	// [ QU - \p{Pf} ] ×
	if (br1 == ucd.LB_QU && ucd.LookupGeneralCategory(cr.r) != ucd.Pi) || (br0 == ucd.LB_QU && ucd.LookupGeneralCategory(cr.prev) != ucd.Pf) {
		*breakOp = breakProhibited
	}
	// LB 19a
	// [^$EastAsian] × QU
	// × QU ( [^$EastAsian] | eot )
	// QU × [^$EastAsian]
	// ( sot | [^$EastAsian] ) QU ×
	if (br1 == ucd.LB_QU && !ucd.IsLargeEastAsian(cr.prev)) ||
		(br1 == ucd.LB_QU && (cr.index == cr.len-1 || !ucd.IsLargeEastAsian(cr.next))) ||
		(br0 == ucd.LB_QU && !ucd.IsLargeEastAsian(cr.r)) ||
		((cr.isPreviousSot || !ucd.IsLargeEastAsian(cr.prevPrev)) && br0 == ucd.LB_QU) {
		*breakOp = breakProhibited
	}

	// LB18 : SP ÷
	if br0 == ucd.LB_SP {
		*breakOp = breakAllowed
	}
	// LB17 : B2 SP* × B2
	spaceM1 := ucd.LookupLineBreak(cr.beforeSpaces)
	if spaceM1 == ucd.LB_B2 && br1 == ucd.LB_B2 {
		*breakOp = breakProhibited
	}
	// LB16 : (CL | CP) SP* × NS
	if (spaceM1 == ucd.LB_CL || spaceM1 == ucd.LB_CP) && br1 == ucd.LB_NS {
		*breakOp = breakProhibited
	}
	// LB15a Do not break after an unresolved initial punctuation that lies at the start of the line, after a space, after opening punctuation, or after an unresolved quotation mark, even after spaces.
	// (sot | BK | CR | LF | NL | OP | QU | GL | SP | ZW) [\p{Pi}&QU] SP* ×
	spaceM2 := ucd.LookupLineBreak(cr.prevBeforeSpaces)
	if (cr.beforeSpacesIndex == 0 || spaceM2 == ucd.LB_BK || spaceM2 == ucd.LB_CR || spaceM2 == ucd.LB_LF || spaceM2 == ucd.LB_NL || spaceM2 == ucd.LB_OP || spaceM2 == ucd.LB_QU || spaceM2 == ucd.LB_GL || spaceM2 == ucd.LB_SP || spaceM2 == ucd.LB_ZW) &&
		(ucd.LookupGeneralCategory(cr.beforeSpaces) == ucd.Pi && spaceM1 == ucd.LB_QU) {
		*breakOp = breakProhibited
	}
	// LB15b Do not break before an unresolved final punctuation that lies at the end of the line, before a space, before a prohibited break, or before an unresolved quotation mark, even after spaces.
	// × [\p{Pf}&QU] ( SP | GL | WJ | CL | QU | CP | EX | IS | SY | BK | CR | LF | NL | ZW | eot)
	br2 := cr.nextLine
	if (ucd.LookupGeneralCategory(cr.r) == ucd.Pf && br1 == ucd.LB_QU) && (br2 == ucd.LB_SP || br2 == ucd.LB_GL || br2 == ucd.LB_WJ || br2 == ucd.LB_CL || br2 == ucd.LB_QU || br2 == ucd.LB_CP || br2 == ucd.LB_EX || br2 == ucd.LB_IS || br2 == ucd.LB_SY || br2 == ucd.LB_BK || br2 == ucd.LB_CR || br2 == ucd.LB_LF || br2 == ucd.LB_NL || br2 == ucd.LB_ZW || cr.index == cr.len-1) {
		*breakOp = breakProhibited
	}
	if br0 == ucd.LB_SP && br1 == ucd.LB_IS && br2 == ucd.LB_NU {
		// LB15c Break before a decimal mark that follows a space, for instance, in ‘subtract .5’.
		// SP ÷ IS NU
		*breakOp = breakAllowed
	} else if br1 == ucd.LB_IS {
		// LB15d Otherwise, do not break before ‘;’, ‘,’, or ‘.’, even after spaces.
		// × IS
		*breakOp = breakProhibited
	}

	// LB14 : OP SP* ×
	if spaceM1 == ucd.LB_OP {
		*breakOp = breakProhibited
	}

	// rule LB13
	// × CL
	// × CP
	// × EX
	// × SY
	if br1 == ucd.LB_CL || br1 == ucd.LB_CP || br1 == ucd.LB_EX || br1 == ucd.LB_SY {
		*breakOp = breakProhibited
	}
	// LB12 : GL ×
	if br0 == ucd.LB_GL {
		*breakOp = breakProhibited
	}
	// LB12a : [^SP BA HY HH] × GL
	if (br0 != ucd.LB_SP && br0 != ucd.LB_BA && br0 != ucd.LB_HY && br0 != ucd.LB_HH) &&
		br1 == ucd.LB_GL {
		*breakOp = breakProhibited
	}
	// LB11
	// × WJ
	// WJ ×
	if br0 == ucd.LB_WJ || br1 == ucd.LB_WJ {
		*breakOp = breakProhibited
	}

	// rule LB9 : "Do not break a combining character sequence"
	// where X is any line break class except BK, CR, LF, NL, SP, or ZW.
	// see also [endIteration]
	if br1 == ucd.LB_CM || br1 == ucd.LB_ZWJ {
		if !(br0 == ucd.LB_BK || br0 == ucd.LB_CR || br0 == ucd.LB_LF ||
			br0 == ucd.LB_NL || br0 == ucd.LB_SP || br0 == ucd.LB_ZW) {
			*breakOp = breakProhibited
		}
	}

	// there is a catch here : prevLine or beforeSpace are not always
	// computed at index i-1, because of rules LB9 and LB10
	// however, rule LB8 and LB8a applies before LB9 and LB10, meaning
	// we need to use the real class

	if ucd.LookupLineBreak(cr.beforeSpaceRaw) == ucd.LB_ZW { // rule LB8 : ZW SP* ÷
		*breakOp = breakAllowed
	} else if ucd.LookupLineBreak(cr.prev) == ucd.LB_ZWJ { // rule LB8a : ZWJ ×
		*breakOp = breakProhibited
	}
}

func (cr *cursor) ruleLB7To4(breakOp *breakOpportunity) {
	// LB7
	// × SP
	// × ZW
	if cr.line == ucd.LB_SP || cr.line == ucd.LB_ZW {
		*breakOp = breakProhibited
	}
	// LB6 : × ( BK | CR | LF | NL )
	if cr.line == ucd.LB_BK || cr.line == ucd.LB_CR || cr.line == ucd.LB_LF || cr.line == ucd.LB_NL {
		*breakOp = breakProhibited
	}

	// LB4 and LB5
	// BK !
	// CR !
	// LF !
	// NL !
	// (CR × LF is actually handled in rule LB6)
	if cr.prevLine == ucd.LB_BK || (cr.prevLine == ucd.LB_CR && cr.r != '\n') ||
		cr.prevLine == ucd.LB_LF || cr.prevLine == ucd.LB_NL {
		*breakOp = breakMandatory
	}
}

// apply rule LB1 to resolve break classses AI, SG, XX, SA and CJ.
// We use the default values specified in https://unicode.org/reports/tr14/#BreakingRules.
func (cr *cursor) ruleLB1() {
	switch cr.line {
	case ucd.LB_AI, ucd.LB_SG, 0:
		cr.line = ucd.LB_AL
	case ucd.LB_SA:
		if cat := ucd.LookupGeneralCategory(cr.r); cat == ucd.Mn || cat == ucd.Mc {
			cr.line = ucd.LB_CM
		} else {
			cr.line = ucd.LB_AL
		}
	case ucd.LB_CJ:
		cr.line = ucd.LB_NS
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
	if cr.line == ucd.LB_CM || cr.line == ucd.LB_ZWJ {
		return false
	}

	switch cr.numSequence {
	case noNumSequence:
		if cr.line == ucd.LB_NU { // start a sequence
			cr.numSequence = inNumSequence
		}
		return false
	case inNumSequence:
		switch cr.line {
		case ucd.LB_NU, ucd.LB_SY, ucd.LB_IS:
			// NU (NU | SY | IS)* × (NU | SY | IS) : the sequence continue
			return true
		case ucd.LB_CL, ucd.LB_CP:
			// NU (NU | SY | IS)* × (CL | CP)
			cr.numSequence = seenCloseNum
			return true
		case ucd.LB_PO, ucd.LB_PR:
			// NU (NU | SY | IS)* × (PO | PR) : close the sequence
			cr.numSequence = noNumSequence
			return true
		default:
			cr.numSequence = noNumSequence
			return false
		}
	case seenCloseNum:
		cr.numSequence = noNumSequence // close the sequence anyway
		if cr.line == ucd.LB_PO || cr.line == ucd.LB_PR {
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

	cr.isExtentedPic = ucd.IsExtendedPictographic(cr.r)
	cr.indicConjunctBreak = ucd.LookupIndicConjunctBreak(cr.r)

	cr.prevGrapheme = cr.grapheme
	cr.grapheme = ucd.LookupGraphemeBreak(cr.r)

	if cr.word != ucd.WB_ExtendFormat {
		cr.prevPrevWord = cr.prevWord
		cr.prevWord = cr.word
		cr.prevWordNoExtend = i - 1
	}
	cr.word = ucd.LookupWordBreak(cr.r)

	// prevPrevLine and prevLine are handled in endIteration
	cr.line = cr.nextLine // avoid calling LookupLineBreakClass twice
	cr.nextLine = ucd.LookupLineBreak(cr.next)
}

// end the current iteration, computing some of the properties
// required for the next rune and respecting rule LB9 and LB10
func (cr *cursor) endIteration() {
	// start by handling rule LB9 and LB10
	if cr.line == ucd.LB_CM || cr.line == ucd.LB_ZWJ {
		isLB10 := cr.prevLine == ucd.LB_BK ||
			cr.prevLine == ucd.LB_CR ||
			cr.prevLine == ucd.LB_LF ||
			cr.prevLine == ucd.LB_NL ||
			cr.prevLine == ucd.LB_SP ||
			cr.prevLine == ucd.LB_ZW
		if cr.index == 0 || isLB10 { // Rule LB10
			cr.prevLine = ucd.LB_AL
		} // else rule LB9 : ignore the rune for prevLine and prevPrevLine

	} else { // regular update
		cr.isPreviousSot = cr.index == 0

		cr.prevPrevLine = cr.prevLine
		cr.prevLine = cr.line

		cr.isPrevPrevDottedCircle = cr.isPrevDottedCircle
		cr.isPrevDottedCircle = cr.r == 0x25CC

		cr.isPrevNonAssignedExtendedPic = cr.isExtentedPic && ucd.LookupGeneralCategory(cr.r) == ucd.Unassigned

		// keep track of the rune before the spaces
		if cr.prevLine != ucd.LB_SP {
			cr.beforeSpaces = cr.r
			cr.prevBeforeSpaces = cr.prev
			cr.beforeSpacesIndex = cr.index
		}
	}

	// keep track of the rune before the spaces
	if cr.prevLine != ucd.LB_SP {
		cr.beforeSpaceRaw = cr.r
	}

	// update RegionalIndicator parity used for LB30a
	if cr.line == ucd.LB_RI {
		cr.isPrevLinebreakRIOdd = !cr.isPrevLinebreakRIOdd
	} else if !(cr.line == ucd.LB_CM || cr.line == ucd.LB_ZWJ) { // beware of the rule LB9: (CM|ZWJ) ignore the update
		cr.isPrevLinebreakRIOdd = false
	}
}
