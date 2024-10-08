// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package opentype

import (
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
)

var (
	// TrueType is the first four bytes of an OpenType file containing a TrueType font
	TrueType = Tag(0x00010000)
	// AppleTrueType is the first four bytes of an OpenType file containing a TrueType font
	// (specifically one designed for Apple products, it's recommended to use TrueType instead)
	AppleTrueType = MustNewTag("true")
	// PostScript1 is the first four bytes of an OpenType file containing a PostScript 1 font
	PostScript1 = MustNewTag("typ1")
	// OpenType is the first four bytes of an OpenType file containing a PostScript Type 2 font
	// as specified by OpenType
	OpenType = MustNewTag("OTTO")

	// signatureWOFF is the magic number at the start of a WOFF file.
	signatureWOFF = MustNewTag("wOFF")

	ttcTag = MustNewTag("ttcf")

	errInvalidDfont = errors.New("invalid dfont")
)

// dfontResourceDataOffset is the assumed value of a dfont file's resource data
// offset.
//
// https://github.com/kreativekorp/ksfl/wiki/Macintosh-Resource-File-Format
// says that "A Mac OS resource file... [starts with an] offset from start of
// file to start of resource data section... [usually] 0x0100". In theory,
// 0x00000100 isn't always a magic number for identifying dfont files. In
// practice, it seems to work.
const dfontResourceDataOffset = 0x00000100

type Resource interface {
	Read([]byte) (int, error)
	ReadAt([]byte, int64) (int, error)
	Seek(int64, int) (int64, error)
}

// tableSection represents a table within the font file.
type tableSection struct {
	offset  uint32 // Offset into the file this table starts.
	length  uint32 // Length of this table within the file.
	zLength uint32 // Uncompressed length of this table.
}

// Loader is the low level font reader, providing
// full control over table loading.
type Loader struct {
	file   Resource             // source, needed to parse each table
	tables map[Tag]tableSection // header only, contents is processed on demand

	// Type represents the kind of this font being loaded.
	// It is one of TrueType, TrueTypeApple, PostScript1, OpenType
	Type Tag
}

// NewLoader reads the `file` header and returns
// a new lazy ot.
// `file` will be used to parse tables, and should not be close.
func NewLoader(file Resource) (*Loader, error) {
	return parseOneFont(file, 0, false)
}

