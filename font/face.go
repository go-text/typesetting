package font

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
)

type GID = fonts.GID
type GlyphMask = harfbuzz.GlyphMask

type Face interface {
	fonts.Face // make sure that we can use textlayout/fonts.Face as Face

	// NominalGlyph returns the glyph identifier used to represent the given rune,
	// or false the rune is not supported by the font.
	NominalGlyph(r rune) (GID, bool)
}
