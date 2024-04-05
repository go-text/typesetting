package harfbuzz

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

// Unit tests for glyph advance Widths and extents of TrueType variable fonts
// ported from harfbuzz/test/api/test-ot-metrics-tt-var.c Copyright © 2019 Adobe Inc. Michiharu Ariza

func TestExtentsTtVar(t *testing.T) {
	ft := openFontFile(t, "fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
	font := NewFont(font.NewFace(ft))

	extents, result := font.GlyphExtents(2)
	tu.Assert(t, result)

	assertEqualInt32(t, extents.XBearing, 10)
	assertEqualInt32(t, extents.YBearing, 846)
	assertEqualInt32(t, extents.Width, 500)
	assertEqualInt32(t, extents.Height, -846)

	coords := [1]float32{500.0}
	font.SetVarCoordsDesign(coords[:])

	extents, result = font.GlyphExtents(2)
	tu.Assert(t, result)
	assertEqualInt32(t, extents.XBearing, 0)
	assertEqualInt32(t, extents.YBearing, 874)
	assertEqualInt32(t, extents.Width, 550)
	assertEqualInt32(t, extents.Height, -874)
}

func TestAdvanceTtVarNohvar(t *testing.T) {
	ft := openFontFile(t, "fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
	font := NewFont(font.NewFace(ft))

	x, y := font.GlyphAdvanceForDirection(2, LeftToRight)

	assertEqualInt32(t, x, 520)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(2, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1257)

	coords := []float32{500.0}
	font.SetVarCoordsDesign(coords)
	x, y = font.GlyphAdvanceForDirection(2, LeftToRight)

	assertEqualInt32(t, x, 551)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(2, TopToBottom)
	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1257)
}

func TestAdvanceTtVarHvarvvar(t *testing.T) {
	ft := openFontFile(t, "fonts/SourceSerifVariable-Roman-VVAR.abc.ttf")
	font := NewFont(font.NewFace(ft))

	x, y := font.GlyphAdvanceForDirection(1, LeftToRight)

	assertEqualInt32(t, x, 508)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(1, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1000)

	coords := []float32{700.0}
	font.SetVarCoordsDesign(coords)
	x, y = font.GlyphAdvanceForDirection(1, LeftToRight)

	assertEqualInt32(t, x, 531)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(1, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1012)
}

func TestAdvanceTtVarAnchor(t *testing.T) {
	ft := openFontFile(t, "fonts/SourceSansVariable-Roman.anchor.ttf")
	font := NewFont(font.NewFace(ft))

	extents, result := font.GlyphExtents(2)
	tu.Assert(t, result)

	assertEqualInt32(t, extents.XBearing, 56)
	assertEqualInt32(t, extents.YBearing, 672)
	assertEqualInt32(t, extents.Width, 556)
	assertEqualInt32(t, extents.Height, -684)

	coords := []float32{500.0}
	font.SetVarCoordsDesign(coords)
	extents, result = font.GlyphExtents(2)
	tu.Assert(t, result)

	assertEqualInt32(t, extents.XBearing, 50)
	assertEqualInt32(t, extents.YBearing, 667)
	assertEqualInt32(t, extents.Width, 592)
	assertEqualInt32(t, extents.Height, -679)
}

func TestExtentsTtVarComp(t *testing.T) {
	ft := openFontFile(t, "fonts/SourceSansVariable-Roman.modcomp.ttf")
	font := NewFont(font.NewFace(ft))

	coords := []float32{800.0}
	font.SetVarCoordsDesign(coords)

	extents, result := font.GlyphExtents(2) /* Ccedilla, cedilla y-scaled by 0.8, with unscaled component offset */
	tu.Assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 663)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -894)

	extents, result = font.GlyphExtents(3) /* Cacute, acute y-scaled by 0.8, with unscaled component offset (default) */
	tu.Assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 909)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -921)

	extents, result = font.GlyphExtents(4) /* Ccaron, caron y-scaled by 0.8, with scaled component offset */
	tu.Assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 866)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -878)
}

func TestAdvanceTtVarCompV(t *testing.T) {
	ft := openFontFile(t, "fonts/SourceSansVariable-Roman.modcomp.ttf")
	font := NewFont(font.NewFace(ft))

	coords := []float32{800.0}
	font.SetVarCoordsDesign(coords)

	x, y := font.GlyphAdvanceForDirection(2, TopToBottom) /* No VVAR; 'C' in composite Ccedilla determines metrics */

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -991)

	x, y = font.getGlyphOriginForDirection(2, TopToBottom)

	assertEqualInt32(t, x, 291)
	assertEqualInt32(t, y, 1012)
}

