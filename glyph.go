package shaping

import "golang.org/x/image/font/sfnt"

type Glyph interface {
	// Segments returns the vector segments that describe the shape of the shaped text.
	Segments() sfnt.Segments
}
