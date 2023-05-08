package shaping

import (
	"reflect"
	"testing"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/opentype/loader"
)

func simpleShape(text string, wordSpacing, letterSpacing int) int {
	text_ := []rune(text)
	input := Input{
		Text:      text_,
		RunStart:  0,
		RunEnd:    len(text_),
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("fr"),
	}
	out := (&HarfbuzzShaper{}).Shape(input)
	// TODO: implement support for wordSpacing and letterSpacing
	return out.Advance.Round()
}

// Custom spacing
func TestCustomSpacing(t *testing.T) {
	t.Skip()

	const text = "Une simple autre phrase."
	reference := simpleShape(text, 0, 0)
	// Word spacing feature
	withWord := simpleShape(text, 5, 0)
	if withWord != reference+3*5 {
		t.Error("word spacing not supported")
	}

	// Letter spacing feature
	nbGraphemes := len(text) // simple case here TODO: precise the rule for complexe graphemes
	withLetter := simpleShape(text, 0, 2)
	// Pango add only half of the latter spacing on boundaries
	// This is to discuss
	pangoBehavior := reference + nbGraphemes*2
	if withLetter != pangoBehavior {
		t.Error("letter spacing not supported")
	}

	// The two features are additive
	withBoth := simpleShape(text, 5, 2)
	if withBoth != reference+3*5+nbGraphemes*2 {
		t.Error("word and letter spacing not supported")
	}
}

func simpleWrap(text string, justify bool) Line {
	text_ := []rune(text)
	input := Input{
		Text:      text_,
		RunStart:  0,
		RunEnd:    len(text_),
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("fr"),
	}
	out := (&HarfbuzzShaper{}).Shape(input)
	var wr LineWrapper
	lines, _ := wr.WrapParagraph(WrapConfig{}, 300, text_, NewSliceIterator([]Output{out}), WrapScratch{})
	// TODO: implement justification
	return lines[0]
}

func TestJustification(t *testing.T) {
	t.Skip()

	const text = "Une simple autre phrase."
	noJustification := simpleWrap(text, false)
	withJustification := simpleWrap(text, true)

	// words inside the line should be justified to fill the extra space
	if reflect.DeepEqual(noJustification, withJustification) {
		t.Error()
	}
}

func TestFontFeatures(t *testing.T) {
	type Feature struct {
		Name  loader.Tag
		Value uint32
	}
	features := []Feature{
		{loader.MustNewTag("liga"), 1},
	}

	// TODO: forward the feature to Harfbuzz, probably in the Input field
	_ = features
}
