package shaping

import "golang.org/x/image/math/fixed"

type Output interface {
	// Advance returns the distance the Dot has advanced.
	Advance() fixed.Int26_6
	// Baseline returns the distance the baseline is from the top.
	Baseline() fixed.Int26_6
	// Bounds returns the smallest rectangle capable of containing the shaped text.
	Bounds() fixed.Rectangle26_6
	// Length returns the number of glyphs in the output.
	Length() int
	// Glyph returns the glyph at the given index.
	Glyph(int) Glyph
}
