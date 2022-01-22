package fontscan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/bitmap"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fonts/type1"
	"github.com/go-text/typesetting/font"
)

// footprint is a condensed summary of the main information
// about a font, serving as a lightweight surrogate
// for the original font file.
type footprint struct {
	// Location stores the adress of the font resource.
	Location Location

	// Family is the general nature of the font, like
	// "Arial"
	Family string

	// Runes is the set of runes supported by the font.
	Runes runeSet

	// Aspect precises the visual characteristics
	// of the font among a family, like "Bold Italic"
	Aspect Aspect

	// Format provides the format to be used
	// to load and create the associated font.Face.
	Format fontFormat
}

func newFootprintFromDescriptor(fd fonts.FontDescriptor, format fontFormat) (out footprint, err error) {
	cmap, err := fd.LoadCmap() // load the cmap...
	if err != nil {
		return footprint{}, err
	}
	out.Runes = newRuneSetFromCmap(cmap) // ... and build the corresponding rune set

	out.Family = fd.Family() // load the family

	out.Aspect = newAspectFromDescriptor(fd)

	// register the correct format
	out.Format = format

	return out, nil
}

// serializeTo serialize the Footprint in binary format,
// by appending to `dst` and returning the slice
func (as footprint) serializeTo(dst []byte) []byte {
	dst = append(dst, serializeString(as.Location.File)...)

	var buffer [4]byte
	binary.BigEndian.PutUint16(buffer[:], as.Location.Index)
	binary.BigEndian.PutUint16(buffer[2:], as.Location.Instance)
	dst = append(dst, buffer[:]...)

	dst = append(dst, serializeString(as.Family)...)
	dst = append(dst, as.Runes.serialize()...)
	dst = append(dst, as.Aspect.serialize()...)
	dst = append(dst, byte(as.Format))
	return dst
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func (as *footprint) deserializeFrom(data []byte) (int, error) {
	n, err := deserializeString(&as.Location.File, data)
	if err != nil {
		return 0, err
	}
	if len(data) < n+4 {
		return 0, errors.New("invalid Location (EOF)")
	}
	as.Location.Index = binary.BigEndian.Uint16(data[n:])
	as.Location.Instance = binary.BigEndian.Uint16(data[n+2:])
	n += 4

	read, err := deserializeString(&as.Family, data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = as.Runes.deserializeFrom(data[n:])
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
	as.Format = fontFormat(data[n])

	return n + 1, nil
}

// loadFromDisk assume the footprint location refers to the file system
func (fp *footprint) loadFromDisk() (font.Face, error) {
	location := fp.Location

	file, err := os.Open(location.File)
	if err != nil {
		return nil, err
	}

	faces, err := fp.Format.Loader()(file)
	if err != nil {
		return nil, err
	}

	if index := int(location.Index); len(faces) <= index {
		// this should only happen if the font file as changed
		// since the last scan (very unlikely)
		return nil, fmt.Errorf("invalid font index in collection: %d >= %d", index, len(faces))
	}

	return faces[location.Index], nil
}

// fontFormat identifies the format of a font file.
type fontFormat uint8

const (
	_          fontFormat = iota // unsupported
	openType                     // .ttf, .ttc, .otf, .otc, .woff
	pcf                          // Bitmap fonts (.pcf)
	adobeType1                   // Adobe Type1 fonts (.pfb)
)

func (ff fontFormat) String() string {
	switch ff {
	case openType:
		return "OpenType"
	case pcf:
		return "PCF"
	case adobeType1:
		return "Type1"
	default:
		return fmt.Sprintf("<format %d>", ff)
	}
}

// Loader returns the loader to use to open a font resource with
// this format.
func (ff fontFormat) Loader() fonts.FontLoader {
	switch ff {
	case openType:
		return truetype.Load
	case pcf:
		return bitmap.Load
	case adobeType1:
		return type1.Load
	default:
		return nil
	}
}
