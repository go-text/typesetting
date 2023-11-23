package shaping

import (
	"os"
	"reflect"
	"testing"
	"unicode"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/opentype/api"
	oFont "github.com/go-text/typesetting/opentype/api/font"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

func Test_ignoreFaceChange(t *testing.T) {
	tests := []struct {
		args rune
		want bool
	}{
		{' ', true},
		{'a', false},
		{'\n', true},
		{'\r', true},
		{'\f', true},
		{'\ufe01', true},
		{'\ufe02', true},
		{'\U000E0100', true},
		{'\u06DD', false},
	}
	for _, tt := range tests {
		if got := ignoreFaceChange(tt.args); got != tt.want {
			t.Errorf("ignoreFaceChange() = %v, want %v", got, tt.want)
		}
	}
}

// support any rune
type universalCmap struct{ api.Cmap }

func (universalCmap) Lookup(rune) (font.GID, bool) { return 0, true }

type upperCmap struct{ api.Cmap }

func (upperCmap) Lookup(r rune) (font.GID, bool) {
	return 0, unicode.IsUpper(r)
}

type lowerCmap struct{ api.Cmap }

func (lowerCmap) Lookup(r rune) (font.GID, bool) {
	return 0, unicode.IsLower(r)
}

func loadOpentypeFont(t testing.TB, filename string) font.Face {
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("opening font file: %s", err)
	}
	face, err := font.ParseTTF(file)
	if err != nil {
		t.Fatalf("parsing font file %s: %s", filename, err)
	}
	return face
}

func TestSplitByFontGlyphs(t *testing.T) {
	type args struct {
		input          Input
		availableFaces []font.Face
	}

	universalFont := &oFont.Face{Font: &oFont.Font{Cmap: universalCmap{}}}
	lowerFont := &oFont.Face{Font: &oFont.Font{Cmap: lowerCmap{}}}
	upperFont := &oFont.Face{Font: &oFont.Font{Cmap: upperCmap{}}}

	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	englishArabic := []rune("Hello " + "تثذرزسشص" + "world" + "لمنهويء")

	tests := []struct {
		name string
		args args
		want []Input
	}{
		{
			"no font change",
			args{
				input: Input{
					Text:     []rune("a simple text"),
					RunStart: 0, RunEnd: len("a simple text"),
				},
				availableFaces: []font.Face{universalFont},
			},
			[]Input{
				{
					Text:     []rune("a simple text"),
					RunStart: 0, RunEnd: len("a simple text"),
					Face: universalFont,
				},
			},
		},
		{
			"one change no spaces",
			args{
				input: Input{
					Text:     []rune("aaaAAA"),
					RunStart: 0, RunEnd: len("aaaAAA"),
				},
				availableFaces: []font.Face{lowerFont, upperFont},
			},
			[]Input{
				{
					Text:     []rune("aaaAAA"),
					RunStart: 0, RunEnd: 3,
					Face: lowerFont,
				},
				{
					Text:     []rune("aaaAAA"),
					RunStart: 3, RunEnd: 6,
					Face: upperFont,
				},
			},
		},
		{
			"one change with spaces",
			args{
				input: Input{
					Text:     []rune("aaa AAA "),
					RunStart: 0, RunEnd: len("aaa AAA "),
				},
				availableFaces: []font.Face{lowerFont, upperFont},
			},
			[]Input{
				{
					Text:     []rune("aaa AAA "),
					RunStart: 0, RunEnd: 4,
					Face: lowerFont,
				},
				{
					Text:     []rune("aaa AAA "),
					RunStart: 4, RunEnd: 8,
					Face: upperFont,
				},
			},
		},
		{
			"no font matched 1",
			args{
				input: Input{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: len("__"),
				},
				availableFaces: []font.Face{lowerFont, upperFont},
			},
			[]Input{
				{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: 2,
					Face: lowerFont,
				},
			},
		},
		{
			"no font matched 2",
			args{
				input: Input{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: len("__"),
				},
				availableFaces: []font.Face{upperFont, lowerFont},
			},
			[]Input{
				{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: 2,
					Face: upperFont,
				},
			},
		},
		{
			"mixed english arabic",
			args{
				input: Input{
					Text:     englishArabic,
					RunStart: 0, RunEnd: len(englishArabic),
				},
				availableFaces: []font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     englishArabic,
					RunStart: 0, RunEnd: 6,
					Face: latinFont,
				},
				{
					Text:     englishArabic,
					RunStart: 6, RunEnd: 14,
					Face: arabicFont,
				},
				{
					Text:     englishArabic,
					RunStart: 14, RunEnd: 19,
					Face: latinFont,
				},
				{
					Text:     englishArabic,
					RunStart: 19, RunEnd: 26,
					Face: arabicFont,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitByFontGlyphs(tt.args.input, tt.args.availableFaces); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitByFontGlyphs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitBidi(t *testing.T) {
	ltrSource := []rune("The quick brown fox jumps over the lazy dog.")
	rtlSource := []rune("الحب سماء لا تمط غير الأحلام")
	bidiSource := []rune("The quick سماء שלום لا fox تمط שלום غير the lazy dog.")
	bidi2Source := []rune("الحب سماء brown привет fox تمط jumps привет over غير الأحلام")
	type run struct {
		start, end int
		dir        di.Direction
	}
	for _, test := range []struct {
		text         []rune
		expectedRuns []run
	}{
		{
			text: ltrSource,
			expectedRuns: []run{
				{0, len(ltrSource), di.DirectionLTR},
			},
		},
		{
			text: rtlSource,
			expectedRuns: []run{
				{0, len(rtlSource), di.DirectionRTL},
			},
		},
		{
			text: bidiSource,
			expectedRuns: []run{
				// spaces are assigned to LTR runs
				{0, 10, di.DirectionLTR},
				{10, 22, di.DirectionRTL},
				{22, 27, di.DirectionLTR},
				{27, 39, di.DirectionRTL},
				{39, 53, di.DirectionLTR},
			},
		},
		{
			text: bidi2Source,
			// spaces are assigned to RTL runs
			expectedRuns: []run{
				{0, 10, di.DirectionRTL},
				{10, 26, di.DirectionLTR},
				{26, 31, di.DirectionRTL},
				{31, 48, di.DirectionLTR},
				{48, 60, di.DirectionRTL},
			},
		},
	} {
		var seg Segmenter
		seg.splitByBidi(test.text, di.DirectionLTR)
		inputs := seg.buffer1
		tu.Assert(t, len(inputs) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := inputs[i]
			tu.Assert(t, got.Direction == run.dir)
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
		}
	}
}
