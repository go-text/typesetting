// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import "github.com/go-text/typesetting/font/opentype"

//go:generate ../../../../typesetting-utils/generators/binarygen/cmd/generator . _src.go

type GlyphID = uint16

// NameID is the ID for entries in the font table.
type NameID uint16

type Tag = opentype.Tag

// Float1616 is a float32, represented in
// fixed 16.16 format in font files.
type Float1616 = float32

func Float1616FromUint(v uint32) Float1616 {
	// value are actually signed integers
	return Float1616(int32(v)) / (1 << 16)
}

func Float1616ToUint(f Float1616) uint32 {
	return uint32(int32(f * (1 << 16)))
}

func Float214FromUint(v uint16) float32 {
	// value are actually signed integers
	return float32(int16(v)) / (1 << 14)
}

// Coord is real number in [-1;1], stored as a fixed 2.14 integer
type Coord int16

func NewCoord(c float64) Coord {
	return Coord(c * (1 << 14))
}

// Number of seconds since 12:00 midnight that started January 1st 1904 in GMT/UTC time zone.
type longdatetime = uint64

// PlatformID represents the platform id for entries in the name table.
type PlatformID uint16

// EncodingID represents the platform specific id for entries in the name table.
// The most common values are provided as constants.
type EncodingID uint16

// LanguageID represents the language used by an entry in the name table
type LanguageID uint16

// Offset16 is an offset into the input byte slice
type Offset16 uint16

// Offset32 is an offset into the input byte slice
type Offset32 uint32
