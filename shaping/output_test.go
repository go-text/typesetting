// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package shaping_test

import (
	"reflect"
	"testing"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

const (
	simpleGID font.GID = iota
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
		GlyphID:  simpleGID,
		XAdvance: fixed.I(int(10)),
		YAdvance: fixed.I(int(10)),
		XOffset:  fixed.I(int(0)),
		YOffset:  fixed.I(int(0)),
		Width:    fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(10)),
	}
	deepGlyph = shaping.Glyph{
		GlyphID:  deepGID,
		XAdvance: fixed.I(int(10)),
		YAdvance: fixed.I(int(10)),
		XOffset:  fixed.I(int(0)),
		YOffset:  fixed.I(int(0)),
		Width:    -fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(0)),
		XBearing: fixed.I(int(0)),
	}
	offsetGlyph = shaping.Glyph{
		GlyphID:  offsetGID,
		XAdvance: fixed.I(int(10)),
		YAdvance: fixed.I(int(10)),
		XOffset:  fixed.I(int(2)),
		YOffset:  fixed.I(int(2)),
		Width:    -fixed.I(int(10)),
		Height:   -fixed.I(int(10)),
		YBearing: fixed.I(int(10)),
		XBearing: fixed.I(int(10)),
	}
)

// TestRecalculate ensures that the Output.RecalculateAll function correctly
// computes the bounds, advance, and baseline of the output.
func TestRecalculate(t *testing.T) {
	type testcase struct {
		Name      string
		Direction di.Direction
		Input     []shaping.Glyph
		Output    shaping.Output
	}
	for _, tc := range []testcase{
		{
			Name: "empty",
			Output: shaping.Output{
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "horizontal single simple glyph",
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
			Name:      "horizontal glyph below baseline",
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
			Name:      "horizontal single complex glyph",
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
			Name:      "vertical single simple glyph",
			Direction: di.DirectionTTB,
			Input:     []shaping.Glyph{simpleGlyph},
			Output: shaping.Output{
				Glyphs:  []shaping.Glyph{simpleGlyph},
				Advance: simpleGlyph.YAdvance,
				GlyphBounds: shaping.Bounds{
					Ascent:  simpleGlyph.XBearing,
					Descent: fixed.I(0),
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "vertical glyph below baseline",
			Direction: di.DirectionTTB,
			Input:     []shaping.Glyph{simpleGlyph, deepGlyph},
			Output: shaping.Output{
				Glyphs:  []shaping.Glyph{simpleGlyph, deepGlyph},
				Advance: simpleGlyph.YAdvance + deepGlyph.YAdvance,
				GlyphBounds: shaping.Bounds{
					Ascent:  simpleGlyph.XBearing,
					Descent: deepGlyph.XBearing + deepGlyph.Width,
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "vertical single complex glyph",
			Direction: di.DirectionTTB,
			Input:     []shaping.Glyph{offsetGlyph},
			Output: shaping.Output{
				Glyphs:  []shaping.Glyph{offsetGlyph},
				Advance: offsetGlyph.YAdvance,
				GlyphBounds: shaping.Bounds{
					Ascent:  offsetGlyph.XBearing + offsetGlyph.XOffset,
					Descent: fixed.I(0),
				},
				LineBounds: expectedFontExtents,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			output := shaping.Output{
				Glyphs:     tc.Input,
				LineBounds: expectedFontExtents,
				Direction:  tc.Direction,
			}
			output.RecalculateAll()
			if output.Advance != tc.Output.Advance {
				t.Errorf("advance mismatch, expected %s, got %s", tc.Output.Advance, output.Advance)
			}
			if !reflect.DeepEqual(output.Glyphs, tc.Output.Glyphs) {
				t.Errorf("glyphs mismatch: expected %v, got %v", tc.Output.Glyphs, output.Glyphs)
			}
			if output.LineBounds != tc.Output.LineBounds {
				t.Errorf("line bounds mismatch, expected %#+v, got %#+v", tc.Output.LineBounds, output.LineBounds)
			}
			if output.GlyphBounds != tc.Output.GlyphBounds {
				t.Errorf("glyph bounds mismatch, expected %#+v, got %#+v", tc.Output.GlyphBounds, output.GlyphBounds)
			}
			if output.Runes != tc.Output.Runes {
				t.Errorf("runes mismatch, expected %#+v, got %#+v", tc.Output.Runes, output.Runes)
			}
		})
	}
}
