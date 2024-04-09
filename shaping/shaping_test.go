package shaping

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	hd "github.com/go-text/typesetting-utils/harfbuzz"
	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

func TestShape(t *testing.T) {
	textInput := []rune("Lorem ipsum.")
	face := benchEnFace
	input := Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}
	shaper := HarfbuzzShaper{}
	out := shaper.Shape(input)
	if expected := (Range{Offset: 0, Count: len(textInput)}); out.Runes != expected {
		t.Errorf("expected runes %#+v, got %#+v", expected, out.Runes)
	}
	if face != out.Face {
		t.Error("shaper did not propagate input font face to output")
	}
	// Ensure properties of the run come out correctly adjusted if it starts after
	// the beginning of the text.
	input.RunStart = 6
	input.RunEnd = 8
	out = shaper.Shape(input)
	if expected := (Range{Offset: 6, Count: 2}); out.Runes != expected {
		t.Errorf("expected runes %#+v, got %#+v", expected, out.Runes)
	}
	if face != out.Face {
		t.Error("shaper did not propagate input font face to output")
	}
	for i, g := range out.Glyphs {
		if g.GlyphCount != 1 {
			t.Errorf("out.Glyphs[%d].GlyphCount != %d, is %d", i, 1, g.GlyphCount)
		}
		if g.RuneCount != 1 {
			t.Errorf("out.Runes[%d].RuneCount != %d, is %d", i, 1, g.RuneCount)
		}
	}
}

func TestCountClusters(t *testing.T) {
	type testcase struct {
		name     string
		textLen  int
		dir      di.Direction
		glyphs   []Glyph
		expected []Glyph
	}
	for _, tc := range []testcase{
		{
			name: "empty",
		},
		{
			name:    "ltr",
			textLen: 8,
			dir:     di.DirectionLTR,
			// Addressing the runes of text as A[0]-A[9] and the glyphs as
			// G[0]-G[5], this input models the following:
			// A[0] => G[0]
			// A[1],A[2] => G[1] (ligature)
			// A[3] => G[2],G[3] (expansion)
			// A[4],A[5],A[6],A[7] => G[4],G[5] (reorder, ligature, etc...)
			glyphs: []Glyph{
				{
					ClusterIndex: 0,
				},
				{
					ClusterIndex: 1,
				},
				{
					ClusterIndex: 3,
				},
				{
					ClusterIndex: 3,
				},
				{
					ClusterIndex: 4,
				},
				{
					ClusterIndex: 4,
				},
			},
			expected: []Glyph{
				{
					ClusterIndex: 0,
					RuneCount:    1,
					GlyphCount:   1,
				},
				{
					ClusterIndex: 1,
					RuneCount:    2,
					GlyphCount:   1,
				},
				{
					ClusterIndex: 3,
					RuneCount:    1,
					GlyphCount:   2,
				},
				{
					ClusterIndex: 3,
					RuneCount:    1,
					GlyphCount:   2,
				},
				{
					ClusterIndex: 4,
					RuneCount:    4,
					GlyphCount:   2,
				},
				{
					ClusterIndex: 4,
					RuneCount:    4,
					GlyphCount:   2,
				},
			},
		},
		{
			name:    "rtl",
			textLen: 8,
			dir:     di.DirectionRTL,
			// Addressing the runes of text as A[0]-A[9] and the glyphs as
			// G[0]-G[5], this input models the following:
			// A[0] => G[5]
			// A[1],A[2] => G[4] (ligature)
			// A[3] => G[2],G[3] (expansion)
			// A[4],A[5],A[6],A[7] => G[0],G[1] (reorder, ligature, etc...)
			glyphs: []Glyph{
				{
					ClusterIndex: 4,
				},
				{
					ClusterIndex: 4,
				},
				{
					ClusterIndex: 3,
				},
				{
					ClusterIndex: 3,
				},
				{
					ClusterIndex: 1,
				},
				{
					ClusterIndex: 0,
				},
			},
			expected: []Glyph{
				{
					ClusterIndex: 4,
					RuneCount:    4,
					GlyphCount:   2,
				},
				{
					ClusterIndex: 4,
					RuneCount:    4,
					GlyphCount:   2,
				},
				{
					ClusterIndex: 3,
					RuneCount:    1,
					GlyphCount:   2,
				},
				{
					ClusterIndex: 3,
					RuneCount:    1,
					GlyphCount:   2,
				},
				{
					ClusterIndex: 1,
					RuneCount:    2,
					GlyphCount:   1,
				},
				{
					ClusterIndex: 0,
					RuneCount:    1,
					GlyphCount:   1,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			countClusters(tc.glyphs, tc.textLen, tc.dir.Progression())
			for i := range tc.glyphs {
				g := tc.glyphs[i]
				e := tc.expected[i]
				if !(g.ClusterIndex == e.ClusterIndex && g.RuneCount == e.RuneCount && g.GlyphCount == e.GlyphCount) {
					t.Errorf("mismatch on glyph %d: expected cluster %d RuneCount %d GlyphCount %d, got cluster %d RuneCount %d GlyphCount %d", i, e.ClusterIndex, e.RuneCount, e.GlyphCount, g.ClusterIndex, g.RuneCount, g.GlyphCount)
				}
			}
		})
	}
}

func BenchmarkShaping(b *testing.B) {
	for _, langInfo := range benchLangs {
		for _, size := range []int{10, 100, 1000} {
			for _, cacheSize := range []int{0, 5} {
				b.Run(fmt.Sprintf("%drunes-%s-%dfontCache", size, langInfo.name, cacheSize), func(b *testing.B) {
					input := Input{
						Text:      langInfo.text[:size],
						RunStart:  0,
						RunEnd:    size,
						Direction: langInfo.dir,
						Face:      langInfo.face,
						Size:      16 * 72,
						Script:    langInfo.script,
						Language:  langInfo.lang,
					}
					var shaper HarfbuzzShaper
					shaper.SetFontCacheSize(cacheSize)
					var out Output
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						out = shaper.Shape(input)
					}
					_ = out
				})
			}
		}
	}
}

