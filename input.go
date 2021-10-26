// SPDX-License-Identifier: Unlicense OR MIT

package shaping

import (
	"github.com/benoitkugler/textlayout/language"
	"github.com/go-text/di"
	"github.com/go-text/font"
	"golang.org/x/image/math/fixed"
)

type Input struct {
	// Text is the body of text being shaped. Only the range Text[RunStart:RunEnd] is considered
	// for shaping, with the rest provided as context for the shaper. This helps with, for example,
	// cross-run Arabic shaping or handling combining marks at the start of a run.
	Text []rune
	// RunStart and RunEnd indicate the subslice of Text being shaped.
	RunStart, RunEnd int
	// Direction is the directionality of the text.
	Direction di.Direction
	// Face is the font face to render the text in.
	Face font.Face
	// Size is the size of the font, eg. 14.
	// TODO is this a scaled value, exact pixels, or display dependand?
	Size fixed.Int26_6

	// Script is an identifier for the writing system used in the text.
	Script language.Script

	// Language is an identifier for the language of the text.
	Language language.Language
}
