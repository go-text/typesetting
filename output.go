package shaping

import "image"

type Output interface {
	// Advance returns the distance the Dot has advanced.
	Advance() int
	// Bounds returns the smallest rectangle capable of containing the shaped text.
	Bounds() image.Rectangle
	// Length returns the number of glyphs in the output.
	Length() int
	// Glyph returns the glyph at the given index.
	Glyph(int) Glyph
}
