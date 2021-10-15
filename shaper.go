package shaping

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/go-text/di"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type Shaper interface {
	// Shape takes an Input and shapes it into the Output.
	Shape(Input) Output
}

// scale returns x divided by unitsPerEm, rounded to the nearest fixed.Int26_6
// value (1/64th of a pixel). Borrowed from sfnt package.
func scale(x fixed.Int26_6, unitsPerEm sfnt.Units) fixed.Int26_6 {
	if x >= 0 {
		x += fixed.Int26_6(unitsPerEm) / 2
	} else {
		x -= fixed.Int26_6(unitsPerEm) / 2
	}
	return x / fixed.Int26_6(unitsPerEm)
}

// scalePosition using specific ppem and upem values.
func scalePosition(gp harfbuzz.GlyphPosition, ppem fixed.Int26_6, upem sfnt.Units) harfbuzz.GlyphPosition {
	gp.XAdvance = int32(scale(fixed.I(int(gp.XAdvance)).Mul(ppem), upem).Round())
	gp.YAdvance = int32(scale(fixed.I(int(gp.YAdvance)).Mul(ppem), upem).Round())
	gp.XOffset = int32(scale(fixed.I(int(gp.XOffset)).Mul(ppem), upem).Round())
	gp.YOffset = int32(scale(fixed.I(int(gp.YOffset)).Mul(ppem), upem).Round())
	return gp
}

// TODO: can we just make Input and Output into structs?
// Interfaces encapsulate varying behavior well, but these
// types have no behavior. They're just structured data.
// It feels silly to wrap them in accessors like this.
type output struct {
	DotAdvance    fixed.Int26_6
	TopToBaseline fixed.Int26_6
	TextBounds    fixed.Rectangle26_6
	Glyphs        []Glyph
}

func (o output) Advance() fixed.Int26_6 {
	return o.DotAdvance
}

func (o output) Baseline() fixed.Int26_6 {
	return o.TopToBaseline
}

func (o output) Bounds() fixed.Rectangle26_6 {
	return o.TextBounds
}

func (o output) Length() int {
	return len(o.Glyphs)
}

func (o output) Index(i int) Glyph {
	return o.Glyphs[i]
}

// MissingGlyphError indicates that the font used in shaping did not
// have a glyph needed to complete the shaping.
type MissingGlyphError struct {
	fonts.GID
}

func (m MissingGlyphError) Error() string {
	return fmt.Sprintf("missing glyph with id %d", m.GID)
}

// Shape turns an input into an output.
func Shape(input Input) (Output, error) {
	// Prepare to shape the text.
	// TODO: maybe reuse these buffers for performance?
	buf := harfbuzz.NewBuffer()
	runes, start, end := input.Text()
	buf.AddRunes(runes, start, end-start)
	// TODO: handle vertical text?
	switch input.Direction() {
	case di.DirectionLTR:
		buf.Props.Direction = harfbuzz.LeftToRight
	case di.DirectionRTL:
		buf.Props.Direction = harfbuzz.RightToLeft
	}
	buf.Props.Language = input.Language()
	buf.Props.Script = input.Script()
	// TODO: figure out what (if anything) to do if this type assertion fails.
	font := harfbuzz.NewFont(input.Face().(harfbuzz.Face))
	upem := sfnt.Units(font.Face().Upem())
	ppem := input.Size()

	// Actually use harfbuzz to shape the text.
	buf.Shape(font, nil)

	// Convert the shaped text into an Output.
	glyphs := make([]Glyph, len(buf.Info))
	for i := range glyphs {
		glyphs[i] = Glyph{
			GlyphInfo:     buf.Info[i],
			GlyphPosition: scalePosition(buf.Pos[i], ppem, upem),
		}
	}
	out := output{
		Glyphs: glyphs,
	}
	var (
		advance  int32
		baseline int32
		tallest  int32
	)

	for i := range out.Glyphs {
		g := &out.Glyphs[i]
		switch input.Direction() {
		case di.DirectionLTR, di.DirectionRTL:
			advance += g.GlyphPosition.XAdvance
			// Look up glyph id in font to get baseline info.
			// TODO: this seems like it shouldn't be necessary.
			// Not sure where else to get this info though.
			extents, ok := font.GlyphExtents(g.GlyphInfo.Glyph)
			if !ok {
				// TODO: can this error happen? Will harfbuzz return a
				// GID for a glyph that isn't in the font?
				return nil, MissingGlyphError{GID: g.GlyphInfo.Glyph}
			}
			if h := -extents.Height; h > tallest {
				tallest = h
			}
			if extents.YBearing > baseline {
				baseline = extents.YBearing
			}
		}
	}
	out.DotAdvance = fixed.I(int(advance))
	out.TextBounds = fixed.Rectangle26_6{
		Max: fixed.Point26_6{
			X: out.DotAdvance,
			Y: scale(fixed.I(int(tallest)).Mul(ppem), upem),
		},
	}
	out.TopToBaseline = scale(fixed.Int26_6(baseline).Mul(ppem), upem)

	return out, nil
}