func BenchmarkFontLoad(b *testing.B) {
	arabicBytes, err := os.ReadFile("../font/testdata/Amiri-Regular.ttf")
	if err != nil {
		b.Errorf("failed loading arabic font data: %v", err)
	}
	robotoBytes, err := os.ReadFile("../font/testdata/Roboto-Regular.ttf")
	if err != nil {
		b.Errorf("failed loading roboto font data: %v", err)
	}
	latinBytes := goregular.TTF
	latinMonoBytes := gomono.TTF
	type benchcase struct {
		name     string
		fontData []byte
	}
	for _, bc := range []benchcase{
		{
			name:     "arabic:amiri regular",
			fontData: arabicBytes,
		},
		{
			name:     "latin:go regular",
			fontData: latinBytes,
		},
		{
			name:     "latin:go mono",
			fontData: latinMonoBytes,
		},
		{
			name:     "latin:roboto regular",
			fontData: robotoBytes,
		},
	} {
		b.Run(bc.name, func(b *testing.B) {
			reader := bytes.NewReader(bc.fontData)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				face, _ := font.ParseTTF(reader)
				_ = face
			}
		})
	}
}

func BenchmarkFontMetadata(b *testing.B) {
	arabicBytes, err := os.ReadFile("../font/testdata/Amiri-Regular.ttf")
	if err != nil {
		b.Errorf("failed loading arabic font data: %v", err)
	}
	robotoBytes, err := os.ReadFile("../font/testdata/Roboto-Regular.ttf")
	if err != nil {
		b.Errorf("failed loading roboto font data: %v", err)
	}
	latinBytes := goregular.TTF
	latinMonoBytes := gomono.TTF
	type benchcase struct {
		name     string
		fontData []byte
	}
	for _, bc := range []benchcase{
		{
			name:     "arabic:amiri regular",
			fontData: arabicBytes,
		},
		{
			name:     "latin:go regular",
			fontData: latinBytes,
		},
		{
			name:     "latin:go mono",
			fontData: latinMonoBytes,
		},
		{
			name:     "latin:roboto regular",
			fontData: robotoBytes,
		},
	} {
		b.Run(bc.name, func(b *testing.B) {
			reader := bytes.NewReader(bc.fontData)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ld, _ := ot.NewLoader(reader)
				_, _ = font.Describe(ld, nil)
			}
		})
	}
}

