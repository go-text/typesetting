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
	runes, start, end := input.Text, input.RunStart, input.RunEnd
	buf.AddRunes(runes, start, end-start)
	// TODO: handle vertical text?
	switch input.Direction {
	case di.DirectionLTR:
		buf.Props.Direction = harfbuzz.LeftToRight
	case di.DirectionRTL:
		buf.Props.Direction = harfbuzz.RightToLeft
	}
	buf.Props.Language = input.Language
	buf.Props.Script = input.Script
	// TODO: figure out what (if anything) to do if this type assertion fails.
	font := harfbuzz.NewFont(input.Face.(harfbuzz.Face))
	font.XScale = int32(input.Size.Ceil())
	font.YScale = font.XScale

	// Actually use harfbuzz to shape the text.
	buf.Shape(font, nil)

	// Convert the shaped text into an Output.
	glyphs := make([]Glyph, len(buf.Info))
	for i := range glyphs {
		glyphs[i] = Glyph{
			GlyphInfo:     buf.Info[i],
			GlyphPosition: buf.Pos[i],
		}
	}
	out := Output{
		Glyphs: glyphs,
	}
	return out, out.Recalculate(input.Direction, font)
}
