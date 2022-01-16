package fontscan

import (
	"encoding/binary"
	"errors"
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
	buffer := make([]byte, 4+len(s)) // len as uint32 + data
	binary.BigEndian.PutUint32(buffer, uint32(len(s)))
	copy(buffer[4:], s)
	return buffer
}

func deserializeString(s *string, data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("invalid string (EOF)")
	}
	L := int(binary.BigEndian.Uint32(data))
	if len(data) < 4+L {
		return 0, errors.New("invalid string length (EOF)")
	}
	*s = string(data[4 : 4+L])
	return 4 + L, nil
}