func BenchmarkFontMetadataAndParse(b *testing.B) {
	arabicBytes, err := os.ReadFile("../font/testdata/Amiri-Regular.ttf")
	if err != nil {
		b.Errorf("failed loading arabic font data: %v", err)
	}
	robotoBytes, err := os.ReadFile("../font/testdata/Roboto-Regular.ttf")
	if err != nil {
		b.Errorf("failed loading roboto font data: %v", err)
	}
	latinBytes := goregular.TTF
	latinMonoBytes := gomono.TTF
	type benchcase struct {
		name     string
		fontData []byte
	}
	for _, bc := range []benchcase{
		{
			name:     "arabic:amiri regular",
			fontData: arabicBytes,
		},
		{
			name:     "latin:go regular",
			fontData: latinBytes,
		},
		{
			name:     "latin:go mono",
			fontData: latinMonoBytes,
		},
		{
			name:     "latin:roboto regular",
			fontData: robotoBytes,
		},
	} {
		b.Run(bc.name, func(b *testing.B) {
			reader := bytes.NewReader(bc.fontData)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ld, _ := ot.NewLoader(reader)
				ft, _ := font.NewFont(ld)
				_ = ft.Describe()
			}
		})
	}
}

func BenchmarkShapingGlyphInfoExtraction(b *testing.B) {
	for _, langInfo := range benchLangs {
		for _, size := range []int{10, 100, 1000} {
			b.Run(fmt.Sprintf("%drunes-%s", size, langInfo.name), func(b *testing.B) {
				input := Input{
					Text:      langInfo.text[:size],
					RunStart:  0,
					RunEnd:    size,
					Direction: langInfo.dir,
					Face:      langInfo.face,
					Size:      16 * 72,
					Script:    langInfo.script,
					Language:  langInfo.lang,
				}
				var shaper HarfbuzzShaper
				out := shaper.Shape(input)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, g := range out.Glyphs {
						_ = out.Face.GlyphData(g.GlyphID)
					}
				}
				_ = out
			})
		}
	}
}

func TestFontLoadHeapSize(t *testing.T) {
	arabicBytes, err := os.ReadFile("../font/testdata/Amiri-Regular.ttf")
	if err != nil {
		t.Errorf("failed loading arabic font data: %v", err)
	}
	robotoBytes, err := os.ReadFile("../font/testdata/Roboto-Regular.ttf")
	if err != nil {
		t.Errorf("failed loading roboto font data: %v", err)
	}
	latinBytes := goregular.TTF
	notoArabicBytes, err := td.Files.ReadFile("common/NotoSansArabic.ttf")
	if err != nil {
		t.Errorf("failed loading noto font data: %v", err)
	}
	freeSerifBytes, err := td.Files.ReadFile("common/FreeSerif.ttf")
	if err != nil {
		t.Errorf("failed loading free font data: %v", err)
	}

	type benchcase struct {
		name     string
		fontData []byte
	}
	for _, bc := range []benchcase{
		{
			name:     "arabic:amiri regular",
			fontData: arabicBytes,
		},
		{
			name:     "latin:go regular",
			fontData: latinBytes,
		},
		{
			name:     "latin:roboto regular",
			fontData: robotoBytes,
		},
		{
			name:     "arabic:noto sans arabic regular",
			fontData: notoArabicBytes,
		},
		{
			name:     "free serif regular",
			fontData: freeSerifBytes,
		},
	} {
		onDiskSize := len(bc.fontData)
		hs := heapSize()

		loader, _ := ot.NewLoader(bytes.NewReader(bc.fontData))
		ft, _ := font.NewFont(loader)
		fontSize := heapSize() - hs
		hs = heapSize()
		face := font.NewFace(ft)
		faceSize := heapSize() - hs
		fmt.Printf("%s On disk: %v KB, font memory: %v KB, face memory %v KB\n", bc.name, onDiskSize/1024, fontSize/1024, faceSize/1024)
		_ = face.Upem()
	}
}

