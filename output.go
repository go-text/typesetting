package shaping

import (
	"github.com/benoitkugler/textlayout/harfbuzz"
	"golang.org/x/image/math/fixed"
)

type Glyph struct {
	harfbuzz.GlyphInfo
	harfbuzz.GlyphPosition
}

type Output struct {
	// Advance is the distance the Dot has advanced.
	Advance fixed.Int26_6
	// Baseline is the distance the baseline is from the top.
	Baseline fixed.Int26_6
	// Bounds is the smallest rectangle capable of containing the shaped text.
	Bounds fixed.Rectangle26_6
	// Glyphs are the shaped output text.
	Glyphs []Glyph
}
