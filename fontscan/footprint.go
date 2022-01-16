package fontscan

import (
	"errors"
	"io"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/bitmap"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fonts/type1"
)

// Footprint is a condensed summary of the main information
// about a font, serving as a lightweight surrogate
// for the original font file.
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

	// Format provides the format to be used
	// to load and create the associated font.Face.
	Format Format
}

// serializeTo serialize the Footprint in binary format
// TODO: handle the Location field
func (as Footprint) serializeTo(w io.Writer) error {
	buffer := serializeString(as.Family)
	buffer = append(buffer, as.Runes.serialize()...)
	buffer = append(buffer, as.Aspect.serialize()...)
	buffer = append(buffer, byte(as.Format))

	_, err := w.Write(buffer[:])
	return err
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
// TODO: handle the Location field
func (as *Footprint) deserializeFrom(data []byte) (int, error) {
	n, err := deserializeString(&as.Family, data)
	if err != nil {
		return 0, err
	}
	read, err := as.Runes.deserializeFrom(data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = as.Aspect.deserializeFrom(data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	if len(data[n:]) < 1 {
		return 0, errors.New("invalid Format (EOF)")
	}
	as.Format = Format(data[n])

	return n + 1, nil
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
