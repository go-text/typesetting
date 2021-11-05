// SPDX-License-Identifier: Unlicense OR MIT

package shaping

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/go-text/di"
	"golang.org/x/image/math/fixed"
)

type Shaper interface {
	// Shape takes an Input and shapes it into the Output.
	Shape(Input) Output
}

// MissingGlyphError indicates that the font used in shaping did not
// have a glyph needed to complete the shaping.
type MissingGlyphError struct {
	fonts.GID
}

func (m MissingGlyphError) Error() string {
	return fmt.Sprintf("missing glyph with id %d", m.GID)
}

// InvalidRunError represents an invalid run of text, either because
// the end is before the start or because start or end is greater
// than the length.
type InvalidRunError struct {
	RunStart, RunEnd, TextLength int
}

func (i InvalidRunError) Error() string {
	return fmt.Sprintf("run from %d to %d is not valid for text len %d", i.RunStart, i.RunEnd, i.TextLength)
}

// Shape turns an input into an output.
func Shape(input Input) (Output, error) {
	// Prepare to shape the text.
	// TODO: maybe reuse these buffers for performance?
	buf := harfbuzz.NewBuffer()
	runes, start, end := input.Text, input.RunStart, input.RunEnd
	if end < start {
		return Output{}, InvalidRunError{RunStart: start, RunEnd: end, TextLength: len(input.Text)}
	}
	buf.AddRunes(runes, start, end-start)
	// TODO: handle vertical text?
	switch input.Direction {
	case di.DirectionLTR:
		buf.Props.Direction = harfbuzz.LeftToRight
	case di.DirectionRTL:
		buf.Props.Direction = harfbuzz.RightToLeft
	default:
		return Output{}, UnimplementedDirectionError{
			Direction: input.Direction,
		}
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
		g := glyphs[i].Glyph
		extents, ok := font.GlyphExtents(g)
		if !ok {
			// TODO: can this error happen? Will harfbuzz return a
			// GID for a glyph that isn't in the font?
			return Output{}, MissingGlyphError{GID: g}
		}
		glyphs[i].GlyphExtents = extents
	}
	out := Output{
		Glyphs: glyphs,
	}
	fontExtents := font.ExtentsForDirection(buf.Props.Direction)
	out.LineBounds = Bounds{
		Ascent:  fixed.I(int(fontExtents.Ascender)),
		Descent: fixed.I(int(fontExtents.Descender)),
		Gap:     fixed.I(int(fontExtents.LineGap)),
	}
	return out, out.RecalculateAll(input.Direction)
}
