// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package shaping

import (
	"bytes"
	"reflect"
	"testing"

	hd "github.com/go-text/typesetting-utils/harfbuzz"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
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
	expectedFontExtents = Bounds{
		Ascent:  fixed.I(15),
		Descent: fixed.I(-15),
		Gap:     fixed.I(0),
	}
	simpleGlyph_ = Glyph{
		GlyphID:  simpleGID,
		XAdvance: fixed.I(10),
		YAdvance: -fixed.I(10),
		Width:    fixed.I(10),
		Height:   -fixed.I(10),
		YBearing: fixed.I(10),
	}
	deepGlyph = Glyph{
		GlyphID:  deepGID,
		XAdvance: fixed.I(10),
		YAdvance: -fixed.I(10),
		XOffset:  -fixed.I(5),
		Width:    fixed.I(10),
		Height:   -fixed.I(10),
		YBearing: fixed.I(0),
	}
	offsetGlyph = Glyph{
		GlyphID:  offsetGID,
		XAdvance: fixed.I(10),
		YAdvance: -fixed.I(10),
		XOffset:  -fixed.I(2),
		YOffset:  fixed.I(2),
		Width:    fixed.I(10),
		Height:   -fixed.I(10),
		YBearing: fixed.I(10),
		XBearing: fixed.I(1),
	}
)

