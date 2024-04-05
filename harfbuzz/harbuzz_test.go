package harfbuzz

import (
	"bytes"
	"fmt"
	"testing"

	td "github.com/go-text/typesetting-utils/harfbuzz"
	otTD "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

func assertEqualInt(t *testing.T, expected, got int) {
	t.Helper()
	tu.AssertC(t, expected == got, fmt.Sprintf("expected %d, got %d", expected, got))
}

func assertEqualInt32(t *testing.T, got, expected int32) {
	t.Helper()
	tu.AssertC(t, expected == got, fmt.Sprintf("expected %d, got %d", expected, got))
}

// opens truetype fonts from opentype testdata.
func openFontFileTT(t *testing.T, filename string) *font.Font {
	t.Helper()

	f, err := otTD.Files.ReadFile(filename)
	tu.AssertNoErr(t, err)

	fp, err := ot.NewLoader(bytes.NewReader(f))
	tu.AssertNoErr(t, err)

	out, err := font.NewFont(fp)
	tu.AssertNoErr(t, err)

	return out
}

// opens truetype fonts from harfbuzz testdata,
// expecting a single file
func openFontFile(t testing.TB, filename string) *font.Font {
	f, err := td.Files.ReadFile(filename)
	tu.AssertNoErr(t, err)

	fp, err := ot.NewLoader(bytes.NewReader(f))
	tu.AssertNoErr(t, err)

	out, err := font.NewFont(fp)
	tu.AssertNoErr(t, err)

	return out
}

func TestDirection(t *testing.T) {
	tu.Assert(t, LeftToRight.isHorizontal() && !LeftToRight.isVertical())
	tu.Assert(t, RightToLeft.isHorizontal() && !RightToLeft.isVertical())
	tu.Assert(t, !TopToBottom.isHorizontal() && TopToBottom.isVertical())
	tu.Assert(t, !BottomToTop.isHorizontal() && BottomToTop.isVertical())

	tu.Assert(t, LeftToRight.isForward())
	tu.Assert(t, TopToBottom.isForward())
	tu.Assert(t, !RightToLeft.isForward())
	tu.Assert(t, !BottomToTop.isForward())

	tu.Assert(t, !LeftToRight.isBackward())
	tu.Assert(t, !TopToBottom.isBackward())
	tu.Assert(t, RightToLeft.isBackward())
	tu.Assert(t, BottomToTop.isBackward())

	tu.Assert(t, BottomToTop.Reverse() == TopToBottom)
	tu.Assert(t, TopToBottom.Reverse() == BottomToTop)
	tu.Assert(t, LeftToRight.Reverse() == RightToLeft)
	tu.Assert(t, RightToLeft.Reverse() == LeftToRight)
}

func TestFlag(t *testing.T) {
	if (glyphFlagDefined & (glyphFlagDefined + 1)) != 0 {
		t.Error("assertion failed")
	}
}

func TestTypesLanguage(t *testing.T) {
	fa := language.NewLanguage("fa")
	faIR := language.NewLanguage("fa_IR")
	faIr := language.NewLanguage("fa-ir")
	en := language.NewLanguage("en")

	tu.Assert(t, fa != "")
	tu.Assert(t, faIR != "")
	tu.Assert(t, faIR == faIr)

	tu.Assert(t, en != "")
	tu.Assert(t, en != fa)

	/* Test recall */
	tu.Assert(t, en == language.NewLanguage("en"))
	tu.Assert(t, en == language.NewLanguage("eN"))
	tu.Assert(t, en == language.NewLanguage("En"))

	tu.Assert(t, language.NewLanguage("") == "")
	tu.Assert(t, language.NewLanguage("e") != "")
}

func TestParseVariations(t *testing.T) {
	datas := [...]struct {
		input    string
		expected font.Variation
	}{
		{" frea=45.78", font.Variation{Tag: ot.MustNewTag("frea"), Value: 45.78}},
		{"G45E=45", font.Variation{Tag: ot.MustNewTag("G45E"), Value: 45}},
		{"fAAD 45.78", font.Variation{Tag: ot.MustNewTag("fAAD"), Value: 45.78}},
		{"fr 45.78", font.Variation{Tag: ot.MustNewTag("fr  "), Value: 45.78}},
		{"fr=45.78", font.Variation{Tag: ot.MustNewTag("fr  "), Value: 45.78}},
		{"fr=-45.4", font.Variation{Tag: ot.MustNewTag("fr  "), Value: -45.4}},
		{"'fr45'=-45.4", font.Variation{Tag: ot.MustNewTag("fr45"), Value: -45.4}}, // with quotes
		{`"frZD"=-45.4`, font.Variation{Tag: ot.MustNewTag("frZD"), Value: -45.4}}, // with quotes
	}
	for _, data := range datas {
		out, err := ParseVariation(data.input)
		if err != nil {
			t.Fatalf("error on %s: %s", data.input, err)
		}
		if out != data.expected {
			t.Fatalf("for %s, expected %v, got %v", data.input, data.expected, out)
		}
	}
}

func TestParseFeature(t *testing.T) {
	inputs := [...]string{
		"kern",
		"+kern",
		"-kern",
		"kern=0",
		"kern=1",
		"aalt=2",
		"kern[]",
		"kern[:]",
		"kern[5:]",
		"kern[:5]",
		"kern[3:5]",
		"kern[3]",
		"aalt[3:5]=2",
	}
	expecteds := [...]Feature{
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 0, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 0, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("aalt"), Value: 2, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 5, End: FeatureGlobalEnd},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 0, End: 5},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 3, End: 5},
		{Tag: ot.MustNewTag("kern"), Value: 1, Start: 3, End: 4},
		{Tag: ot.MustNewTag("aalt"), Value: 2, Start: 3, End: 5},
	}
	for i, input := range inputs {
		f, err := ParseFeature(input)
		if err != nil {
			t.Fatalf("unexpected error on input <%s> : %s", input, err)
		}
		if exp := expecteds[i]; f != exp {
			t.Fatalf("for <%s>, expected %v, got %v", input, exp, f)
		}
	}
}

func TestExample(t *testing.T) {
	ft := openFontFileTT(t, "common/NotoSansArabic.ttf")
	buffer := NewBuffer()

	// runes := []rune("This is a line to shape..")
	runes := []rune{0x0633, 0x064F, 0x0644, 0x064E, 0x0651, 0x0627, 0x0651, 0x0650, 0x0645, 0x062A, 0x06CC}
	buffer.AddRunes(runes, 0, -1)

	face := font.NewFace(ft)
	font := NewFont(face)
	buffer.GuessSegmentProperties()
	buffer.Shape(font, nil)

	for i, pos := range buffer.Pos {
		info := buffer.Info[i]
		ext, ok := face.GlyphExtents(info.Glyph)
		tu.AssertC(t, ok, fmt.Sprintf("invalid glyph %d", info.Glyph))

		fmt.Println(pos.XAdvance, pos.XOffset, ext.Width, ext.XBearing)
	}
}