func heapSize() uint64 {
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}

func TestFeatures(t *testing.T) {
	face := loadOpentypeFont(t, "../font/testdata/UbuntuMono-R.ttf")
	textInput := []rune("1/2")
	input := Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}
	shaper := HarfbuzzShaper{}
	// without 'frac' feature
	out := shaper.Shape(input)
	tu.Assert(t, len(out.Glyphs) == 3)

	// now with 'frac' enabled
	input.FontFeatures = []FontFeature{{Tag: ot.MustNewTag("frac"), Value: 1}}
	out = shaper.Shape(input)
	tu.Assert(t, len(out.Glyphs) == 1)
}

func TestShapeVertical(t *testing.T) {
	// consistency check on the internal axis switch
	// for sideways vertical text
	textInput := []rune("Lorem ipsum.")
	face := benchEnFace
	input := Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionTTB,
		Face:      face,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}
	shaper := HarfbuzzShaper{}

	for _, test := range []struct {
		dir      di.Direction
		sideways bool
	}{
		{di.DirectionTTB, false},
		{di.DirectionTTB, true},
		{di.DirectionBTT, false},
		{di.DirectionBTT, true},
	} {
		input.Direction = test.dir
		input.Direction.SetSideways(test.sideways)
		out := shaper.Shape(input)
		tu.Assert(t, out.Direction.Progression() == test.dir.Progression())
		tu.Assert(t, out.Direction.IsSideways() == test.sideways)
		tu.Assert(t, out.Advance < 0)
		tu.Assert(t, out.GlyphBounds.Ascent > 0 && out.GlyphBounds.Descent < 0)
	}
}

func TestCFF2(t *testing.T) {
	// regression test for https://github.com/go-text/typesetting/issues/118
	b, err := td.Files.ReadFile("common/NotoSansCJKjp-VF.otf")
	tu.AssertNoErr(t, err)

	face, err := font.ParseTTF(bytes.NewReader(b))
	tu.AssertNoErr(t, err)

	str := []rune("abcあいう")
	input := Input{
		Text:      str,
		RunStart:  0,
		RunEnd:    len(str),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      fixed.I(10),
	}
	out := (&HarfbuzzShaper{}).Shape(input)
	for _, g := range out.Glyphs {
		tu.Assert(t, g.Width > 0 && g.Height < 0)
	}
	tu.Assert(t, out.Advance > 0)
}

