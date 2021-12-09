// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

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

const (
	simpleGID fonts.GID = iota
	leftExtentGID
	rightExtentGID
	deepGID
	offsetGID
)

var (
	expectedFontExtents = shaping.Bounds{
		Ascent:  fixed.I(int(15)),
		Descent: fixed.I(int(-15)),
		Gap:     fixed.I(int(0)),
	}
	simpleGlyph = shaping.Glyph{
		GlyphID:    simpleGID,
		XAdvance: fixed.I(int(10)),
		YAdvance: fixed.I(int(10)),
		XOffset:  fixed.I(int(0)),
		YOffset:  fixed.I(int(0)),
		Width:    fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(10)),
	}
	leftExtentGlyph = shaping.Glyph{
		GlyphID:    leftExtentGID,
		XAdvance: fixed.I(int(5)),
		YAdvance: fixed.I(int(5)),
		XOffset:  fixed.I(int(0)),
		YOffset:  fixed.I(int(0)),
		Width:    fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(10)),
		XBearing: fixed.I(int(5)),
	}
	rightExtentGlyph = shaping.Glyph{
		GlyphID:    rightExtentGID,
		XAdvance: fixed.I(int(5)),
		YAdvance: fixed.I(int(5)),
		XOffset:  fixed.I(int(0)),
		YOffset:  fixed.I(int(0)),
		Width:    fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(10)),
		XBearing: fixed.I(int(0)),
	}
	deepGlyph = shaping.Glyph{
		GlyphID:    deepGID,
		XAdvance: fixed.I(int(10)),
		YAdvance: fixed.I(int(10)),
		XOffset:  fixed.I(int(0)),
		YOffset:  fixed.I(int(0)),
		Width:    fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(0)),
		XBearing: fixed.I(int(0)),
	}
	offsetGlyph = shaping.Glyph{
		GlyphID:    offsetGID,
		XAdvance: fixed.I(int(10)),
		YAdvance: fixed.I(int(10)),
		XOffset:  fixed.I(int(2)),
		YOffset:  fixed.I(int(2)),
		Width:    fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(10)),
		XBearing: fixed.I(int(0)),
	}
)

// TestRecalculate ensures that the Output.Recalculate function correctly
// computes the bounds, advance, and baseline of the output.
func TestRecalculate(t *testing.T) {
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
			Name:      "single simple glyph",
			Direction: di.DirectionLTR,
			Input:     []shaping.Glyph{simpleGlyph},
			Output: shaping.Output{
				Glyphs:  []shaping.Glyph{simpleGlyph},
				Advance: simpleGlyph.XAdvance,
				GlyphBounds: shaping.Bounds{
					Ascent:  simpleGlyph.YBearing,
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
				Advance: simpleGlyph.XAdvance + deepGlyph.XAdvance,
				GlyphBounds: shaping.Bounds{
					Ascent:  simpleGlyph.YBearing,
					Descent: deepGlyph.YBearing + deepGlyph.Height,
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
				Advance: offsetGlyph.XAdvance,
				GlyphBounds: shaping.Bounds{
					Ascent:  offsetGlyph.YBearing + offsetGlyph.YOffset,
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
			output := shaping.Output{
				Glyphs:     tc.Input,
				LineBounds: expectedFontExtents,
			}
			err := output.RecalculateAll(tc.Direction)
			if tc.Error != nil && !errors.As(err, &tc.Error) {
				t.Errorf("expected error of type %T, got %T", tc.Error, err)
			} else if tc.Error == nil && !reflect.DeepEqual(output, tc.Output) {
				t.Errorf("recalculation incorrect: expected %v, got %v", tc.Output, output)
			}
		})
	}
}
