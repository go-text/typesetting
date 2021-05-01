package shaping

import "image/color"

type Input interface {
	// Text returns the characters to be shaped.
	Text() []rune
	// Color returns the color to render the text in.
	Color() color.Color

	// TODO should these be wrapped in an interface, eg. Font() image/font?

	// FontFamily returns the name of the font family, eg. Helvetica.
	FontFamily() string
	// FontFace returns the name of the font face, eg. Bold Italic.
	// TODO would it be better to split these out into Bold() bool, Italic() bool, Monospace() bool, etc?
	FontFace() string
	// FontSize returns the size of the font, eg. 14.
	// TODO is this a scaled value, exact pixels, or display dependand?
	FontSize() int

	// TODO should these be wrapped in an interface, eg Script() Script, where Script has enums Regular, Super, and Sub?
	Superscript() bool
	Subscript() bool

	// TODO are these useful?
	Language() string
	Locale() string
}
