package shaping

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/go-text/di"
	"golang.org/x/image/math/fixed"
)

type Glyph struct {
	harfbuzz.GlyphInfo
	harfbuzz.GlyphPosition
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
}

// GlyphExtenter provides extent information for glyphs.
type GlyphExtenter interface {
	GlyphExtents(fonts.GID) (harfbuzz.GlyphExtents, bool)
}

// Recalculate updates the Bounds, Advance, and Baseline fields in
// the Output to match the current contents of the Glyphs field. This
// operation requires querying extra information from the font as well
// as knowing the direction of the text.
func (o *Output) Recalculate(dir di.Direction, font GlyphExtenter) error {
	var (
		advance      int32
		bearingWidth int32
		tallest      int32
		lowest       int32
	)

	switch dir {
	case di.DirectionLTR, di.DirectionRTL:
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
			if i == 0 {
				// If this is the first glyph, add its left bearing to the
				// output bounds.
				bearingWidth += extents.XBearing
			} else if i == len(o.Glyphs)-1 {
				// If this is the last glyph, add its right bearing to the
				// output bounds.
				bearingWidth += extents.Width - g.XAdvance - extents.XBearing
			}
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

	return nil
}
