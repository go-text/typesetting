package fontscan

import (
	"bytes"
	"compress/gzip"
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
		// add buffer to store the length of the encoded footprint,
		// needed when decoding from a stream
		n := len(buffer)
		buffer = append(buffer, make([]byte, 4)...)

		buffer = fp.serializeTo(buffer)

		size := len(buffer) - n - 4
		binary.BigEndian.PutUint32(buffer[n:], uint32(size))
	}

	wr := gzip.NewWriter(w)
	_, err := wr.Write(buffer)
	if err != nil {
		return fmt.Errorf("serializing font footprints: %s", err)
	}
	err = wr.Close()
	if err != nil {
		return fmt.Errorf("compressing serialized font footprints: %s", err)
	}
	return nil
}

func deserializeFootprints(src io.Reader) ([]Footprint, error) {
	r, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("invalid compressed font footprint file: %s", err)
	}
	defer r.Close()

	var (
		buf    [4]byte
		out    []Footprint
		buffer bytes.Buffer
	)

	// read the expected length
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, fmt.Errorf("invalid font footprint format: %s", err)
	}
	L := binary.BigEndian.Uint32(buf[:])

	for i := uint32(0); i < L; i++ {
		// size of the encoded footprint
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return nil, fmt.Errorf("invalid fontset: %s", err)
		}
		size := binary.BigEndian.Uint32(buf[:])

		// buffer the footprint segment
		buffer.Reset()
		_, err := io.CopyN(&buffer, r, int64(size))
		if err != nil {
			return nil, fmt.Errorf("invalid fontset: %s", err)
		}

		var fp Footprint
		_, err = fp.deserializeFrom(buffer.Bytes())
		if err != nil {
			return nil, fmt.Errorf("invalid font set: %s", err)
		}

		out = append(out, fp)
	}

	return out, nil
}
