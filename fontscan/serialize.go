package fontscan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// defines the routines to serialize a font set to
// the disk

// assume len(dst) >= 4
func serializeFloat(f float32, dst []byte) {
	binary.BigEndian.PutUint32(dst, math.Float32bits(f))
}

// assume len(src) >= 4
func deserializeFloat(src []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(src))
}

func serializeString(s string) []byte {
	L := len(s)
	if L > math.MaxUint16 { // never happen in practice
		L = math.MaxUint16
	}
	buffer := make([]byte, 2+L) // len as uint16 + data
	binary.BigEndian.PutUint16(buffer, uint16(L))
	copy(buffer[2:], s)
	return buffer
}

func deserializeString(s *string, data []byte) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("invalid string (EOF)")
	}
	L := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+L {
		return 0, errors.New("invalid string length (EOF)")
	}
	*s = string(data[2 : 2+L])
	return 2 + L, nil
}

func serializeFootprints(footprints []Footprint, w io.Writer) error {
	// len as uint32 + minimum size for a footprint
	buffer := make([]byte, 4, 4+len(footprints)*(aspectSize+1+2))
	binary.BigEndian.PutUint32(buffer[:], uint32(len(footprints)))

	for _, fp := range footprints {
		buffer = fp.serializeTo(buffer)
	}

	_, err := w.Write(buffer)
	if err != nil {
		return fmt.Errorf("serializing font footprints: %s", err)
	}
	return nil
}

func deserializeFootprints(data []byte) ([]Footprint, error) {
	// read the expected length
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid font set (EOF)")
	}
	L := binary.BigEndian.Uint32(data)
	n := 4
	var out []Footprint
	for i := uint32(0); i < L; i++ {
		var fp Footprint
		read, err := fp.deserializeFrom(data[n:])
		if err != nil {
			return nil, fmt.Errorf("invalid font set: %s", err)
		}
		n += read

		out = append(out, fp)
	}
	return out, nil
}