// TestRecalculate ensures that the Output.RecalculateAll function correctly
// computes the bounds, advance, and baseline of the output.
func TestRecalculate(t *testing.T) {
	type testcase struct {
		Name      string
		Direction di.Direction
		Input     []Glyph
		Output    Output
	}
	for _, tc := range []testcase{
		{
			Name: "empty",
			Output: Output{
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "horizontal single simple glyph",
			Direction: di.DirectionLTR,
			Input:     []Glyph{simpleGlyph_},
			Output: Output{
				Glyphs:  []Glyph{simpleGlyph_},
				Advance: simpleGlyph_.XAdvance,
				GlyphBounds: Bounds{
					Ascent:  simpleGlyph_.YBearing,
					Descent: fixed.I(0),
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "horizontal glyph below baseline",
			Direction: di.DirectionLTR,
			Input:     []Glyph{simpleGlyph_, deepGlyph},
			Output: Output{
				Glyphs:  []Glyph{simpleGlyph_, deepGlyph},
				Advance: simpleGlyph_.XAdvance + deepGlyph.XAdvance,
				GlyphBounds: Bounds{
					Ascent:  simpleGlyph_.YBearing,
					Descent: deepGlyph.YBearing + deepGlyph.Height,
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "horizontal single complex glyph",
			Direction: di.DirectionLTR,
			Input:     []Glyph{offsetGlyph},
			Output: Output{
				Glyphs:  []Glyph{offsetGlyph},
				Advance: offsetGlyph.XAdvance,
				GlyphBounds: Bounds{
					Ascent:  offsetGlyph.YBearing + offsetGlyph.YOffset,
					Descent: fixed.I(0),
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "vertical single simple glyph",
			Direction: di.DirectionTTB,
			Input:     []Glyph{simpleGlyph_},
			Output: Output{
				Glyphs:  []Glyph{simpleGlyph_},
				Advance: simpleGlyph_.YAdvance,
				GlyphBounds: Bounds{
					Ascent:  simpleGlyph_.Width,
					Descent: 0,
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "vertical glyph below baseline",
			Direction: di.DirectionTTB,
			Input:     []Glyph{simpleGlyph_, deepGlyph},
			Output: Output{
				Glyphs:  []Glyph{simpleGlyph_, deepGlyph},
				Advance: simpleGlyph_.YAdvance + deepGlyph.YAdvance,
				GlyphBounds: Bounds{
					Ascent:  simpleGlyph_.Width,
					Descent: deepGlyph.XOffset + deepGlyph.XBearing,
				},
				LineBounds: expectedFontExtents,
			},
		},
		{
			Name:      "vertical single complex glyph",
			Direction: di.DirectionTTB,
			Input:     []Glyph{offsetGlyph},
			Output: Output{
				Glyphs:  []Glyph{offsetGlyph},
				Advance: offsetGlyph.YAdvance,
				GlyphBounds: Bounds{
					Ascent:  offsetGlyph.Width + offsetGlyph.XOffset + offsetGlyph.XBearing,
					Descent: offsetGlyph.XOffset + offsetGlyph.XBearing,
				},
				LineBounds: expectedFontExtents,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			output := Output{
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

func TestRotate(t *testing.T) {
	glyphs1 := Output{
		Glyphs: []Glyph{
			simpleGlyph_, deepGlyph, offsetGlyph, deepGlyph, deepGlyph, offsetGlyph,
		},
		Direction: di.DirectionRTL,
	}
	glyphs2 := func() Output {
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

		horiz := shaper.Shape(input)
		return horiz
	}()
	for _, horiz := range []Output{glyphs1, glyphs2} {
		horiz.RecalculateAll()
		vert := horiz
		vert.sideways()
		vert.RecalculateAll()

		tu.Assert(t, vert.Direction.IsVertical())
		tu.Assert(t, vert.Direction.Progression() == horiz.Direction.Progression())
		// test that rotate actually preserve the bounds
		tu.Assert(t, horiz.LineBounds == vert.LineBounds)
		tu.Assert(t, horiz.GlyphBounds == vert.GlyphBounds)
		tu.Assert(t, horiz.Advance == -vert.Advance)
	}
}

func TestConvertUnit(t *testing.T) {
	f := benchEnFace
	tu.Assert(t, f.Upem() == 2048)

	for _, test := range []struct {
		size   fixed.Int26_6
		font   float32
		scaled fixed.Int26_6
	}{
		{fixed.I(100), 2048, fixed.I(100)},
		{fixed.I(100), 1024, fixed.I(50)},
		{fixed.I(100), 204.8, fixed.I(10)},
		{fixed.I(30), 2048, fixed.I(30)},
		{fixed.I(30), 1024, fixed.I(15)},
		{fixed.I(12), 1024, fixed.I(6)},
		{fixed.Int26_6(1<<6 + 20), 1024, fixed.Int26_6(1<<6+20) / 2},
	} {
		o := Output{Size: test.size, Face: f}
		tu.Assert(t, o.FromFontUnit(test.font) == test.scaled)
		tu.Assert(t, o.ToFontUnit(test.scaled) == test.font)
	}
}

func TestLine_AdjustBaseline(t *testing.T) {
	var sideways di.Direction
	sideways.SetSideways(true)

	glyph := func(offset, width int) Glyph { return Glyph{Width: fixed.I(width), XOffset: fixed.I(offset)} }
	tests := []struct {
		l        Line
		ascents  []int
		descents []int
	}{
		{Line{}, []int{}, []int{}},                               // no-op
		{Line{{Direction: di.DirectionLTR}}, []int{0}, []int{0}}, // no-op
		{Line{
			{
				Direction:   di.DirectionTTB,
				Glyphs:      []Glyph{glyph(0, 20), glyph(0, 30)},
				GlyphBounds: Bounds{Ascent: fixed.I(30), Descent: fixed.I(0)},
			},
			{
				Direction:   sideways,
				Glyphs:      []Glyph{glyph(-10, 20), glyph(-10, 40)},
				GlyphBounds: Bounds{Ascent: fixed.I(40), Descent: fixed.I(-10)},
			},
		}, []int{30, 25}, []int{0, -25}}, // no-op
		{Line{
			{
				Direction:   sideways,
				Glyphs:      []Glyph{glyph(-10, 20), glyph(-10, 20)},
				GlyphBounds: Bounds{Ascent: fixed.I(20), Descent: fixed.I(-10)},
			},
			{
				Direction:   di.DirectionTTB,
				Glyphs:      []Glyph{glyph(0, 20), glyph(0, 30)},
				GlyphBounds: Bounds{Ascent: fixed.I(30), Descent: fixed.I(0)},
			},
			{
				Direction:   sideways,
				Glyphs:      []Glyph{glyph(0, 20), glyph(0, 40)},
				GlyphBounds: Bounds{Ascent: fixed.I(40), Descent: fixed.I(0)},
			},
		}, []int{5, 30, 25}, []int{-25, 0, -15}}, // no-op

	}
	for _, tt := range tests {
		tt.l.AdjustBaselines()
		for i, exp := range tt.ascents {
			tu.Assert(t, tt.l[i].GlyphBounds.Ascent == fixed.I(exp))
		}
		for i, exp := range tt.descents {
			tu.Assert(t, tt.l[i].GlyphBounds.Descent == fixed.I(exp))
		}
	}
}
