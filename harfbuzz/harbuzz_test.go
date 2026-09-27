package harfbuzz

import (
	"bytes"
	"fmt"
	"testing"

	td "github.com/go-text/typesetting-utils/harfbuzz"
	otTD "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
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
	kern := ot.MustNewTag("kern")
	abcd := ot.MustNewTag("abcd")
	for _, test := range []struct {
		input    string
		expected Feature
	}{
		{"kern", Feature{kern, 1, 0, FeatureGlobalEnd}},
		{"+kern", Feature{kern, 1, 0, FeatureGlobalEnd}},
		{"-kern", Feature{kern, 0, 0, FeatureGlobalEnd}},
		{"kern=0", Feature{kern, 0, 0, FeatureGlobalEnd}},
		{"kern=1", Feature{kern, 1, 0, FeatureGlobalEnd}},
		{"aalt=2", Feature{ot.MustNewTag("aalt"), 2, 0, FeatureGlobalEnd}},
		{"kern[]", Feature{kern, 1, 0, FeatureGlobalEnd}},
		{"kern[:]", Feature{kern, 1, 0, FeatureGlobalEnd}},
		{"kern[5:]", Feature{kern, 1, 5, FeatureGlobalEnd}},
		{"kern[:5]", Feature{kern, 1, 0, 5}},
		{"kern[3:5]", Feature{kern, 1, 3, 5}},
		{"kern[3]", Feature{kern, 1, 3, 4}},
		{"aalt[3:5]=2", Feature{ot.MustNewTag("aalt"), 2, 3, 5}},
		{"abcd", Feature{abcd, 1, 0, FeatureGlobalEnd}},
		{"abcd=1", Feature{abcd, 1, 0, FeatureGlobalEnd}},
		{"+abcd", Feature{abcd, 1, 0, FeatureGlobalEnd}},
		{"abcd=0", Feature{abcd, 0, 0, FeatureGlobalEnd}},
		{"-abcd", Feature{abcd, 0, 0, FeatureGlobalEnd}},
		{"abcd=2", Feature{abcd, 2, 0, FeatureGlobalEnd}},
		{"+abcd=2", Feature{abcd, 2, 0, FeatureGlobalEnd}},
		{"-abcd=2", Feature{abcd, 2, 0, FeatureGlobalEnd}},
		{`"abcd" on`, Feature{abcd, 1, 0, FeatureGlobalEnd}},
		{`"abcd" off`, Feature{abcd, 0, 0, FeatureGlobalEnd}},
		{`"abcd" 1`, Feature{abcd, 1, 0, FeatureGlobalEnd}},
		{`"abcd" 0`, Feature{abcd, 0, 0, FeatureGlobalEnd}},
		{`"abcd" 2`, Feature{abcd, 2, 0, FeatureGlobalEnd}},
		{"abcd[0]", Feature{abcd, 1, 0, 1}},
		{"abcd[1]", Feature{abcd, 1, 1, 2}},
		{"abcd[1]=1", Feature{abcd, 1, 1, 2}},
		{"abcd[1]=2", Feature{abcd, 2, 1, 2}},
		{"abcd[1]=0", Feature{abcd, 0, 1, 2}},
		{"abcd[]", Feature{abcd, 1, 0, FeatureGlobalEnd}},
		{"abcd[:]", Feature{abcd, 1, 0, FeatureGlobalEnd}},
		{"abcd[1:]", Feature{abcd, 1, 1, FeatureGlobalEnd}},
		{"abcd[:1]", Feature{abcd, 1, 0, 1}},
		{"abcd[1:3]", Feature{abcd, 1, 1, 3}},
		{"abcd[1:2]=1", Feature{abcd, 1, 1, 2}},
		{"abcd[1:4]=2", Feature{abcd, 2, 1, 4}},
	} {
		f, err := ParseFeature(test.input)
		tu.AssertNoErr(t, err)
		tu.Assert(t, f == test.expected)
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

func TestPropagateAttachmentOffsetsNegativeChain(t *testing.T) {
	// cross-stream kerx attaches every glyph to the previous one, including the first
	pos := []GlyphPosition{{attachChain: -1, attachType: attachTypeCursive}}
	propagateAttachmentOffsets(pos, 0, LeftToRight) // must not panic
}

func TestWouldApplyContext2OutOfRangeClass(t *testing.T) {
	c := wouldApplyContext{glyphs: []GID{0}}
	classDef := tables.ClassDef1{ClassValueArray: []uint16{5}}
	tu.Assert(t, !c.wouldApplyLookupContext2(tables.SequenceContextFormat2{ClassDef: classDef}, 0, 0))
	tu.Assert(t, !c.wouldApplyLookupChainedContext2(tables.ChainedSequenceContextFormat2{InputClassDef: classDef}, 0, 0))
}

func TestMarkFilteringSetOutOfRange(t *testing.T) {
	var c otApplyContext
	props := uint32(font.UseMarkFilteringSet) | 3<<16
	tu.Assert(t, !c.matchPropertiesMark(&GlyphInfo{}, tables.GPMark, props))
}

func TestApplyForwardBufferGrowth(t *testing.T) {
	cov := func(gs ...tables.GlyphID) tables.Coverage1 { return tables.Coverage1{Glyphs: gs} }
	ft := &font.Font{}
	ft.GSUB.Lookups = []font.GSUBLookup{
		// the lookup records run out of order. The multiple substitution goes
		// first, then applyLookup rewinds the buffer and grows buffer.Info
		{Subtables: []tables.GSUBLookup{tables.ContextualSubs{Data: tables.ContextualSubs3{
			Coverages:        []tables.Coverage{cov(1), cov(2)},
			SeqLookupRecords: []tables.SequenceLookupRecord{{SequenceIndex: 1, LookupListIndex: 2}, {SequenceIndex: 0, LookupListIndex: 1}},
		}}}},
		{Subtables: []tables.GSUBLookup{tables.SingleSubs{Data: tables.SingleSubstData2{Coverage: cov(1), SubstituteGlyphIDs: []tables.GlyphID{10}}}}},
		{Subtables: []tables.GSUBLookup{tables.MultipleSubs{Coverage: cov(2), Sequences: []tables.Sequence{{SubstituteGlyphIDs: []tables.GlyphID{20, 21}}}}}},
	}
	fnt := NewFont(font.NewFace(ft))

	b := NewBuffer()
	b.Info = []GlyphInfo{{Glyph: 1, Mask: 1}, {Glyph: 2, Mask: 1}, {Glyph: 1, Mask: 1}, {Glyph: 2, Mask: 1}}
	var c otApplyContext
	c.reset(0, fnt, b)
	c.recurseFunc = applyRecurseGSUB
	c.substituteLookup(&fnt.gsubAccels[0])

	var got []GID
	for _, info := range b.Info {
		got = append(got, info.Glyph)
	}
	tu.AssertC(t, fmt.Sprint(got) == "[10 20 21 10 20 21]", fmt.Sprint(got))
}

func TestShapePlanCacheVariations(t *testing.T) {
	// Commissioner substitutes '$' through an 'rvrn' feature variation at heavy weights
	fnt := NewFont(font.NewFace(openFontFileTT(t, "common/Commissioner-VF.ttf")))
	buffer := NewBuffer()
	shape := func(weight float32) GID {
		fnt.SetVarCoordsDesign([]float32{weight, 0, 0, 0})
		buffer.Clear()
		buffer.AddRunes([]rune("$"), 0, -1)
		buffer.Props = SegmentProperties{Direction: LeftToRight, Script: language.Latin, Language: "en"}
		buffer.Shape(fnt, nil)
		return buffer.Info[0].Glyph
	}
	assertEqualInt(t, 954, int(shape(400)))
	assertEqualInt(t, 1117, int(shape(900)))
}
