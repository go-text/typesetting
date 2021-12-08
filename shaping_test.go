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
					Cluster: 0,
				},
				{
					Cluster: 1,
				},
				{
					Cluster: 3,
				},
				{
					Cluster: 3,
				},
				{
					Cluster: 4,
				},
				{
					Cluster: 4,
				},
			},
			expected: []Glyph{
				{
					Cluster:   0,
					NumRunes:  1,
					NumGlyphs: 1,
				},
				{
					Cluster:   1,
					NumRunes:  2,
					NumGlyphs: 1,
				},
				{
					Cluster:   3,
					NumRunes:  1,
					NumGlyphs: 2,
				},
				{
					Cluster:   3,
					NumRunes:  1,
					NumGlyphs: 2,
				},
				{
					Cluster:   4,
					NumRunes:  4,
					NumGlyphs: 2,
				},
				{
					Cluster:   4,
					NumRunes:  4,
					NumGlyphs: 2,
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
					Cluster: 4,
				},
				{
					Cluster: 4,
				},
				{
					Cluster: 3,
				},
				{
					Cluster: 3,
				},
				{
					Cluster: 1,
				},
				{
					Cluster: 0,
				},
			},
			expected: []Glyph{
				{
					Cluster:   4,
					NumRunes:  4,
					NumGlyphs: 2,
				},
				{
					Cluster:   4,
					NumRunes:  4,
					NumGlyphs: 2,
				},
				{
					Cluster:   3,
					NumRunes:  1,
					NumGlyphs: 2,
				},
				{
					Cluster:   3,
					NumRunes:  1,
					NumGlyphs: 2,
				},
				{
					Cluster:   1,
					NumRunes:  2,
					NumGlyphs: 1,
				},
				{
					Cluster:   0,
					NumRunes:  1,
					NumGlyphs: 1,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			countClusters(tc.glyphs, tc.textLen, tc.dir)
			for i := range tc.glyphs {
				g := tc.glyphs[i]
				e := tc.expected[i]
				if !(g.Cluster == e.Cluster && g.NumRunes == e.NumRunes && g.NumGlyphs == e.NumGlyphs) {
					t.Errorf("mismatch on glyph %d: expected cluster %d numRunes %d numGlyphs %d, got cluster %d numRunes %d numGlyphs %d", i, e.Cluster, e.NumRunes, e.NumGlyphs, g.Cluster, g.NumRunes, g.NumGlyphs)
				}
			}
		})
	}
}
