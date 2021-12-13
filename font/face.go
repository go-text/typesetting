package font

import "github.com/benoitkugler/textlayout/fonts"

// make sure that we can use textlayout/fonts.Face as Face
var _ Face = (fonts.Face)(nil)

type GID = fonts.GID

type Face interface {
	// NominalGlyph returns the glyph identifier used to represent the given rune,
	// or false the rune is not supported by the font.
	NominalGlyph(r rune) (GID, bool)
}
