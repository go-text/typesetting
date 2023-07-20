//go:build go1.18
// +build go1.18

package shaping

import (
	"reflect"
	"testing"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/segmenter"
	"golang.org/x/image/math/fixed"
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
		out := []Output{shaper.Shape(Input{
			Text:      textInput,
			RunStart:  0,
			RunEnd:    len(textInput),
			Direction: di.DirectionRTL,
			Face:      face,
			Size:      16 * 72,
			Script:    language.Arabic,
			Language:  language.NewLanguage("AR"),
		})}
		if out[0].Runes.Count != len(textInput) {
			t.Errorf("expected %d shaped runes, got %d", len(textInput), out[0].Runes.Count)
		}
		var l LineWrapper
		outs, _ := l.WrapParagraph(WrapConfig{}, 100, textInput, NewSliceIterator(out))
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

// FuzzE2EVariableSize shapes and wraps large strings at varying text sizes and line widths.
func FuzzE2EVariableSize(f *testing.F) {
	face := loadOpentypeFont(f, "../font/testdata/Amiri-Regular.ttf")
	f.Add(benchParagraphLatin, byte(16), byte(100), byte(0))
	f.Add(benchParagraphArabic, byte(16), byte(100), byte(0))
	f.Fuzz(func(t *testing.T, input string, textSize byte, lineWidth byte, truncateBreakCont byte) {
		// We pack the wrapconfig fields into a byte in order to ensure that the fuzzer spends more time
		// exploring interesting parts of the space instead of varying bits that don't matter.
		wc := WrapConfig{
			TruncateAfterLines: int(truncateBreakCont & 0b11111),
			TextContinues:      truncateBreakCont&0b100000 != 0,
			BreakPolicy:        LineBreakPolicy((truncateBreakCont & 0b11000000 >> 5) % 3),
		}
		textInput := []rune(input)
		var shaper HarfbuzzShaper
		out := []Output{shaper.Shape(Input{
			Text:      textInput,
			RunStart:  0,
			RunEnd:    len(textInput),
			Direction: di.DirectionRTL,
			Face:      face,
			Size:      fixed.I(int(textSize)),
			Script:    language.Arabic,
			Language:  language.NewLanguage("AR"),
		})}
		if out[0].Runes.Count != len(textInput) {
			t.Errorf("expected %d shaped runes, got %d", len(textInput), out[0].Runes.Count)
		}
		var l LineWrapper
		outs, truncated := l.WrapParagraph(wc, int(lineWidth), textInput, NewSliceIterator(out))
		totalRunes := 0
		for _, l := range outs {
			for _, run := range l {
				if run.Runes.Offset != totalRunes {
					if reflect.DeepEqual(run, wc.Truncator) {
						continue
					}
					t.Errorf("expected rune offset %d, got %d", totalRunes, run.Runes.Offset)
				}
				totalRunes += run.Runes.Count
			}
		}
		totalRunes += truncated
		if totalRunes != len(textInput) {
			t.Errorf("mismatched runes! expected %d, but wrapped output only contains %d", len(textInput), totalRunes)
		}
		_ = outs
	})
}

func FuzzBreakOptions(f *testing.F) {
	f.Add(string([]rune{183067, 318808839, 476266048}))
	f.Add(benchParagraphArabic)
	f.Add(benchParagraphLatin)
	f.Fuzz(func(t *testing.T, input string) {
		runes := []rune(input)
		breaker := newBreaker(&segmenter.Segmenter{}, runes)
		var wordOptions []breakOption
		for b, ok := breaker.nextWordBreak(); ok; b, ok = breaker.nextWordBreak() {
			prevRuneIndex := 0
			if len(wordOptions) > 0 {
				prevRuneIndex = wordOptions[len(wordOptions)-1].breakAtRune + 1
			}
			segmentRunes := runes[prevRuneIndex : b.breakAtRune+1]
			segmentGraphemes := []breakOption{}
			for b, ok := breaker.nextGraphemeBreak(); ok; b, ok = breaker.nextGraphemeBreak() {
				// Adjust break offset to be relative to the start of the segment.
				b.breakAtRune -= prevRuneIndex
				segmentGraphemes = append(segmentGraphemes, b)
			}
			seg := segmenter.Segmenter{}
			seg.Init(segmentRunes)
			correctGraphemes := []int{}
			count := 0
			for iter := seg.GraphemeIterator(); iter.Next(); {
				g := iter.Grapheme()
				breakAt := g.Offset + len(g.Text) - 1
				firstGraphemeInText := prevRuneIndex == 0
				if !firstGraphemeInText {
					continue
				}
				count++
				correctGraphemes = append(correctGraphemes, breakAt)
			}
			if count > 0 && len(segmentGraphemes) != count {
				t.Errorf("runes[%d:%d] expected %d graphemes, got %d", prevRuneIndex, b.breakAtRune+1, count, len(segmentGraphemes))
				t.Errorf("correct graphemes: %v\ngot graphemes: %v", correctGraphemes, segmentGraphemes)
			}
			checkOptions(t, segmentRunes, segmentGraphemes)
			wordOptions = append(wordOptions, b)
		}
		checkOptions(t, runes, wordOptions)
	})
}
