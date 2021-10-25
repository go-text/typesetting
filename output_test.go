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

// Ensure that *extenter is a shaping.GlyphExtenter
var _ shaping.GlyphExtenter = (*extenter)(nil)

const (
	simpleGID fonts.GID = iota
)

var simpleGlyph = shaping.Glyph{
	GlyphInfo: harfbuzz.GlyphInfo{
		Glyph: simpleGID,
	},
	GlyphPosition: harfbuzz.GlyphPosition{
		XAdvance: 10,
		YAdvance: 10,
		XOffset:  0,
		YOffset:  0,
	},
}
var simpleGlyphExtents = harfbuzz.GlyphExtents{
	Width:    10,
	Height:   10,
	YBearing: 10,
}

// TestRecalculate ensures that the Output.Recalculate function correctly
// computes the bounds, advance, and baseline of the output.
func TestRecalculate(t *testing.T) {
	extenter := &extenter{
		extents: map[fonts.GID]harfbuzz.GlyphExtents{
			simpleGID: simpleGlyphExtents,
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
		},
		{
			Name:      "single simple glyph",
			Direction: di.DirectionLTR,
			Input:     []shaping.Glyph{simpleGlyph},
			Output: shaping.Output{
				Glyphs:   []shaping.Glyph{simpleGlyph},
				Advance:  fixed.I(int(simpleGlyph.XAdvance)),
				Baseline: fixed.I(int(simpleGlyphExtents.Height)),
				Bounds: fixed.Rectangle26_6{
					Max: fixed.Point26_6{
						X: fixed.I(int(simpleGlyphExtents.Width)),
						Y: fixed.I(int(simpleGlyphExtents.Height)),
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			output := shaping.Output{Glyphs: tc.Input}
			err := output.Recalculate(tc.Direction, extenter)
			if tc.Error != nil && !errors.As(err, &tc.Error) {
				t.Errorf("expected error of type %T, got %T", tc.Error, err)
			}
			if !reflect.DeepEqual(output, tc.Output) {
				t.Errorf("recalculation incorrect: expected %v, got %v", tc.Output, output)
			}
		})
	}
}
