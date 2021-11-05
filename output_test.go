// SPDX-License-Identifier: Unlicense OR MIT

package shaping_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/go-text/di"
	"github.com/go-text/shaping"
	"golang.org/x/image/math/fixed"
)

// extenter fulfills the shaping.GlyphExtenter interface
// by providing extents from its configured map.
type extenter struct {
	extents map[fonts.GID]harfbuzz.GlyphExtents
}

// GlyphExtents returns the extents for the provided gid, if they are set.
func (e *extenter) GlyphExtents(gid fonts.GID) (harfbuzz.GlyphExtents, bool) {
	extents, ok := e.extents[gid]
	return extents, ok
}

// ExtentsForDirection is a stub. This feature is totally font-dependent with no
// actual processing logic in this package.
func (e *extenter) ExtentsForDirection(_ harfbuzz.Direction) fonts.FontExtents {
	return fonts.FontExtents{
		LineGap:   0,
		Ascender:  15,
		Descender: -15,
	}
}

// Ensure that *extenter is a shaping.GlyphExtenter
var _ shaping.Extenter = (*extenter)(nil)

const (
	simpleGID fonts.GID = iota
	leftExtentGID
	rightExtentGID
	deepGID
	offsetGID
	missingGID
)

var (
	expectedFontExtents = shaping.Bounds{
		Ascent:  fixed.I(int(15)),
		Descent: fixed.I(int(-15)),
		Gap:     fixed.I(int(0)),
	}
	simpleGlyph = shaping.Glyph{
		GlyphInfo: harfbuzz.GlyphInfo{
			Glyph: simpleGID,
		},
		GlyphPosition: harfbuzz.GlyphPosition{
			XAdvance: 10,
			YAdvance: 10,
			XOffset:  0,
			YOffset:  0,
		},
		GlyphExtents: harfbuzz.GlyphExtents{
			Width:    10,
			Height:   -10,
			YBearing: 10,
		},
	}
	leftExtentGlyph = shaping.Glyph{
		GlyphInfo: harfbuzz.GlyphInfo{
			Glyph: leftExtentGID,
		},
		GlyphPosition: harfbuzz.GlyphPosition{
			XAdvance: 5,
			YAdvance: 5,
			XOffset:  0,
			YOffset:  0,
		},
		GlyphExtents: harfbuzz.GlyphExtents{
			Width:    10,
			Height:   -10,
			YBearing: 10,
			XBearing: 5,
		},
	}
	rightExtentGlyph = shaping.Glyph{
		GlyphInfo: harfbuzz.GlyphInfo{
			Glyph: rightExtentGID,
		},
		GlyphPosition: harfbuzz.GlyphPosition{
			XAdvance: 5,
			YAdvance: 5,
			XOffset:  0,
			YOffset:  0,
		},
		GlyphExtents: harfbuzz.GlyphExtents{
			Width:    10,
			Height:   -10,
			YBearing: 10,
			XBearing: 0,
		},
	}
	deepGlyph = shaping.Glyph{
		GlyphInfo: harfbuzz.GlyphInfo{
			Glyph: deepGID,
		},
		GlyphPosition: harfbuzz.GlyphPosition{
			XAdvance: 10,
			YAdvance: 10,
			XOffset:  0,
			YOffset:  0,
		},
		GlyphExtents: harfbuzz.GlyphExtents{
			Width:    10,
			Height:   -10,
			YBearing: 0,
			XBearing: 0,
		},
	}
	offsetGlyph = shaping.Glyph{
		GlyphInfo: harfbuzz.GlyphInfo{
			Glyph: offsetGID,
		},
		GlyphPosition: harfbuzz.GlyphPosition{
			XAdvance: 10,
			YAdvance: 10,
			XOffset:  2,
			YOffset:  2,
		},
		GlyphExtents: harfbuzz.GlyphExtents{
			Width:    10,
			Height:   -10,
			YBearing: 10,
			XBearing: 0,
		},
	}
	missingGlyph = shaping.Glyph{
		GlyphInfo: harfbuzz.GlyphInfo{
			Glyph: missingGID,
		},
	}
)

// TestRecalculate ensures that the Output.Recalculate function correctly
// computes the bounds, advance, and baseline of the output.
func TestRecalculate(t *testing.T) {
	extenter := &extenter{
		extents: map[fonts.GID]harfbuzz.GlyphExtents{
			simpleGID:      simpleGlyph.GlyphExtents,
			leftExtentGID:  leftExtentGlyph.GlyphExtents,
			rightExtentGID: rightExtentGlyph.GlyphExtents,
			deepGID:        deepGlyph.GlyphExtents,
			offsetGID:      offsetGlyph.GlyphExtents,
		},
	}
	type testcase struct {
		Name      string
		Direction di.Direction
		Input     []shaping.Glyph
		Output    shaping.Output
		Error     error
	}
	for _, tc := range []testcase{
		{
			Name: "empty",
			Output: shaping.Output{
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "missing glyph should error",
			Direction: di.DirectionLTR,
			Input:     []shaping.Glyph{missingGlyph},
			Error:     shaping.MissingGlyphError{GID: missingGID},
		},
		{
			Name:      "single simple glyph",
			Direction: di.DirectionLTR,
			Input:     []shaping.Glyph{simpleGlyph},
			Output: shaping.Output{
				Glyphs:  []shaping.Glyph{simpleGlyph},
				Advance: fixed.I(int(simpleGlyph.XAdvance)),
				GlyphBounds: shaping.Bounds{
					Ascent:  fixed.I(int(simpleGlyph.GlyphExtents.YBearing)),
					Descent: fixed.I(0),
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "glyph below baseline",
			Direction: di.DirectionLTR,
			Input:     []shaping.Glyph{simpleGlyph, deepGlyph},
			Output: shaping.Output{
				Glyphs:  []shaping.Glyph{simpleGlyph, deepGlyph},
				Advance: fixed.I(int(simpleGlyph.XAdvance + deepGlyph.XAdvance)),
				GlyphBounds: shaping.Bounds{
					Ascent:  fixed.I(int(simpleGlyph.GlyphExtents.YBearing)),
					Descent: fixed.I(int(deepGlyph.GlyphExtents.YBearing + deepGlyph.GlyphExtents.Height)),
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "single complex glyph",
			Direction: di.DirectionLTR,
			Input:     []shaping.Glyph{offsetGlyph},
			Output: shaping.Output{
				Glyphs:  []shaping.Glyph{offsetGlyph},
				Advance: fixed.I(int(offsetGlyph.XAdvance)),
				GlyphBounds: shaping.Bounds{
					Ascent:  fixed.I(int(offsetGlyph.GlyphExtents.YBearing + offsetGlyph.YOffset)),
					Descent: fixed.I(0),
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "vertical text not supported",
			Direction: di.Direction(harfbuzz.BottomToTop),
			Error:     shaping.UnimplementedDirectionError{},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			output := shaping.Output{Glyphs: tc.Input}
			err := output.Recalculate(tc.Direction, extenter)
			if tc.Error != nil && !errors.As(err, &tc.Error) {
				t.Errorf("expected error of type %T, got %T", tc.Error, err)
			} else if tc.Error == nil && !reflect.DeepEqual(output, tc.Output) {
				t.Errorf("recalculation incorrect: expected %v, got %v", tc.Output, output)
			}
		})
	}
}
