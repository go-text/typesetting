package shaping

import (
	"testing"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/opentype/testutils"
	"golang.org/x/image/math/fixed"
)

func simpleShape(text []rune, face font.Face, dir di.Direction) Output {
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
	english := []rune("Hello world ! : the end")
	arabic := []rune("تثذرزسشص لمنهويء")

	addSpacing := fixed.I(20)
	out := simpleShape(english, latinFont, di.DirectionLTR)
	withoutSpacing := out.Advance
	out.AddWordSpacing(english, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+5*addSpacing)

	out = simpleShape(arabic, arabicFont, di.DirectionRTL)
	withoutSpacing = out.Advance
	out.AddWordSpacing(arabic, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+1*addSpacing)

	// vertical
	out = simpleShape(english, latinFont, di.DirectionTTB)
	withoutSpacing = out.Advance
	out.AddWordSpacing(english, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+5*addSpacing)
}

func TestOutput_addLetterSpacing(t *testing.T) {
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	english := []rune("Hello world ! : the end")
	arabic := []rune("تثذرزسشص لمنهويء")

	addSpacing := fixed.I(4)
	for _, test := range []struct {
		text                 []rune
		face                 font.Face
		dir                  di.Direction
		expectedBonusAdvance fixed.Int26_6
	}{
		{english, latinFont, di.DirectionLTR, 23 * addSpacing},
		{arabic, arabicFont, di.DirectionRTL, 16 * addSpacing},
		// vertical
		{english, latinFont, di.DirectionTTB, 23 * addSpacing},
	} {
		out := simpleShape(test.text, test.face, test.dir)
		withoutSpacing := out.Advance
		out.AddLetterSpacing(addSpacing)
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
	out.AddLetterSpacing(letterSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+5*wordSpacing+23*letterSpacing)
}

func TestTrimSpace(t *testing.T) {
	letterSpacing := fixed.I(4)
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")

	english := []rune("Hello world ! : the end")

	l := Line{simpleShape(english, latinFont, di.DirectionLTR)}
	run := &l[0]
	advance := run.Advance

	// no-op
	l.TrimLetterSpacing(0)
	tu.Assert(t, advance == run.Advance)

	run.AddLetterSpacing(letterSpacing)
	beforeTrim := run.Advance
	l.TrimLetterSpacing(letterSpacing)
	tu.Assert(t, run.Advance == beforeTrim-letterSpacing)

	// vertical
	l = Line{simpleShape(english, latinFont, di.DirectionTTB)}
	run = &l[0]
	advance = run.Advance

	// no-op
	l.TrimLetterSpacing(0)
	tu.Assert(t, advance == run.Advance)

	run.AddLetterSpacing(letterSpacing)
	beforeTrim = run.Advance
	l.TrimLetterSpacing(letterSpacing)
	tu.Assert(t, run.Advance == beforeTrim-letterSpacing)
}

func TestTrimSpaceWrap(t *testing.T) {
	letterSpacing, _ := fixed.I(4), fixed.I(20)
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	english := []rune("Hello world ! : the end")

	wrap := func(dir di.Direction, letterSpacing fixed.Int26_6) Line {
		runs := cutRunInto(simpleShape(english, latinFont, dir), 2)
		// apply letter spacing
		for i := range runs {
			runs[i].AddLetterSpacing(letterSpacing)
		}

		// split in two lines
		lines, _ := (&LineWrapper{}).WrapParagraph(WrapConfig{}, 1200, english, NewSliceIterator(runs))
		tu.Assert(t, len(lines) == 2)
		return lines[0]
	}

	line := wrap(di.DirectionLTR, letterSpacing)
	tu.Assert(t, len(line) == 2)
	withSpacing := line[0].Advance + line[1].Advance

	line.TrimLetterSpacing(letterSpacing)
	withTrimmedSpacing := line[0].Advance + line[1].Advance
	tu.Assert(t, withTrimmedSpacing == withSpacing-letterSpacing)
}
