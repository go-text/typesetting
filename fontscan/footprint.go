package fontscan

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/bitmap"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fonts/type1"
)

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
	Aspect Aspect

	Format Format
}

// Format identifies the format of a font file.
type Format uint8

const (
	_        Format = iota // unsupported
	OpenType               // .ttf, .ttc, .otf, .otc, .woff
	PCF                    // Bitmap fonts (.pcf)
	Type1                  // Adobe Type1 fonts (.pfb)
)

// Loader returns the loader to use to open a font resource with
// this format.
func (ff Format) Loader() fonts.FontLoader {
	switch ff {
	case OpenType:
		return truetype.Load
	case PCF:
		return bitmap.Load
	case Type1:
		return type1.Load
	default:
		return nil
	}
}
