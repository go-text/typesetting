package shaping

import (
	"fmt"
	"testing"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"golang.org/x/image/math/fixed"
)

func simpleShape(text []rune, face font.Face) Output {
	input := Input{
		Text:      text,
		RunStart:  0,
		RunEnd:    len(text),
		Direction: di.DirectionRTL,
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

	out := simpleShape(english, latinFont)
	for _, g := range out.Glyphs {
		fmt.Println(g.XAdvance.Round())
	}
	fmt.Println("Total", out.Advance.Round())
	out.addWordSpacing(english, fixed.I(20))
	fmt.Println("Total", out.Advance.Round())

	out = simpleShape(arabic, arabicFont)
	for _, g := range out.Glyphs {
		fmt.Println(g.XAdvance, g.XOffset)
	}
}
