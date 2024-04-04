package shaping

import (
	"os"
	"reflect"
	"testing"
	"unicode"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
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
type universalCmap struct{ font.Cmap }

func (universalCmap) Lookup(rune) (font.GID, bool) { return 0, true }

type upperCmap struct{ font.Cmap }

func (upperCmap) Lookup(r rune) (font.GID, bool) {
	return 0, unicode.IsUpper(r)
}

type lowerCmap struct{ font.Cmap }

func (lowerCmap) Lookup(r rune) (font.GID, bool) {
	return 0, unicode.IsLower(r)
}

func loadOpentypeFont(t testing.TB, filename string) *font.Face {
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
		availableFaces []*font.Face
	}

	universalFont := &font.Face{Font: &font.Font{Cmap: universalCmap{}}}
	lowerFont := &font.Face{Font: &font.Font{Cmap: lowerCmap{}}}
	upperFont := &font.Face{Font: &font.Font{Cmap: upperCmap{}}}

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
				availableFaces: []*font.Face{universalFont},
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
				availableFaces: []*font.Face{lowerFont, upperFont},
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
				availableFaces: []*font.Face{lowerFont, upperFont},
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
				availableFaces: []*font.Face{lowerFont, upperFont},
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
				availableFaces: []*font.Face{upperFont, lowerFont},
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
				availableFaces: []*font.Face{latinFont, arabicFont},
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
		{
			"no change on starting space",
			args{
				input: Input{
					Text:     []rune(" غير الأحلام"),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام")),
				},
				availableFaces: []*font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     []rune(" غير الأحلام"),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام")),
					Face: arabicFont,
				},
			},
		},
		{
			"entirely whitespace",
			args{
				input: Input{
					Text:     []rune("   "),
					RunStart: 0, RunEnd: 3,
				},
				availableFaces: []*font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     []rune("   "),
					RunStart: 0, RunEnd: 3,
					Face: latinFont,
				},
			},
		},
		{
			"no change on ending space",
			args{
				input: Input{
					Text:     []rune(" غير الأحلام "),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام ")),
				},
				availableFaces: []*font.Face{latinFont, arabicFont},
			},
			[]Input{
				{
					Text:     []rune(" غير الأحلام "),
					RunStart: 0, RunEnd: len([]rune(" غير الأحلام ")),
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
		text             []rune
		defaultDirection di.Direction
		expectedRuns     []run
	}{
		{
			text:             ltrSource,
			defaultDirection: di.DirectionLTR,
			expectedRuns: []run{
				{0, len(ltrSource), di.DirectionLTR},
			},
		},
		{
			text:             ltrSource,
			defaultDirection: di.DirectionTTB,
			expectedRuns: []run{
				{0, len(ltrSource), di.DirectionTTB},
			},
		},
		{
			text:             ltrSource,
			defaultDirection: di.DirectionRTL,
			expectedRuns: []run{
				{0, len(ltrSource) - 1, di.DirectionLTR},
				{len(ltrSource) - 1, len(ltrSource), di.DirectionRTL},
			},
		},
		{
			text:             ltrSource,
			defaultDirection: di.DirectionBTT,
			expectedRuns: []run{
				{0, len(ltrSource) - 1, di.DirectionTTB},
				{len(ltrSource) - 1, len(ltrSource), di.DirectionBTT},
			},
		},
		{
			text:             rtlSource,
			defaultDirection: di.DirectionRTL,
			expectedRuns: []run{
				{0, len(rtlSource), di.DirectionRTL},
			},
		},
		{
			text:             rtlSource,
			defaultDirection: di.DirectionBTT,
			expectedRuns: []run{
				{0, len(rtlSource), di.DirectionBTT},
			},
		},
		{
			text:             bidiSource,
			defaultDirection: di.DirectionLTR,
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
			text:             bidi2Source,
			defaultDirection: di.DirectionLTR,
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
		seg.splitByBidi(Input{Text: test.text, RunEnd: len(test.text), Direction: test.defaultDirection})
		tu.AssertC(t, len(seg.output) == len(test.expectedRuns), string(test.text))
		for i, run := range test.expectedRuns {
			got := seg.output[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Direction == run.dir)
		}
	}
}

func TestSplitScript(t *testing.T) {
	ltrSource := []rune("The quick brown fox jumps over the lazy dog.")
	rtlSource := []rune("الحب سماء لا تمط غير الأحلام")
	mixedLTRSource := []rune("The quick привет")
	mixedRTLSource := []rune("تمط שלום غي")
	commonSource := []rune("()[](][ gamma") // Common at first
	commonSource2 := []rune("gamma (Γ) est une lettre")
	commonSource3 := []rune("gamma (Γ [п] Γ) est une lettre") // nested delimiters
	withInherited := []rune("لمّا")
	type run struct {
		start, end int
		script     language.Script
	}
	for _, test := range []struct {
		text         []rune
		expectedRuns []run
	}{
		{ltrSource, []run{
			{0, len(ltrSource), language.Latin},
		}},
		{rtlSource, []run{
			{0, len(rtlSource), language.Arabic},
		}},
		{mixedLTRSource, []run{
			{0, 10, language.Latin},
			{10, 16, language.Cyrillic},
		}},
		{mixedRTLSource, []run{
			{0, 4, language.Arabic},
			{4, 9, language.Hebrew},
			{9, 11, language.Arabic},
		}},
		{commonSource, []run{
			{0, 13, language.Latin},
		}},
		{commonSource2, []run{
			{0, 7, language.Latin},
			{7, 8, language.Greek},
			{8, 24, language.Latin},
		}},
		{commonSource3, []run{
			{0, 7, language.Latin},
			{7, 10, language.Greek},
			{10, 11, language.Cyrillic},
			{11, 14, language.Greek},
			{14, 30, language.Latin},
		}},
		{withInherited, []run{
			{0, 4, language.Arabic},
		}},
	} {
		var seg Segmenter
		seg.splitByBidi(Input{Text: test.text, RunEnd: len(test.text), Direction: di.DirectionLTR})
		tu.Assert(t, len(seg.output) == 1)
		seg.input, seg.output = seg.output, seg.input

		seg.splitByScript()
		tu.Assert(t, len(seg.output) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := seg.output[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Script == run.script)
		}
	}
}

func TestSplitVertOrientation(t *testing.T) {
	type run struct {
		start, end int
		sideways   bool
	}
	for _, test := range []struct {
		text         []rune
		expectedRuns []run
	}{
		{
			[]rune("A regular latin sentence."),
			[]run{
				{0, 25, true},
			},
		},
		{
			[]rune("ごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみ"),
			[]run{
				{0, 44, false}, // Hiragana is upright
			},
		},
		{
			[]rune("ごさざ.しじす;ずせ"),
			[]run{
				{0, 10, false}, // Hiragana is upright
			},
		},
		{
			[]rune("もつ青いそら…"),
			[]run{
				{0, 2, false}, // Hiragana is upright
				{2, 3, false},
				{3, 7, false}, // Hiragana is upright
			},
		},
	} {
		var seg Segmenter
		seg.input = []Input{{Text: test.text, RunEnd: len(test.text), Direction: di.DirectionTTB}}

		seg.splitByScript()
		seg.input, seg.output = seg.output, seg.input
		seg.output = seg.output[:0]
		seg.splitByVertOrientation()
		tu.Assert(t, len(seg.output) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := seg.output[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Direction.HasVerticalOrientation())
			tu.Assert(t, got.Direction.IsSideways() == run.sideways)
		}
	}
}

func TestSplit(t *testing.T) {
	latinFont := loadOpentypeFont(t, "../font/testdata/Roboto-Regular.ttf")
	arabicFont := loadOpentypeFont(t, "../font/testdata/Amiri-Regular.ttf")
	fm := fixedFontmap{latinFont, arabicFont}

	sideways, upright := di.DirectionTTB, di.DirectionTTB
	sideways.SetSideways(true)
	upright.SetSideways(false)

	var seg Segmenter

	type run struct {
		start, end int
		dir        di.Direction
		script     language.Script
		face       *font.Face
	}
	for _, test := range []struct {
		text         string
		dir          di.Direction
		expectedRuns []run
	}{
		{
			"",
			di.DirectionLTR,
			[]run{{0, 0, di.DirectionLTR, language.Common, nil}},
		},
		{
			"The quick brown fox jumps over the lazy dog.",
			di.DirectionLTR,
			[]run{{0, 44, di.DirectionLTR, language.Latin, latinFont}},
		},
		{
			"الحب سماء لا تمط غير الأحلام",
			di.DirectionLTR,
			[]run{{0, 28, di.DirectionRTL, language.Arabic, arabicFont}},
		},
		{
			"The quick سماء שלום لا fox تمط שלום غير the lazy dog.",
			di.DirectionLTR,
			[]run{
				{0, 10, di.DirectionLTR, language.Latin, latinFont},
				{10, 15, di.DirectionRTL, language.Arabic, arabicFont},
				{15, 20, di.DirectionRTL, language.Hebrew, latinFont},
				{20, 22, di.DirectionRTL, language.Arabic, arabicFont},
				{22, 27, di.DirectionLTR, language.Latin, latinFont},
				{27, 31, di.DirectionRTL, language.Arabic, arabicFont},
				{31, 36, di.DirectionRTL, language.Hebrew, latinFont},
				{36, 39, di.DirectionRTL, language.Arabic, arabicFont},
				{39, 53, di.DirectionLTR, language.Latin, latinFont},
			},
		},
		{
			"الحب سماء brown привет fox تمط jumps привет over غير الأحلام",
			di.DirectionLTR,
			[]run{
				{0, 10, di.DirectionRTL, language.Arabic, arabicFont},
				{10, 16, di.DirectionLTR, language.Latin, latinFont},
				{16, 23, di.DirectionLTR, language.Cyrillic, latinFont},
				{23, 26, di.DirectionLTR, language.Latin, latinFont},
				{26, 31, di.DirectionRTL, language.Arabic, arabicFont},
				{31, 37, di.DirectionLTR, language.Latin, latinFont},
				{37, 44, di.DirectionLTR, language.Cyrillic, latinFont},
				{44, 48, di.DirectionLTR, language.Latin, latinFont},
				{48, 60, di.DirectionRTL, language.Arabic, arabicFont},
			},
		},
		// vertical text
		{
			"A french word",
			di.DirectionTTB,
			[]run{
				{0, 13, sideways, language.Latin, latinFont},
			},
		},
		{
			"A french word",
			upright,
			[]run{
				{0, 13, upright, language.Latin, latinFont},
			},
		},
		{
			"with upright \uff21\uff22\uff23",
			di.DirectionTTB,
			[]run{
				{0, 13, sideways, language.Latin, latinFont},
				{13, 16, upright, language.Latin, latinFont},
			},
		},
		{
			"ᠬᠦᠮᠦᠨ ᠪᠦ",
			di.DirectionTTB,
			[]run{
				{0, 8, sideways, language.Mongolian, latinFont},
			},
		},
	} {
		inputs := seg.Split(Input{
			Text:      []rune(test.text),
			RunEnd:    len([]rune(test.text)),
			Direction: test.dir,

			Size:     10,
			Language: "fr",
		}, fm)
		tu.Assert(t, len(inputs) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := inputs[i]
			tu.Assert(t, got.RunStart == run.start)
			tu.Assert(t, got.RunEnd == run.end)
			tu.Assert(t, got.Direction == run.dir)
			tu.Assert(t, got.Script == run.script)
			tu.Assert(t, got.Face == run.face)
			// check that input properties are properly copied
			tu.Assert(t, got.Size == 10)
			tu.Assert(t, got.Language == "fr")
		}

		// check that spliting a "middle" text slice is supported
		inputs = seg.Split(Input{
			Text:      []rune("DUMMY" + test.text + "DUMMY"),
			RunStart:  5,
			RunEnd:    5 + len([]rune(test.text)),
			Direction: test.dir,
		}, fm)
		tu.Assert(t, len(inputs) == len(test.expectedRuns))
		for i, run := range test.expectedRuns {
			got := inputs[i]
			tu.Assert(t, got.RunStart == 5+run.start)
			tu.Assert(t, got.RunEnd == 5+run.end)
			tu.Assert(t, got.Direction == run.dir)
			tu.Assert(t, got.Script == run.script)
			tu.Assert(t, got.Face == run.face)
		}
	}
}

func TestIssue127(t *testing.T) {
	// regression test for https://github.com/go-text/typesetting/issues/127
	str := []rune("لمّا")
	input := Input{
		Text:      str,
		RunStart:  0,
		RunEnd:    len(str),
		Direction: di.DirectionRTL,
		Language:  language.NewLanguage("ar"),
	}

	inputs := (&Segmenter{}).Split(input, fixedFontmap{benchArFace})
	// make sure Inherited script does no create a new run
	tu.Assert(t, len(inputs) == 1)
}
