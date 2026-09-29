package harfbuzz

import (
	"testing"

	"github.com/go-text/typesetting/font"
)

func TestRecategorize(t *testing.T) {
	runes := []rune{1615, 1617, 1614, 1616}
	ccc := []uint8{32, 27, 31, 33}
	exps := []uint8{230, 230, 230, 220}
	for i, r := range runes {
		exp := exps[i]
		got := recategorizeCombiningClass(r, ccc[i])
		if exp != got {
			t.Fatalf("for rune %d and class %d, expected %d, got %d", r, ccc[i], exp, got)
		}
	}
}

func TestFallbackFigureSpace(t *testing.T) {
	fnt := NewFont(font.NewFace(openFontFileTT(t, "common/Raleway-v4020-Regular.otf"))) // proportional digits
	b := NewBuffer()
	b.AddRunes([]rune{0x2007}, 0, -1)
	b.Info[0].setUnicodeProps(b)
	b.Info[0].setUnicodeSpaceFallbackType(spaceFigure)
	b.Pos = make([]GlyphPosition, 1)
	b.Props.Direction = LeftToRight
	fallbackSpaces(fnt, b)
	zero, _ := fnt.face.NominalGlyph('0')
	if exp, got := fnt.GlyphHAdvance(zero), b.Pos[0].XAdvance; exp != got {
		t.Fatalf("expected %d, got %d", exp, got)
	}
}
