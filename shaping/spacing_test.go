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

func TestRTL(t *testing.T) {
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	english := []rune("Hello world ! : the end")
	arabic := []rune("تثذرزسشص لمنهويء")

	addSpacing := fixed.I(20)
	out := simpleShape(english, latinFont, di.DirectionLTR)
	withoutSpacing := out.Advance
	out.addWordSpacing(english, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+5*addSpacing)

	out = simpleShape(arabic, arabicFont, di.DirectionRTL)
	withoutSpacing = out.Advance
	out.addWordSpacing(arabic, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+1*addSpacing)

	// vertical
	out = simpleShape(english, latinFont, di.DirectionTTB)
	withoutSpacing = out.Advance
	out.addWordSpacing(english, addSpacing)
	tu.Assert(t, out.Advance == withoutSpacing+5*addSpacing)
}
