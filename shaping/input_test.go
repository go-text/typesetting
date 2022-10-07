package shaping

import (
	"os"
	"reflect"
	"testing"
	"unicode"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/go-text/typesetting/font"
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
type universalFont struct {
	fonts.Face
}

func (universalFont) NominalGlyph(rune) (font.GID, bool) { return 0, true }

type upperFont struct {
	fonts.Face
}

func (upperFont) NominalGlyph(r rune) (font.GID, bool) {
	return 0, unicode.IsUpper(r)
}

type lowerFont struct {
	fonts.Face
}

func (lowerFont) NominalGlyph(r rune) (font.GID, bool) {
	return 0, unicode.IsLower(r)
}

type failer interface {
	Fatalf(string, ...interface{})
}

func loadOpentypeFont(t failer, filename string) font.Face {
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
				availableFaces: []font.Face{universalFont{}},
			},
			[]Input{
				{
					Text:     []rune("a simple text"),
					RunStart: 0, RunEnd: len("a simple text"),
					Face: universalFont{},
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
				availableFaces: []font.Face{lowerFont{}, upperFont{}},
			},
			[]Input{
				{
					Text:     []rune("aaaAAA"),
					RunStart: 0, RunEnd: 3,
					Face: lowerFont{},
				},
				{
					Text:     []rune("aaaAAA"),
					RunStart: 3, RunEnd: 6,
					Face: upperFont{},
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
				availableFaces: []font.Face{lowerFont{}, upperFont{}},
			},
			[]Input{
				{
					Text:     []rune("aaa AAA "),
					RunStart: 0, RunEnd: 4,
					Face: lowerFont{},
				},
				{
					Text:     []rune("aaa AAA "),
					RunStart: 4, RunEnd: 8,
					Face: upperFont{},
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
				availableFaces: []font.Face{lowerFont{}, upperFont{}},
			},
			[]Input{
				{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: 2,
					Face: lowerFont{},
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
				availableFaces: []font.Face{upperFont{}, lowerFont{}},
			},
			[]Input{
				{
					Text:     []rune("__"),
					RunStart: 0, RunEnd: 2,
					Face: upperFont{},
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
