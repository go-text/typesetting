package fontscan

import (
	"encoding/binary"
	"errors"
	"fmt"

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

func newFootprintFromDescriptor(fd fonts.FontDescriptor, format Format) (out Footprint, err error) {
	cmap, err := fd.LoadCmap() // load the cmap...
	if err != nil {
		return Footprint{}, err
	}
	out.Runes = NewRuneSetFromCmap(cmap) // ... and build the corresponding rune set

	out.Family = fd.Family() // load the family

	sty, wei, str := fd.Aspect() // load the aspect properties ...
	out.Aspect.Style = sty
	out.Aspect.Weight = wei
	out.Aspect.Stretch = str

	// and try to fill the missing one with the "style"
	style := fd.AdditionalStyle()
	out.Aspect.inferFromStyle(style)

	// register the correct format
	out.Format = format

	return out, nil
}

// serializeTo serialize the Footprint in binary format,
// by appending to `dst` and returning the slice
func (as Footprint) serializeTo(dst []byte) []byte {
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
func (as *Footprint) deserializeFrom(data []byte) (int, error) {
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

func (ff Format) String() string {
	switch ff {
	case OpenType:
		return "OpenType"
	case PCF:
		return "PCF"
	case Type1:
		return "Type1"
	default:
		return fmt.Sprintf("<format %d>", ff)
	}
}

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

// We use a tree representation to facilitate
// the consistency check of a saved fontset against the
// file system.

// SystemFontset stores the font footprints scanned
// from the disk, from one or several source directories
type SystemFontset []fontsetNode

// fontsetNode represents one directory
type fontsetNode struct {
	directory  string
	children   []fontsetNode // sorted by `directory`
	footprints []Footprint   // the fonts contained in this directory, sorted by `Location.Filename`
}
