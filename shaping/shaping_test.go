package shaping

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	apiFont "github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
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
