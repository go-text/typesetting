package shaping

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	hd "github.com/go-text/typesetting-utils/harfbuzz"
	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	apiFont "github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
	tu "github.com/go-text/typesetting/opentype/testutils"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
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
			countClusters(tc.glyphs, tc.textLen, tc.dir)
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
				ld, _ := loader.NewLoader(reader)
				_ = metadata.Metadata(ld)
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
				ld, _ := loader.NewLoader(reader)
				_ = metadata.Metadata(ld)
				_, _ = apiFont.NewFont(ld)
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
		allocBefore := heapSize()
		face, _ := font.ParseTTF(bytes.NewReader(bc.fontData))
		allocAfter := heapSize()
		additionnalAlloc := allocAfter - allocBefore
		fmt.Printf("%s On disk: %v KB, additional memory: %v KB\n", bc.name, onDiskSize/1024, additionnalAlloc/1024)
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
	input.FontFeatures = []FontFeature{{Tag: loader.MustNewTag("frac"), Value: 1}}
	out = shaper.Shape(input)
	tu.Assert(t, len(out.Glyphs) == 1)
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
		Size:      16 * 72 * 10,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}

	drawHGlyphs(shaper.Shape(input), filepath.Join(os.TempDir(), "shape_horiz.png"))

	input.Direction = di.DirectionTTB
	drawVGlyphs(shaper.Shape(input), filepath.Join(os.TempDir(), "shape_vert.png"))

	// Output:
}

var (
	red   = color.RGBA{R: 0xFF, A: 0xFF}
	green = color.RGBA{G: 0xFF, A: 0xFF}
	blue  = color.RGBA{B: 0xFF, A: 0xFF}
)

func drawVLine(img *image.RGBA, start image.Point, height int, c color.RGBA) {
	for y := start.Y; y <= start.Y+height; y++ {
		img.SetRGBA(start.X, y, c)
	}
}

func drawHLine(img *image.RGBA, start image.Point, width int, c color.RGBA) {
	for x := start.X; x <= start.X+width; x++ {
		img.SetRGBA(x, start.Y, c)
	}
}

func drawRect(img *image.RGBA, min, max image.Point, c color.RGBA) {
	for x := min.X; x <= max.X; x++ {
		for y := min.Y; y <= max.Y; y++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// assume horizontal direction
func drawHGlyphs(out Output, file string) {
	baseline := out.LineBounds.Ascent.Round()
	height := out.LineBounds.LineThickness().Round()
	width := out.Advance.Round()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// white background
	draw.Draw(img, img.Rect, image.NewUniform(color.White), image.Point{}, draw.Src)

	drawHLine(img, image.Pt(0, baseline), width, color.RGBA{A: 0xFF})

	dot := 0
	for _, g := range out.Glyphs {
		minX := dot + g.XOffset.Round() + g.XBearing.Round()
		maxX := minX + g.Width.Round()
		minY := baseline + g.YOffset.Round() - g.YBearing.Round()
		maxY := minY - g.Height.Round()

		drawRect(img, image.Pt(minX, minY), image.Pt(maxX, maxY), green)

		// draw the dot ...
		drawRect(img, image.Pt(dot-1, baseline-1), image.Pt(dot+1, baseline+1), color.RGBA{A: 0xFF})

		// ... and advance
		dot += g.XAdvance.Round()
		drawVLine(img, image.Pt(dot, 0), height, red)
	}

	f, _ := os.Create(file)
	_ = png.Encode(f, img)
}

// assume vertical direction
func drawVGlyphs(out Output, file string) {
	baseline := -out.GlyphBounds.Descent.Round()
	width := out.GlyphBounds.LineThickness().Round()
	height := -out.Advance.Round()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// white background
	draw.Draw(img, img.Rect, image.NewUniform(color.White), image.Point{}, draw.Src)

	drawVLine(img, image.Pt(baseline, 0), height, color.RGBA{A: 0xFF})

	dot := 0
	for _, g := range out.Glyphs {
		dot += -g.YAdvance.Round()

		minX := baseline + g.XOffset.Round() + g.XBearing.Round()
		maxX := minX + g.Width.Round()

		minY := dot + g.YOffset.Round() - g.YBearing.Round()
		maxY := minY - g.Height.Round()

		drawRect(img, image.Pt(minX, minY), image.Pt(maxX, maxY), green)

		// draw the dot ...
		drawRect(img, image.Pt(baseline-1, dot-1), image.Pt(baseline+1, dot+1), color.RGBA{A: 0xFF})

		// ... and advance
		drawHLine(img, image.Pt(0, dot), width, red)
	}

	f, _ := os.Create(file)
	_ = png.Encode(f, img)
}
