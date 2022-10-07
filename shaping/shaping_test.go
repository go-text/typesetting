package shaping

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
	"github.com/go-text/typesetting/di"
	"golang.org/x/image/font/gofont/goregular"
)

func TestShape(t *testing.T) {
	textInput := []rune("Lorem ipsum.")
	face, err := truetype.Parse(bytes.NewReader(goregular.TTF))
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
	out, err := shaper.Shape(input)
	if err != nil {
		t.Errorf("failed shaping: %v", err)
	}
	if expected := (Range{Offset: 0, Count: len(textInput)}); out.Runes != expected {
		t.Errorf("expected runes %#+v, got %#+v", expected, out.Runes)
	}
	if face != out.Face {
		t.Error("shaper did not propagate input font face to output")
	}
	input.RunStart = 6
	out, err = shaper.Shape(input)
	if err != nil {
		t.Errorf("failed shaping: %v", err)
	}
	if expected := (Range{Offset: 6, Count: len(textInput) - 6}); out.Runes != expected {
		t.Errorf("expected runes %#+v, got %#+v", expected, out.Runes)
	}
	if face != out.Face {
		t.Error("shaper did not propagate input font face to output")
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

func BenchmarkShapingLatin(b *testing.B) {
	textInput := []rune(benchParagraphLatin)
	face, err := truetype.Parse(bytes.NewReader(goregular.TTF))
	if err != nil {
		b.Skipf("failed parsing font: %v", err)
	}
	for _, size := range []int{10, 100, 1000, len(textInput)} {
		b.Run(fmt.Sprintf("%drunes", size), func(b *testing.B) {
			input := Input{
				Text:      textInput,
				RunStart:  0,
				RunEnd:    size,
				Direction: di.DirectionLTR,
				Face:      face,
				Size:      16,
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			}
			var shaper HarfbuzzShaper
			var out Output
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out, _ = shaper.Shape(input)
			}
			_ = out
		})
	}
}

func BenchmarkShapingArabic(b *testing.B) {
	textInput := []rune(benchParagraphArabic)
	for _, size := range []int{10, 100, 1000, len(textInput)} {
		b.Run(fmt.Sprintf("%drunes", size), func(b *testing.B) {
			input := Input{
				Text:      textInput,
				RunStart:  0,
				RunEnd:    size,
				Direction: di.DirectionLTR,
				Face:      urdu,
				Size:      16,
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			}
			var shaper HarfbuzzShaper
			var out Output
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out, _ = shaper.Shape(input)
			}
			_ = out
		})
	}
}
