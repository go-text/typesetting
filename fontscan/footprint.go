package fontscan

import "github.com/benoitkugler/textlayout/fonts"

type Footprint struct {
	// Location stores the adress of the font file.
	Location fonts.FaceID

	// Family is the general nature of the font, like
	// "Arial"
	Family string

	// Runes is the set of runes supported by the font.
	Runes RuneSet

	// Aspect precises the visual characteristics
	// of the font among a family, like "Bold Italic"
	Aspect
}