// NewLoaders is the same as `NewLoader`, but supports collections.
func NewLoaders(file Resource) ([]*Loader, error) {
	_, err := file.Seek(0, io.SeekStart) // file might have been used before
	if err != nil {
		return nil, err
	}

	var bytes [4]byte
	_, err = file.Read(bytes[:])
	if err != nil {
		return nil, err
	}
	magic := NewTag(bytes[0], bytes[1], bytes[2], bytes[3])

	file.Seek(0, io.SeekStart)

	var (
		pr             *Loader
		offsets        []uint32
		relativeOffset bool
	)
	switch magic {
	case signatureWOFF, TrueType, OpenType, PostScript1, AppleTrueType:
		pr, err = parseOneFont(file, 0, false)
	case ttcTag:
		offsets, err = parseTTCHeader(file)
	case dfontResourceDataOffset:
		offsets, err = parseDfont(file)
		relativeOffset = true
	default:
		return nil, fmt.Errorf("unsupported font format %v", bytes)
	}
	if err != nil {
		return nil, err
	}

	// only one font
	if pr != nil {
		return []*Loader{pr}, nil
	}

	// collection
	out := make([]*Loader, len(offsets))
	for i, o := range offsets {
		out[i], err = parseOneFont(file, o, relativeOffset)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// dst is an optional storage which may be provided to reduce allocations.
func (pr *Loader) findTableBuffer(s tableSection, dst []byte) ([]byte, error) {
	if s.length != 0 && s.length < s.zLength {
		zbuf := io.NewSectionReader(pr.file, int64(s.offset), int64(s.length))
		r, err := zlib.NewReader(zbuf)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		if cap(dst) < int(s.zLength) {
			dst = make([]byte, s.zLength)
		}
		dst = dst[0:s.zLength]
		if _, err := io.ReadFull(r, dst); err != nil {
			return nil, err
		}
	} else {
		if cap(dst) < int(s.length) {
			dst = make([]byte, s.length)
		}
		dst = dst[0:s.length]
		if _, err := pr.file.ReadAt(dst, int64(s.offset)); err != nil {
			return nil, err
		}
	}
	return dst, nil
}

// HasTable returns true if [table] is present.
func (pr *Loader) HasTable(table Tag) bool {
	_, has := pr.tables[table]
	return has
}

// Tables returns all the tables found in the file,
// as a sorted slice.
func (ld *Loader) Tables() []Tag {
	out := make([]Tag, 0, len(ld.tables))
	for tag := range ld.tables {
		out = append(out, tag)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// RawTable returns the binary content of the given table,
// or an error if not found.
func (pr *Loader) RawTable(tag Tag) ([]byte, error) {
	return pr.RawTableTo(tag, nil)
}

// RawTable writes the binary content of the given table to [dst], returning it,
// or an error if not found.
func (pr *Loader) RawTableTo(tag Tag, dst []byte) ([]byte, error) {
	s, found := pr.tables[tag]
	if !found {
		return nil, fmt.Errorf("missing table %s", tag)
	}

	return pr.findTableBuffer(s, dst)
}

func parseOneFont(file Resource, offset uint32, relativeOffset bool) (parser *Loader, err error) {
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("invalid offset: %s", err)
	}

	var bytes [4]byte
	_, err = file.Read(bytes[:])
	if err != nil {
		return nil, err
	}
	magic := NewTag(bytes[0], bytes[1], bytes[2], bytes[3])

	switch magic {
	case signatureWOFF:
		parser, err = parseWOFF(file, offset, relativeOffset)
	case TrueType, OpenType, PostScript1, AppleTrueType:
		parser, err = parseOTF(file, offset, relativeOffset)
	case ttcTag, dfontResourceDataOffset: // no more collections allowed here
		return nil, errors.New("collections not allowed")
	default:
		return nil, fmt.Errorf("unknown font format tag %v", bytes)
	}

	if err != nil {
		return nil, err
	}

	return parser, nil
}

// support for collections

const maxNumFonts = 2048 // security implementation limit

// returns the offsets of each font
func parseTTCHeader(r io.Reader) ([]uint32, error) {
	// The https://www.microsoft.com/typography/otspec/otff.htm "Font
	// Collections" section describes the TTC header.
	var buf [12]byte
	if _, err := r.Read(buf[:]); err != nil {
		return nil, err
	}
	// skip versions
	numFonts := binary.BigEndian.Uint32(buf[8:])
	if numFonts == 0 {
		return nil, errors.New("empty font collection")
	}
	if numFonts > maxNumFonts {
		return nil, fmt.Errorf("number of fonts (%d) in collection exceed implementation limit (%d)",
			numFonts, maxNumFonts)
	}

	offsetsBytes := make([]byte, numFonts*4)
	_, err := io.ReadFull(r, offsetsBytes)
	if err != nil {
		return nil, err
	}
	return parseUint32s(offsetsBytes, int(numFonts)), nil
}

// parseDfont parses a dfont resource map, as per
// https://github.com/kreativekorp/ksfl/wiki/Macintosh-Resource-File-Format
//
// That unofficial wiki page lists all of its fields as *signed* integers,
// which looks unusual. The actual file format might use *unsigned* integers in
// various places, but until we have either an official specification or an
// actual dfont file where this matters, we'll use signed integers and treat
// negative values as invalid.
func parseDfont(r Resource) ([]uint32, error) {
	var buf [16]byte
	if _, err := r.Read(buf[:]); err != nil {
		return nil, err
	}
	resourceMapOffset := binary.BigEndian.Uint32(buf[4:])
	resourceMapLength := binary.BigEndian.Uint32(buf[12:])

	const (
		// (maxTableOffset + maxTableLength) will not overflow an int32.
		maxTableLength = 1 << 29
		maxTableOffset = 1 << 29
	)
	if resourceMapOffset > maxTableOffset || resourceMapLength > maxTableLength {
		return nil, errors.New("unsupported table offset or length")
	}

	const headerSize = 28
	if resourceMapLength < headerSize {
		return nil, errInvalidDfont
	}
	_, err := r.ReadAt(buf[:2], int64(resourceMapOffset+24))
	if err != nil {
		return nil, err
	}
	typeListOffset := int64(int16(binary.BigEndian.Uint16(buf[:])))
	if typeListOffset < headerSize || resourceMapLength < uint32(typeListOffset)+2 {
		return nil, errInvalidDfont
	}
	_, err = r.ReadAt(buf[:2], int64(resourceMapOffset)+typeListOffset)
	if err != nil {
		return nil, err
	}
	typeCount := int(binary.BigEndian.Uint16(buf[:])) // The number of types, minus one.
	if typeCount == 0xFFFF {
		return nil, errInvalidDfont
	}
	typeCount += 1

	const tSize = 8
	if tSize*uint32(typeCount) > resourceMapLength-uint32(typeListOffset)-2 {
		return nil, errInvalidDfont
	}

	typeList := make([]byte, tSize*typeCount)
	_, err = r.ReadAt(typeList, int64(resourceMapOffset)+typeListOffset+2)
	if err != nil {
		return nil, err
	}
	numFonts, resourceListOffset := 0, 0
	for i := 0; i < typeCount; i++ {
		if binary.BigEndian.Uint32(typeList[tSize*i:]) != 0x73666e74 { // "sfnt".
			continue
		}

		numFonts = int(int16(binary.BigEndian.Uint16(typeList[tSize*i+4:])))
		if numFonts < 0 {
			return nil, errInvalidDfont
		}
		// https://github.com/kreativekorp/ksfl/wiki/Macintosh-Resource-File-Format
		// says that the value in the wire format is "the number of
		// resources of this type, minus one."
		numFonts++

		resourceListOffset = int(int16(binary.BigEndian.Uint16((typeList[tSize*i+6:]))))
		if resourceListOffset < 0 {
			return nil, errInvalidDfont
		}
	}
	if numFonts == 0 {
		return nil, errInvalidDfont
	}
	if numFonts > maxNumFonts {
		return nil, fmt.Errorf("number of fonts (%d) in collection exceed implementation limit (%d)",
			numFonts, maxNumFonts)
	}

	const rSize = 12
	o, n := uint32(int(typeListOffset)+resourceListOffset), rSize*uint32(numFonts)
	if o > resourceMapLength || n > resourceMapLength-o {
		return nil, errInvalidDfont
	}

	offsetsBytes := make([]byte, n)
	_, err = r.ReadAt(offsetsBytes, int64(resourceMapOffset+o))
	if err != nil {
		return nil, err
	}

	offsets := make([]uint32, numFonts)
	for i := range offsets {
		o := 0xffffff & binary.BigEndian.Uint32(offsetsBytes[rSize*i+4:])
		// Offsets are relative to the resource data start, not the file start.
		// A particular resource's data also starts with a 4-byte length, which
		// we skip.
		o += dfontResourceDataOffset + 4
		if o > maxTableOffset {
			return nil, errors.New("unsupported table offset or length")
		}
		offsets[i] = o
	}

	return offsets, nil
}

// data length must have been checked
func parseUint32s(data []byte, count int) []uint32 {
	out := make([]uint32, count)
	for i := range out {
		out[i] = binary.BigEndian.Uint32(data[4*i:])
	}
	return out
}
