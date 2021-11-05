// SPDX-License-Identifier: Unlicense OR MIT

package shaping

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/go-text/di"
	"golang.org/x/image/math/fixed"
)

// Glyph describes the attributes of a single glyph from a single
// font face in a shaped output.
type Glyph struct {
	harfbuzz.GlyphInfo
	harfbuzz.GlyphPosition
	harfbuzz.GlyphExtents
}

// LeftSideBearing returns the distance from the glyph's X origin to
// its leftmost edge. This value can be negative if the glyph extends
// across the origin.
func (g Glyph) LeftSideBearing() int32 {
	return g.GlyphExtents.XBearing
}

// RightSideBearing returns the distance from the glyph's right edge to
// the edge of the glyph's advance. This value can be negative if the glyph's
// right edge is before the end of its advance.
func (g Glyph) RightSideBearing() int32 {
	return g.GlyphPosition.XAdvance - g.GlyphExtents.Width - g.GlyphExtents.XBearing
}

// Bounds describes the minor-axis bounds of a line of text. In a LTR or RTL
// layout, it describes the vertical axis. In a BTT or TTB layout, it describes
// the horizontal.
//
// For horizontal text:
//
//     - Ascent      GLYPH
//     |             GLYPH
//     |             GLYPH
//     |             GLYPH
//     |             GLYPH
//     - Baseline    GLYPH
//     |             GLYPH
//     |             GLYPH
//     |             GLYPH
//     - Descent     GLYPH
//     |
//     - Gap
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

// LineHeight returns the height of a horizontal line of text described by b.
func (b Bounds) LineHeight() fixed.Int26_6 {
	return b.Ascent - b.Descent + b.Gap
}

// Output describes the dimensions and content of shaped text.
type Output struct {
	// Advance is the distance the Dot has advanced.
	Advance fixed.Int26_6
	// Glyphs are the shaped output text.
	Glyphs []Glyph
	// LineBounds describes the font's suggested line bounding dimensions. The
	// dimensions described should contain any glyphs from the given font.
	LineBounds Bounds
	// GlyphBounds describes a tight bounding box on the specific glyphs contained
	// within this output. The dimensions may not be sufficient to contain all
	// glyphs within the chosen font.
	GlyphBounds Bounds
}

// Extenter provides extent information for glyphs and lines of text
// in a given font.
type Extenter interface {
	GlyphExtents(fonts.GID) (harfbuzz.GlyphExtents, bool)
	ExtentsForDirection(harfbuzz.Direction) fonts.FontExtents
}

// UnimplementedDirectionError is returned when a function does not support the
// provided layout direction yet.
type UnimplementedDirectionError struct {
	Direction di.Direction
}

// Error formats the error into a string message.
func (u UnimplementedDirectionError) Error() string {
	return fmt.Sprintf("support for text direction %v is not implemented yet", u.Direction)
}

// Recalculate updates the Bounds, Advance, and Baseline fields in
// the Output to match the current contents of the Glyphs field. This
// operation requires querying extra information from the font as well
// as knowing the direction of the text.
//
// This method will fail with UnimplementedDirectionError if the provided
// direction is unimplemented, and with MissingGlyphError if the provided
// Extenter cannot resolve a required glyph.
func (o *Output) Recalculate(dir di.Direction, font Extenter) error {
	var (
		advance int32
		tallest int32
		lowest  int32
		hbDir   harfbuzz.Direction
	)

	switch dir {
	default:
		return UnimplementedDirectionError{Direction: dir}
	case di.DirectionLTR, di.DirectionRTL:
		hbDir = harfbuzz.LeftToRight
		if dir == di.DirectionRTL {
			hbDir = harfbuzz.RightToLeft
		}
		for i := range o.Glyphs {
			g := &o.Glyphs[i]
			advance += g.GlyphPosition.XAdvance
			// Look up glyph id in font to get baseline info.
			// TODO: this seems like it shouldn't be necessary.
			// Not sure where else to get this info though.
			extents, ok := font.GlyphExtents(g.GlyphInfo.Glyph)
			if !ok {
				// TODO: can this error happen? Will harfbuzz return a
				// GID for a glyph that isn't in the font?
				return MissingGlyphError{GID: g.GlyphInfo.Glyph}
			}
			g.GlyphExtents = extents
			height := extents.YBearing + g.YOffset
			if height > tallest {
				tallest = height
			}
			depth := height + extents.Height
			if depth < lowest {
				lowest = depth
			}
		}
	}
	o.Advance = fixed.I(int(advance))
	o.GlyphBounds = Bounds{
		Ascent:  fixed.I(int(tallest)),
		Descent: fixed.I(int(lowest)),
	}

	fontExtents := font.ExtentsForDirection(hbDir)
	o.LineBounds = Bounds{
		Ascent:  fixed.I(int(fontExtents.Ascender)),
		Descent: fixed.I(int(fontExtents.Descender)),
		Gap:     fixed.I(int(fontExtents.LineGap)),
	}

	return nil
}
