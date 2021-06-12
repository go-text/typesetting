package shaping

import "golang.org/x/image/font/sfnt"

type Glyph interface {
	// Advance returns the distance the Dot has advanced.
	Advance() fixed.Int26_6
	// Segments returns the vector segments that describe the shape of the shaped text.
	Segments() sfnt.Segments
}
