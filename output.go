// SPDX-License-Identifier: Unlicense OR MIT

package shaping

import (
	"fmt"
	"log"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/go-text/di"
	"golang.org/x/image/math/fixed"
)

type Glyph struct {
	harfbuzz.GlyphInfo
	harfbuzz.GlyphPosition
	harfbuzz.GlyphExtents
}

type Output struct {
	// Advance is the distance the Dot has advanced.
	Advance fixed.Int26_6
	// Baseline is the distance the baseline is from the top.
	Baseline fixed.Int26_6
	// Bounds is the smallest rectangle capable of containing the shaped text.
	Bounds fixed.Rectangle26_6
	// Glyphs are the shaped output text.
	Glyphs []Glyph
	// LineBounds describes the font's suggested line bounding dimensions.
	LineBounds LineBounds
}

// LineBounds provides a font's self-described line dimensions.
type LineBounds struct {
	// MaxAscent is the highest ascent that layout should account for in the
	// given font. This value is typically positive in coordiate systems that
	// grow up.
	MaxAscent fixed.Int26_6
	// MaxDescent is the lowest descent below the baseline that layout should
	// account for in the given font. This value is typically negative in
	// coordinate systems that grow up.
	MaxDescent fixed.Int26_6
	// LineGap is the suggested gap of empty space between lines in the font.
	LineGap fixed.Int26_6
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
		advance      int32
		bearingWidth int32
		tallest      int32
		lowest       int32
		hbDir        harfbuzz.Direction
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
			log.Printf("glyph %d, advance %d, offset %d, width %d, extent %d", g.Glyph, g.XAdvance, g.XOffset, extents.Width, extents.XBearing)
			if i == 0 {
				// If this is the first glyph, add its left bearing to the
				// output bounds.
				bearingWidth += extents.XBearing
			} else if i == len(o.Glyphs)-1 {
				// If this is the last glyph, add its right bearing to the
				// output bounds.
				bearingWidth += extents.Width - g.XAdvance - extents.XBearing
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
	o.Bounds = fixed.Rectangle26_6{
		Max: fixed.Point26_6{
			X: fixed.I(int(advance + bearingWidth)),
			Y: fixed.I(int(tallest - lowest)),
		},
	}
	o.Baseline = fixed.I(int(tallest))

	fontExtents := font.ExtentsForDirection(hbDir)
	o.LineBounds = LineBounds{
		MaxAscent:  fixed.I(int(fontExtents.Ascender)),
		MaxDescent: fixed.I(int(fontExtents.Descender)),
		LineGap:    fixed.I(int(fontExtents.LineGap)),
	}

	return nil
}
