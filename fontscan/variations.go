package fontscan

import (
	"encoding/binary"
	"errors"

	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
)

// axisFlag is a flag set if
// a font has a variation for the axis
type axisFlag uint8

const (
	axisWght axisFlag = 1 << iota
	axisWdth
	axisSlnt
	axisItal
)

type axisRange struct {
	Min, Max tables.Float1616
}

// write to the start of buffer, assuming len(buffer) >= 8
func (ar axisRange) serializeTo(buffer []byte) {
	binary.BigEndian.PutUint32(buffer[:], tables.Float1616ToUint(ar.Min))
	binary.BigEndian.PutUint32(buffer[4:], tables.Float1616ToUint(ar.Max))
}

func (ar *axisRange) deserializeFrom(buffer []byte) {
	ar.Min = tables.Float1616FromUint(binary.BigEndian.Uint32(buffer[:]))
	ar.Max = tables.Float1616FromUint(binary.BigEndian.Uint32(buffer[4:]))
}

// see also https://learn.microsoft.com/en-us/typography/opentype/spec/dvaraxisreg#registered-axis-tags
type Variations struct {
	flag axisFlag
	Wght axisRange
	Wdth axisRange
	Slnt axisRange
	Ital axisRange
}

var (
	fvarTag = ot.MustNewTag("fvar")
	wghtTag = ot.MustNewTag("wght")
	wdthTag = ot.MustNewTag("wdth")
	slntTag = ot.MustNewTag("slnt")
	italTag = ot.MustNewTag("ital")
)

// newVariations process the axis, selecting 'wght', 'wdth', 'slnt' or 'ital'
func newVariations(axes []font.VariationAxis) Variations {
	var out Variations
	for _, axis := range axes {
		switch axis.Tag {
		case wghtTag:
			out.flag |= axisWght
			out.Wght = axisRange{Min: axis.Minimum, Max: axis.Maximum}
		case wdthTag:
			out.flag |= axisWdth
			out.Wdth = axisRange{Min: axis.Minimum, Max: axis.Maximum}
		case slntTag:
			out.flag |= axisSlnt
			out.Slnt = axisRange{Min: axis.Minimum, Max: axis.Maximum}
		case italTag:
			out.flag |= axisItal
			out.Ital = axisRange{Min: axis.Minimum, Max: axis.Maximum}
		}
	}
	return out
}

const fullVariationsSize = 1 + 4*2*4

func (ls Variations) serialize() []byte {
	if ls.flag == 0 { // optimize non variable fonts : just store the flag
		return []byte{0}
	}
	var buffer [fullVariationsSize]byte
	buffer[0] = byte(ls.flag)
	ls.Wght.serializeTo(buffer[1:])
	ls.Wdth.serializeTo(buffer[1+8:])
	ls.Slnt.serializeTo(buffer[1+2*8:])
	ls.Ital.serializeTo(buffer[1+3*8:])
	return buffer[:]
}

// deserializeFrom reads the binary format produced by serialize
// it returns the number of bytes read from `data`
func (ls *Variations) deserializeFrom(data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("invalid variations (EOF)")
	}
	ls.flag = axisFlag(data[0])

	if ls.flag == 0 { // no variations
		return 1, nil
	}

	if len(data) < fullVariationsSize {
		return 0, errors.New("invalid complete variations (EOF)")
	}
	ls.Wght.deserializeFrom(data[1:])
	ls.Wdth.deserializeFrom(data[1+8:])
	ls.Slnt.deserializeFrom(data[1+2*8:])
	ls.Ital.deserializeFrom(data[1+3*8:])

	return fullVariationsSize, nil
}
