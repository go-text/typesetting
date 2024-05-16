package shaping

import (
	"testing"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
	"golang.org/x/image/math/fixed"
)

func simpleShape(text []rune, face *font.Face, dir di.Direction) Output {
	input := Input{
		Text:      text,
		RunStart:  0,
		RunEnd:    len(text),
		Direction: dir,
		Face:      face,
		Size:      16 * 72 * 10,
		Script:    language.LookupScript(text[0]),
	}
	return (&HarfbuzzShaper{}).Shape(input)
}

func TestOutput_addWordSpacing(t *testing.T) {
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	english := []rune("\U0001039FHello\u1361world ! : the\u00A0end")
	arabic := []rune("تثذرزسشص لمنهويء")

	addSpacing := fixed.I(20)
	out := simpleShape(english, latinFont, di.DirectionLTR)
	withoutSpacing := out.Advance
	out.AddWordSpacing(english, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+6*addSpacing)

	out = simpleShape(arabic, arabicFont, di.DirectionRTL)
	withoutSpacing = out.Advance
	out.AddWordSpacing(arabic, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+1*addSpacing)

	// vertical
	out = simpleShape(english, latinFont, di.DirectionTTB)
	withoutSpacing = out.Advance
	out.AddWordSpacing(english, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+6*addSpacing)
}

func TestOutput_addLetterSpacing(t *testing.T) {
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	english := []rune("Hello world ! : the end")
	englishWithLigature := []rune("Hello final")
	arabic := []rune("تثذرزسشص لمنهويء")

	addSpacing := fixed.I(4)
	halfSpacing := addSpacing / 2
	for _, test := range []struct {
		text                 []rune
		face                 *font.Face
		dir                  di.Direction
		start, end           bool
		expectedBonusAdvance fixed.Int26_6
	}{
		// LTR
		{english, latinFont, di.DirectionLTR, false, false, 23 * addSpacing},
		{english, latinFont, di.DirectionLTR, true, true, 22 * addSpacing},
		{english, latinFont, di.DirectionLTR, true, false, 22*addSpacing + halfSpacing},
		{english, latinFont, di.DirectionLTR, false, true, 22*addSpacing + halfSpacing},
		{englishWithLigature, latinFont, di.DirectionLTR, true, true, 9 * addSpacing}, // not 10
		// RTL
		{arabic, arabicFont, di.DirectionRTL, false, false, 16 * addSpacing},
		{arabic, arabicFont, di.DirectionRTL, true, true, 15 * addSpacing},
		{arabic, arabicFont, di.DirectionRTL, true, false, 15*addSpacing + halfSpacing},
		{arabic, arabicFont, di.DirectionRTL, false, true, 15*addSpacing + halfSpacing},
		// vertical
		{english, latinFont, di.DirectionTTB, false, false, 23 * addSpacing},
		{english, latinFont, di.DirectionTTB, true, true, 22 * addSpacing},
		{english, latinFont, di.DirectionTTB, true, false, 22*addSpacing + halfSpacing},
		{english, latinFont, di.DirectionTTB, false, true, 22*addSpacing + halfSpacing},
	} {
		out := simpleShape(test.text, test.face, test.dir)
		withoutSpacing := out.Advance
		out.AddLetterSpacing(addSpacing, test.start, test.end)
		tu.Assert(t, out.Advance == withoutSpacing+test.expectedBonusAdvance)
	}
}

func TestCustomSpacing(t *testing.T) {
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	english := []rune("Hello world ! : the end")

	letterSpacing, wordSpacing := fixed.I(4), fixed.I(20)
	out := simpleShape(english, latinFont, di.DirectionLTR)
	withoutSpacing := out.Advance
	out.AddWordSpacing(english, wordSpacing)
	out.AddLetterSpacing(letterSpacing, false, false)
	tu.Assert(t, out.Advance == withoutSpacing+5*wordSpacing+23*letterSpacing)
}

// make sure that additional letter spacing is properly removed
// at the start and end of wrapped lines
func TestTrailingSpaces(t *testing.T) {
	letterSpacing, charAdvance := fixed.I(8), fixed.I(90)
	halfSpacing := letterSpacing / 2
	monoFont := loadOpentypeFont(t, "../font/testdata/UbuntuMono-R.ttf")

	text := []rune("Hello world ! : the end_")

	out := simpleShape(text, monoFont, di.DirectionLTR)
	tu.Assert(t, out.Advance == fixed.Int26_6(len(text))*charAdvance) // assume 1:1 rune glyph mapping

	type test struct {
		toWrap       []Output
		policy       LineBreakPolicy
		width        int
		expectedRuns [][][2]fixed.Int26_6 // first and last advance, for each line and each run
	}

	for _, test := range []test{
		{ // from one run
			[]Output{out.copy()},
			0, 1800,
			[][][2]fixed.Int26_6{
				{{charAdvance + halfSpacing, 0}},                         // line 1
				{{charAdvance + halfSpacing, charAdvance + halfSpacing}}, // line 2
			},
		},
		{ // from two runs, break between
			cutRunInto(out.copy(), 2), Always, 1172, // end of the first run
			[][][2]fixed.Int26_6{
				{{charAdvance + halfSpacing, 0}},                         // line 1
				{{charAdvance + halfSpacing, charAdvance + halfSpacing}}, // line 2
			},
		},
		{ // from two runs, break inside
			cutRunInto(out.copy(), 2), 0, 1800,
			[][][2]fixed.Int26_6{
				{{charAdvance + halfSpacing, charAdvance + letterSpacing}, {charAdvance + letterSpacing, 0}}, // line 1
				{{charAdvance + halfSpacing, charAdvance + halfSpacing}},                                     // line 2
			},
		},
	} {
		AddSpacing(test.toWrap, text, 0, letterSpacing)
		lines, _ := (&LineWrapper{}).WrapParagraph(WrapConfig{BreakPolicy: test.policy}, test.width, text, NewSliceIterator(test.toWrap))
		tu.Assert(t, len(lines) == len(test.expectedRuns))
		for i, expLine := range test.expectedRuns {
			gotLine := lines[i]
			tu.Assert(t, len(gotLine) == len(expLine))
			for j, run := range expLine {
				gotRun := gotLine[j]
				tu.Assert(t, gotRun.Glyphs[0].XAdvance == run[0])
				tu.Assert(t, gotRun.Glyphs[len(gotRun.Glyphs)-1].XAdvance == run[1])
			}
		}
	}
}