func TestShapeVerticalScripts(t *testing.T) {
	b, _ := td.Files.ReadFile("common/NotoSansMongolian-Regular.ttf")
	monF, _ := font.ParseTTF(bytes.NewReader(b))
	b, _ = td.Files.ReadFile("common/mplus-1p-regular.ttf")
	japF, _ := font.ParseTTF(bytes.NewReader(b))

	monT := []rune("ᠬᠦᠮᠦᠨ ᠪᠦᠷ ᠲᠥᠷᠥᠵᠦ")
	japT := []rune("青いそら…")
	mixedT := []rune("あHelloあUne phrase")

	var (
		seg    Segmenter
		shaper HarfbuzzShaper
	)

	{
		runs := seg.Split(Input{
			Text:      monT,
			RunEnd:    len(monT),
			Language:  language.NewLanguage("mn"),
			Size:      fixed.I(12 * 16),
			Direction: di.DirectionTTB,
		}, fixedFontmap{monF})
		tu.Assert(t, len(runs) == 1)

		line := Line{shaper.Shape(runs[0])}
		err := drawTextLine(line, filepath.Join(os.TempDir(), "shape_vert_mongolian.png"))
		tu.AssertNoErr(t, err)

		line.AdjustBaselines()
		err = drawTextLine(line, filepath.Join(os.TempDir(), "shape_vert_mongolian_adjusted.png"))
		tu.AssertNoErr(t, err)
	}
	{
		runs := seg.Split(Input{
			Text:      japT,
			RunEnd:    len(japT),
			Language:  language.NewLanguage("ja"),
			Size:      fixed.I(12 * 16),
			Direction: di.DirectionTTB,
		}, fixedFontmap{japF})
		tu.Assert(t, len(runs) == 2)
		line := Line{shaper.Shape(runs[0]), shaper.Shape(runs[1])}
		err := drawTextLine(line, filepath.Join(os.TempDir(), "shape_vert_japanese.png"))
		tu.AssertNoErr(t, err)
	}
	{
		runs := seg.Split(Input{
			Text:      mixedT,
			RunEnd:    len(mixedT),
			Language:  language.NewLanguage("ja"),
			Size:      fixed.I(12 * 16),
			Direction: di.DirectionTTB,
		}, fixedFontmap{japF})
		tu.Assert(t, len(runs) == 4)
		line := Line{shaper.Shape(runs[0]), shaper.Shape(runs[1]), shaper.Shape(runs[2]), shaper.Shape(runs[3])}
		err := drawTextLine(line, filepath.Join(os.TempDir(), "shape_vert_mixed.png"))
		tu.AssertNoErr(t, err)

		line.AdjustBaselines()
		err = drawTextLine(line, filepath.Join(os.TempDir(), "shape_vert_mixed_adjusted.png"))
		tu.AssertNoErr(t, err)
	}
}

func ExampleShaper_Shape() {
	textInput := []rune("abcdefghijklmnop")
	withKerningFont := "harfbuzz_reference/in-house/fonts/e39391c77a6321c2ac7a2d644de0396470cd4bfe.ttf"
	b, _ := hd.Files.ReadFile(withKerningFont)
	face, _ := font.ParseTTF(bytes.NewReader(b))

	shaper := HarfbuzzShaper{}
	input := Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      fixed.I(16 * 1000 / 72),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}

	horiz := shaper.Shape(input)
	drawTextLine(Line{horiz}, filepath.Join(os.TempDir(), "shape_horiz.png"))

	input.Direction = di.DirectionTTB
	drawTextLine(Line{shaper.Shape(input)}, filepath.Join(os.TempDir(), "shape_vert.png"))

	input.Direction.SetSideways(true)
	drawTextLine(Line{shaper.Shape(input)}, filepath.Join(os.TempDir(), "shape_vert_rotated.png"))

	input.Direction = di.DirectionBTT
	drawTextLine(Line{shaper.Shape(input)}, filepath.Join(os.TempDir(), "shape_vert_rev.png"))

	input.Direction.SetSideways(true)
	drawTextLine(Line{shaper.Shape(input)}, filepath.Join(os.TempDir(), "shape_vert_rev_rotated.png"))

	// Output:
}

func BenchmarkGlyphExtents(b *testing.B) {
	shaper := HarfbuzzShaper{}
	input := Input{
		Text:      []rune(benchParagraphLatin),
		RunStart:  0,
		RunEnd:    len([]rune(benchParagraphLatin)),
		Direction: di.DirectionLTR,
		Size:      fixed.I(16 * 1000 / 72),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}

	for i, file := range []string{
		"common/DejaVuSans.ttf",
		"common/Raleway-v4020-Regular.otf",
		"common/Commissioner-VF.ttf",
		"common/Commissioner-VF.ttf", // with variations
	} {
		r, _ := td.Files.ReadFile(file)
		face, err := font.ParseTTF(bytes.NewReader(r))
		tu.AssertNoErr(b, err)

		if i == 3 {
			face.SetVariations([]font.Variation{
				{Tag: ot.MustNewTag("wght"), Value: 200},
				{Tag: ot.MustNewTag("slnt"), Value: -8},
				{Tag: ot.MustNewTag("FLAR"), Value: 30},
				{Tag: ot.MustNewTag("VOLM"), Value: 80},
			})
		}
		input.Face = face

		b.Run(file, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = shaper.Shape(input)
			}
		})
	}
}
