package harfbuzz

import (
	"testing"

	td "github.com/go-text/typesetting-utils/harfbuzz"
	"github.com/go-text/typesetting/font"
	tu "github.com/go-text/typesetting/testutils"
)

// ported from harfbuzz/perf

func BenchmarkShaping(b *testing.B) {
	runs := []struct {
		fontFile string
		textFile string
	}{
		{
			"perf_reference/fonts/Roboto-Regular.ttf",
			"perf_reference/texts/en-thelittleprince.txt",
		},
		{
			"perf_reference/fonts/Roboto-Regular.ttf",
			"perf_reference/texts/en-words.txt",
		},
		{
			"perf_reference/fonts/SourceSerifVariable-Roman.ttf",
			"perf_reference/texts/react-dom.txt",
		},
		{
			"perf_reference/fonts/NotoSansDevanagari-Regular.ttf",
			"perf_reference/texts/hi-words.txt",
		},
		{
			"perf_reference/fonts/Amiri-Regular.ttf",
			"perf_reference/texts/fa-thelittleprince.txt",
		},
		{
			"perf_reference/fonts/NotoNastaliqUrdu-Regular.ttf",
			"perf_reference/texts/fa-thelittleprince.txt",
		},
		{
			"perf_reference/fonts/NotoNastaliqUrdu-Regular.ttf",
			"perf_reference/texts/fa-words.txt",
		},
		{
			"perf_reference/fonts/Gulzar-Regular.ttf",
			"perf_reference/texts/fa-thelittleprince.txt",
		},
		{
			"perf_reference/fonts/Gulzar-Regular.ttf",
			"perf_reference/texts/fa-words.txt",
		},
		{
			"perf_reference/fonts/NotoSansDuployan-Regular.otf",
			"perf_reference/texts/duployan.txt",
		},
	}

	for _, run := range runs {
		b.Run(run.textFile, func(b *testing.B) {
			shapeOne(b, run.textFile, run.fontFile)
		})
	}
}

func shapeOne(b *testing.B, textFile, fontFile string) {
	ft := openFontFile(b, fontFile)

	font := NewFont(font.NewFace(ft))

	textB, err := td.Files.ReadFile(textFile)
	tu.AssertNoErr(b, err)

	text := []rune(string(textB))

	buf := NewBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.AddRunes(text, 0, -1)
		buf.GuessSegmentProperties()
		buf.Shape(font, nil)
		buf.Clear()
	}
}
