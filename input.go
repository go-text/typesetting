package shaping

import (
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Input interface {
	// Text returns the characters to be shaped.
	Text() []rune
	// Direction returns the directionality of the text.
	Direction() Direction
	// Face returns the font face to render the text in.
	Face() font.Face
	// Size returns the size of the font, eg. 14.
	// TODO is this a scaled value, exact pixels, or display dependand?
	Size() fixed.Int26_6
}
