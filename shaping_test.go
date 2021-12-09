package shaping

import (
	"testing"

	"github.com/go-text/di"
)

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
					ClusterIndex:    0,
					RuneCount:  1,
					GlyphCount: 1,
				},
				{
					ClusterIndex:    1,
					RuneCount:  2,
					GlyphCount: 1,
				},
				{
					ClusterIndex:    3,
					RuneCount:  1,
					GlyphCount: 2,
				},
				{
					ClusterIndex:    3,
					RuneCount:  1,
					GlyphCount: 2,
				},
				{
					ClusterIndex:    4,
					RuneCount:  4,
					GlyphCount: 2,
				},
				{
					ClusterIndex:    4,
					RuneCount:  4,
					GlyphCount: 2,
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
					ClusterIndex:    4,
					RuneCount:  4,
					GlyphCount: 2,
				},
				{
					ClusterIndex:    4,
					RuneCount:  4,
					GlyphCount: 2,
				},
				{
					ClusterIndex:    3,
					RuneCount:  1,
					GlyphCount: 2,
				},
				{
					ClusterIndex:    3,
					RuneCount:  1,
					GlyphCount: 2,
				},
				{
					ClusterIndex:    1,
					RuneCount:  2,
					GlyphCount: 1,
				},
				{
					ClusterIndex:    0,
					RuneCount:  1,
					GlyphCount: 1,
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
