// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package shaping

import (
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"golang.org/x/image/math/fixed"
)

// Glyph describes the attributes of a single glyph from a single
// font face in a shaped output.
type Glyph struct {
	// Width is the width of the glyph content,
	// expressed as a distance from the [XBearing],
	// typically positive
	Width fixed.Int26_6
	// Height is the height of the glyph content,
	// expressed as a distance from the [YBearing],
	// typically negative
	Height fixed.Int26_6
	// XBearing is the distance between the dot (with offset applied) and
	// the glyph content, typically positive
	XBearing fixed.Int26_6
	// YBearing is the distance between the dot (with offset applied) and
	// the top of the glyph content, typically positive
	YBearing fixed.Int26_6
	// XAdvance is the distance between the current dot (without offset applied) and the next dot.
	// It is typically positive for horizontal text, and always zero for vertical text.
	XAdvance fixed.Int26_6
	// YAdvance is the distance between the current dot (without offset applied) and the next dot.
	// It is typically negative for vertical text, and always zero for horizontal text.
	YAdvance fixed.Int26_6

	// Offsets to be applied to the dot before actually drawing
	// the glyph.
	XOffset, YOffset fixed.Int26_6

	// ClusterIndex is the lowest rune index of all runes shaped into
	// this glyph cluster. All glyphs sharing the same cluster value
	// are part of the same cluster and will have identical RuneCount
	// and GlyphCount fields.
	ClusterIndex int
	// RuneCount is the number of input runes shaped into this output
	// glyph cluster.
	RuneCount int
	// GlyphCount is the number of glyphs in this output glyph cluster.
	GlyphCount int
	GlyphID    font.GID
	Mask       font.GlyphMask
}

// LeftSideBearing returns the distance from the glyph's X origin to
// its leftmost edge. This value can be negative if the glyph extends
// across the origin.
func (g Glyph) LeftSideBearing() fixed.Int26_6 {
	return g.XBearing
}

// RightSideBearing returns the distance from the glyph's right edge to
// the edge of the glyph's advance. This value can be negative if the glyph's
// right edge is after the end of its advance.
func (g Glyph) RightSideBearing() fixed.Int26_6 {
	return g.XAdvance - g.Width - g.XBearing
}

// Bounds describes the minor-axis bounds of a line of text. In a LTR or RTL
// layout, it describes the vertical axis. In a BTT or TTB layout, it describes
// the horizontal.
//
// For horizontal text:
//
//   - Ascent      GLYPH
//     |             GLYPH
//     |             GLYPH
//     |             GLYPH
//     |             GLYPH
//   - Baseline    GLYPH
//     |             GLYPH
//     |             GLYPH
//     |             GLYPH
//   - Descent     GLYPH
//     |
//   - Gap
//
// For vertical text:
//
//	Descent ------- Baseline --------------- Ascent --- Gap
//		|				| 						|		|
//		GLYPH		  GLYPH					GLYPH
//			GLYPH GLYPH		GLYPH GLYPH GLYPH
type Bounds struct {
	// Ascent is the maximum ascent away from the baseline. This value is typically
	// positive in coordiate systems that grow up.
	Ascent fixed.Int26_6
	// Descent is the maximum descent away from the baseline. This value is typically
	// negative in coordinate systems that grow up.
	Descent fixed.Int26_6
	// Gap is the height of empty pixels between lines. This value is typically positive
	// in coordinate systems that grow up.
	Gap fixed.Int26_6
}

// LineThickness returns the thickness of a line of text described by b,
// that is its height for horizontal text, its width for vertical text.
func (b Bounds) LineThickness() fixed.Int26_6 {
	return b.Ascent - b.Descent + b.Gap
}

// Output describes the dimensions and content of shaped text.
type Output struct {
	// Advance is the distance the Dot has advanced.
	Advance fixed.Int26_6
	// Size is copied from the shaping.Input.Size that produced this Output.
	Size fixed.Int26_6
	// Glyphs are the shaped output text.
	Glyphs []Glyph
	// LineBounds describes the font's suggested line bounding dimensions. The
	// dimensions described should contain any glyphs from the given font.
	LineBounds Bounds
	// GlyphBounds describes a tight bounding box on the specific glyphs contained
	// within this output. The dimensions may not be sufficient to contain all
	// glyphs within the chosen font.
	GlyphBounds Bounds

	// Direction is the direction used to shape the text,
	// as provided in the Input.
	Direction di.Direction

	// Runes describes the runes this output represents from the input text.
	Runes Range

	// Face is the font face that this output is rendered in. This is needed in
	// the output in order to render each run in a multi-font sequence in the
	// correct font.
	Face font.Face
}

// RecomputeAdvance updates only the Advance field based on the current
// contents of the Glyphs field. It is faster than RecalculateAll(),
// and can be used to speed up line wrapping logic.
func (o *Output) RecomputeAdvance() {
	advance := fixed.Int26_6(0)
	if o.Direction.IsVertical() {
		for _, g := range o.Glyphs {
			advance += g.YAdvance
		}
	} else { // horizontal
		for _, g := range o.Glyphs {
			advance += g.XAdvance
		}
	}
	o.Advance = advance
}

// advanceSpaceAware adjust the value in [Advance]
// if a white space character ends the run.
// TODO: should we take into account multiple spaces ?
func (o *Output) advanceSpaceAware() fixed.Int26_6 {
	L := len(o.Glyphs)
	if L == 0 {
		return o.Advance
	}

	// adjust the last to account for spaces
	if o.Direction.IsVertical() {
		if g := o.Glyphs[L-1]; g.Height == 0 {
			return o.Advance - g.YAdvance
		}
	} else { // horizontal
		if g := o.Glyphs[L-1]; g.Width == 0 {
			return o.Advance - g.XAdvance
		}
	}

	return o.Advance
}

// RecalculateAll updates the all other fields of the Output
// to match the current contents of the Glyphs field.
// This method will fail with UnimplementedDirectionError if the Output
// direction is unimplemented.
func (o *Output) RecalculateAll() {
	var (
		advance fixed.Int26_6
		tallest fixed.Int26_6
		lowest  fixed.Int26_6
	)

	if o.Direction.IsVertical() {
		for i := range o.Glyphs {
			g := &o.Glyphs[i]
			advance += g.YAdvance
			depth := g.XOffset + g.XBearing // start of the glyph
			if depth < lowest {
				lowest = depth
			}
			height := g.XOffset + g.Width // end of the glyph
			if height > tallest {
				tallest = height
			}
		}
	} else { // horizontal
		for i := range o.Glyphs {
			g := &o.Glyphs[i]
			advance += g.XAdvance
			height := g.YBearing + g.YOffset
			if height > tallest {
				tallest = height
			}
			depth := height + g.Height
			if depth < lowest {
				lowest = depth
			}
		}
	}
	o.Advance = advance
	o.GlyphBounds = Bounds{
		Ascent:  tallest,
		Descent: lowest,
	}
}
