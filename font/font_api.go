// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

// Package font provides an high level API to access
// Opentype font properties.
// See packages [opentype] and [opentype/tables] for a lower level, more detailled API.
package font

import (
	"fmt"
	"math"

	ot "github.com/go-text/typesetting/font/opentype"
)

type Resource = ot.Resource

// ParseTTF parse an Opentype font file (.otf, .ttf).
// See ParseTTC for support for collections.
func ParseTTF(file Resource) (*Face, error) {
	ld, err := ot.NewLoader(file)
	if err != nil {
		return nil, err
	}
	ft, err := NewFont(ld)
	if err != nil {
		return nil, err
	}
	return &Face{Font: ft}, nil
}

// ParseTTC parse an Opentype font file, with support for collections.
// Single font files are supported, returning a slice with length 1.
func ParseTTC(file Resource) ([]*Face, error) {
	lds, err := ot.NewLoaders(file)
	if err != nil {
		return nil, err
	}
	out := make([]*Face, len(lds))
	for i, ld := range lds {
		ft, err := NewFont(ld)
		if err != nil {
			return nil, fmt.Errorf("reading font %d of collection: %s", i, err)
		}
		out[i] = &Face{Font: ft}
	}

	return out, nil
}

type GID = ot.GID

// EmptyGlyph represents an invisible glyph, which should not be drawn,
// but whose advance and offsets should still be accounted for when rendering.
const EmptyGlyph GID = math.MaxUint32

// CmapIter is an interator over a Cmap.
type CmapIter interface {
	// Next returns true if the iterator still has data to yield
	Next() bool

	// Char must be called only when `Next` has returned `true`
	Char() (rune, GID)
}

// Cmap stores a compact representation of a cmap,
// offering both on-demand rune lookup and full rune range.
// It is conceptually equivalent to a map[rune]GID, but is often
// implemented more efficiently.
type Cmap interface {
	// Iter returns a new iterator over the cmap
	// Multiple iterators may be used over the same cmap
	// The returned interface is garanted not to be nil.
	Iter() CmapIter

	// Lookup avoid the construction of a map and provides
	// an alternative when only few runes need to be fetched.
	// It returns a default value and false when no glyph is provided.
	Lookup(rune) (GID, bool)
}

// FontExtents exposes font-wide extent values, measured in font units.
// Note that typically ascender is positive and descender negative in coordinate systems that grow up.
type FontExtents struct {
	Ascender  float32 // Typographic ascender.
	Descender float32 // Typographic descender.
	LineGap   float32 // Suggested line spacing gap.
}

// LineMetric identifies one metric about the font.
type LineMetric uint8

const (
	// Distance above the baseline of the top of the underline.
	// Since most fonts have underline positions beneath the baseline, this value is typically negative.
	UnderlinePosition LineMetric = iota

	// Suggested thickness to draw for the underline.
	UnderlineThickness

	// Distance above the baseline of the top of the strikethrough.
	StrikethroughPosition

	// Suggested thickness to draw for the strikethrough.
	StrikethroughThickness

	SuperscriptEmYSize
	SuperscriptEmXOffset

	SubscriptEmYSize
	SubscriptEmYOffset
	SubscriptEmXOffset

	CapHeight
	XHeight
)

// GlyphData describe how to graw a glyph.
// It is either an GlyphOutline, GlyphSVG or GlyphBitmap.
type GlyphData interface {
	isGlyphData()
}

func (GlyphOutline) isGlyphData() {}
func (GlyphSVG) isGlyphData()     {}
func (GlyphBitmap) isGlyphData()  {}

// GlyphOutline exposes the path to draw for
// vector glyph.
// Coordinates are expressed in fonts units.
type GlyphOutline struct {
	Segments []Segment
}

type (
	GlyphExtents = ot.GlyphExtents
	Segment      = ot.Segment
	SegmentPoint = ot.SegmentPoint
)

// GlyphSVG is an SVG description for the glyph,
// as found in Opentype SVG table.
type GlyphSVG struct {
	// The SVG image content, decompressed if needed.
	// The actual glyph description is an SVG element
	// with id="glyph<GID>" (as in id="glyph12"),
	// and several glyphs may share the same Source
	Source []byte

	// According to the specification, a fallback outline
	// should be specified for each SVG glyphs
	Outline GlyphOutline
}

type GlyphBitmap struct {
	// The actual image content, whose interpretation depends
	// on the Format field.
	Data          []byte
	Format        BitmapFormat
	Width, Height int // number of columns and rows

	// Outline may be specified to be drawn with bitmap
	Outline *GlyphOutline
}

// BitmapFormat identifies the format on the glyph
// raw data. Across the various font files, many formats
// may be encountered : black and white bitmaps, PNG, TIFF, JPG.
type BitmapFormat uint8

const (
	_ BitmapFormat = iota
	// The [GlyphBitmap.Data] slice stores a black or white (0/1)
	// bit image, whose length L satisfies
	// L * 8 >= [GlyphBitmap.Width] * [GlyphBitmap.Height]
	BlackAndWhite
	// The [GlyphBitmap.Data] slice stores a PNG encoded image
	PNG
	// The [GlyphBitmap.Data] slice stores a JPG encoded image
	JPG
	// The [GlyphBitmap.Data] slice stores a TIFF encoded image
	TIFF
)

// BitmapSize expose the size of bitmap glyphs.
// One font may contain several sizes.
type BitmapSize struct {
	Height, Width uint16
	XPpem, YPpem  uint16
}

// FontID represents an identifier of a font (possibly in a collection),
// and an optional variable instance.
type FontID struct {
	File string // The filename or identifier of the font file.

	// The index of the face in a collection. It is always 0 for
	// single font files.
	Index uint16

	// For variable fonts, stores 1 + the instance index.
	// It is set to 0 to ignore variations, or for non variable fonts.
	Instance uint16
}