func TestAdvanceTtVarGvarInfer(t *testing.T) {
	ft := openFontFile(t, "fonts/TestGVAREight.ttf")
	coords := []font.VarCoord{100}

	face := font.NewFace(ft)
	face.SetCoords(coords)
	font := NewFont(face)
	_, ok := font.GlyphExtents(4)
	tu.Assert(t, ok)
}

func TestLigCarets(t *testing.T) {
	ft := openFontFile(t, "fonts/NotoNastaliqUrdu-Regular.ttf")
	font := NewFont(font.NewFace(ft))
	font.XScale, font.YScale = int32(ft.Upem())*2, int32(ft.Upem())*4

	/* call with no result */
	if L := len(font.GetOTLigatureCarets(LeftToRight, 188)); L != 0 {
		t.Fatalf("for glyph %d, expected %d, got %d", 188, 0, L)
	}
	if L := len(font.GetOTLigatureCarets(LeftToRight, 1021)); L != 0 {
		t.Fatalf("for glyph %d, expected %d, got %d", 1021, 0, L)
	}

	/* a glyph with 3 ligature carets */
	carets := font.GetOTLigatureCarets(LeftToRight, 1020)
	expected := []Position{2718, 5438, 8156}
	if !reflect.DeepEqual(expected, carets) {
		t.Fatalf("for glyph %d, expected %v, got %v", 1020, expected, carets)
	}

	/* a glyph with 1 ligature caret */
	carets = font.GetOTLigatureCarets(LeftToRight, 1022)
	expected = []Position{3530}
	if !reflect.DeepEqual(expected, carets) {
		t.Fatalf("for glyph %d, expected %v, got %v", 1022, expected, carets)
	}

	/* a glyph with 2 ligature carets */
	carets = font.GetOTLigatureCarets(LeftToRight, 1023)
	expected = []Position{2352, 4706}
	if !reflect.DeepEqual(expected, carets) {
		t.Fatalf("for glyph %d, expected %v, got %v", 1023, expected, carets)
	}
}

func TestColorGlyphExtents(t *testing.T) {
	// TODO: Support COLR table
	t.Skip()

	/* This font contains a COLRv1 glyph with a ClipBox,
	 * and various components without. The main thing
	 * we test here is that glyphs with no paint return
	 * 0,0,0,0 and not meaningless numbers.
	 */
	ft := openFontFile(t, "fonts/adwaita.ttf")
	font := NewFont(font.NewFace(ft))

	for _, test := range []struct {
		gid     GID
		extents GlyphExtents
		ok      bool
	}{
		{0, GlyphExtents{0, 0, 0, 0}, true},
		{1, GlyphExtents{0, 0, 0, 0}, true},
		{2, GlyphExtents{180, 960, 1060, -1220}, true},
		{3, GlyphExtents{188, 950, 900, -1200}, true},
		{4, GlyphExtents{413, 50, 150, -75}, true},
		{5, GlyphExtents{638, 350, 600, -600}, true},
		{1000, GlyphExtents{}, false},
	} {
		gotExtents, gotOK := font.GlyphExtents(test.gid)
		tu.Assert(t, gotOK == test.ok)
		tu.AssertC(t, gotExtents == test.extents, fmt.Sprintf("Glyph %d", test.gid))
	}
}

func TestNames(t *testing.T) {
	// from 'post' table
	ft := openFontFile(t, "fonts/adwaita.ttf")

	tu.Assert(t, ft.GlyphName(0) == ".notdef")
	tu.Assert(t, ft.GlyphName(1) == ".space")
	tu.Assert(t, ft.GlyphName(2) == "icon0")
	tu.Assert(t, ft.GlyphName(3) == "icon0.0")
	tu.Assert(t, ft.GlyphName(4) == "icon0.1")
	tu.Assert(t, ft.GlyphName(5) == "icon0.2")
	/* beyond last glyph */
	tu.Assert(t, ft.GlyphName(1000) == "")

	// from 'cff' table
	ft = openFontFile(t, "fonts/SourceSansPro-Regular.otf")

	tu.Assert(t, ft.GlyphName(0) == ".notdef")
	tu.Assert(t, ft.GlyphName(1) == "space")
	tu.Assert(t, ft.GlyphName(2) == "A")
	/* beyond last glyph */
	tu.Assert(t, ft.GlyphName(2000) == "")
}

func TestUnifont(t *testing.T) {
	// https://github.com/go-text/typesetting/issues/140
	ft := openFontFileTT(t, "bitmap/unifont-15.1.05.otf")

	buf := NewBuffer()
	buf.Props.Language = "en-us"
	buf.Props.Script = language.Latin
	buf.Props.Direction = LeftToRight
	buf.AddRunes([]rune{'a'}, 0, 1)
	font := NewFont(font.NewFace(ft))
	buf.Shape(font, nil) // just check for crashes
}
