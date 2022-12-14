//go:build go1.18
// +build go1.18

package shaping

import (
	"testing"

	"github.com/benoitkugler/textlayout/language"
	"github.com/go-text/typesetting/di"
)

// FuzzE2E shapes and wraps large strings looking for unshapable text or failures
// in rune accounting.
func FuzzE2E(f *testing.F) {
	face := loadOpentypeFont(f, "../font/testdata/Amiri-Regular.ttf")
	f.Add(benchParagraphLatin)
	f.Add(benchParagraphArabic)
	f.Fuzz(func(t *testing.T, input string) {
		textInput := []rune(input)
		var shaper HarfbuzzShaper
		out := shaper.Shape(Input{
			Text:      textInput,
			RunStart:  0,
			RunEnd:    len(textInput),
			Direction: di.DirectionRTL,
			Face:      face,
			Size:      16 * 72,
			Script:    language.Arabic,
			Language:  language.NewLanguage("AR"),
		})
		if out.Runes.Count != len(textInput) {
			t.Errorf("expected %d shaped runes, got %d", len(textInput), out.Runes.Count)
		}
		var l LineWrapper
		outs := l.WrapParagraph(WrapConfig{}, 100, textInput, out)
		totalRunes := 0
		for _, l := range outs {
			for _, run := range l {
				if run.Runes.Offset != totalRunes {
					t.Errorf("expected rune offset %d, got %d", totalRunes, run.Runes.Offset)
				}
				totalRunes += run.Runes.Count
			}
		}
		if totalRunes != len(textInput) {
			t.Errorf("mismatched runes! expected %d, but wrapped output only contains %d", len(textInput), totalRunes)
		}
		_ = outs
	})
}
