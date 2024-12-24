package fontscan

import (
	"reflect"
	"testing"

	"github.com/go-text/typesetting/font"
)

func allIndices(fs fontSet) []int {
	out := make([]int, len(fs))
	for i := range fs {
		out[i] = i
	}
	return out
}

func fontsFromFamilies(families ...string) (out fontSet) {
	for _, family := range families {
		out = append(out, Footprint{Family: font.NormalizeFamily(family)})
	}
	return out
}

func TestFontMap_selectByFamilyExact(t *testing.T) {
	tests := []struct {
		fontset    fontSet
		family     string
		substitute bool
		want       []int
	}{
		{nil, "", false, nil},                                                          // no match on empty fontset
		{fontsFromFamilies("xxx"), "Arial", false, nil},                                // no match
		{fontsFromFamilies("xxx"), "serif", false, nil},                                // no match on generic
		{fontsFromFamilies("xxx"), Math, false, nil},                                   // no match on generic
		{fontsFromFamilies("arial", "notosans", "cambriamath"), Math, false, []int{2}}, // strong match on generic
		// simple match
		{fontsFromFamilies("arial"), "Arial", false, []int{0}},
		// blank and case
		{fontsFromFamilies("ar Ial"), "Arial", false, []int{0}},
		// two fonts
		{fontsFromFamilies("ar Ial", "emoji"), "Arial", false, []int{0}},
		// substitution
		{fontsFromFamilies("arial"), "Helvetica", false, nil},
		// generic families
		{fontsFromFamilies("norasi", "XXX"), "serif", false, []int{0}},
		{fontsFromFamilies("norasi", "norasi", "XXX"), "serif", false, []int{0, 1}},              // many footprints with same family
		{fontsFromFamilies("norasi", "norasi", "XXX", "norasi"), "serif", false, []int{0, 1, 3}}, // many footprints with same family
		{fontsFromFamilies("rachana", "norasi", "XXX"), "serif", false, []int{1}},                // restrict to only one match
		// user provided precedence
		{
			fontSet{
				{Family: "Times", isUserProvided: true},
				{Family: "arial", isUserProvided: false},
				{Family: "arial", isUserProvided: true},
			},
			"arial",
			false,
			[]int{2, 1},
		},
	}
	for _, tt := range tests {
		if got := tt.fontset.selectByFamilyExact(tt.family, make(familyCrible), &scoredFootprints{}, &scoredFootprints{}); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FontMap.selectByFamily(%s) = \n%v, want \n%v", tt.family, got, tt.want)
		}
	}
}

func TestFontMap_selectByFamilyList(t *testing.T) {
	tests := []struct {
		fontset      fontSet
		family       string
		wantS, wantW []int
	}{
		{nil, "", nil, nil}, // no match on empty fontset
		// weak substitute
		{fontsFromFamilies("arial"), "Helvetica", nil, []int{0}},
		// string substitute
		{fontsFromFamilies("caladea", "XXX"), "cambria", []int{0}, nil},
		// substitution, with order
		{fontsFromFamilies("arial", "Helvetica"), "Helvetica", []int{1}, []int{0}},
		{fontsFromFamilies("arial", "Helvetica", "XXX"), "Helvetica", []int{1}, []int{0}},
		// default to generic families
		{fontsFromFamilies("DEjaVuSerif", "XXX"), "cambria", nil, []int{0}},
		// more complex substitutions
		{
			fontsFromFamilies("Nimbus Roman", "Tinos", "Liberation Serif", "DejaVu Serif", "arial"),
			"Times",
			[]int{0},
			[]int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		gotS, gotW := tt.fontset.selectByFamilyWithSubs([]string{tt.family}, make(familyCrible), &scoredFootprints{}, &scoredFootprints{})
		if !(reflect.DeepEqual(gotS, tt.wantS) && reflect.DeepEqual(gotW, tt.wantW)) {
			t.Errorf("FontMap.selectByFamily() = \n%v\n%v, want \n%v\n%v", gotS, gotW, tt.wantS, tt.wantW)
		}
	}
}

func fontsetFromStretches(sts ...font.Stretch) (out fontSet) {
	for _, stretch := range sts {
		out = append(out, Footprint{Aspect: font.Aspect{Stretch: stretch}})
	}
	return out
}

func fontsetFromStyles(sts ...font.Style) (out fontSet) {
	for _, style := range sts {
		out = append(out, Footprint{Aspect: font.Aspect{Style: style}})
	}
	return out
}

func fontsetFromWeights(sts ...font.Weight) (out fontSet) {
	for _, weight := range sts {
		out = append(out, Footprint{Aspect: font.Aspect{Weight: weight}})
	}
	return out
}

func TestFontSet_matchStretch(t *testing.T) {
	tests := []struct {
		name string
		fs   fontSet
		args font.Stretch
		want font.Stretch
	}{
		{"exact match", fontsetFromStretches(1, 2, 3), 1, 1},
		{"approximate match narow1", fontsetFromStretches(0.5, 1.1, 3), 0.9, 0.5},
		{"approximate match narow2", fontsetFromStretches(1.2, 1.1, 3), 0.9, 1.1},
		{"approximate match wide1", fontsetFromStretches(1.2, 1.1, 3), 1.11, 1.2},
		{"approximate match wide2", fontsetFromStretches(1.2, 1.1, 3), 4, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.matchStretch(allIndices(tt.fs), tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.matchStretch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFontSet_matchStyle(t *testing.T) {
	tests := []struct {
		name string
		fs   fontSet
		args font.Style
		want font.Style
	}{
		{"exact match 1", fontsetFromStyles(font.StyleNormal, styleOblique, font.StyleItalic), font.StyleNormal, font.StyleNormal},
		{"exact match 2", fontsetFromStyles(font.StyleNormal, styleOblique, font.StyleItalic), styleOblique, styleOblique},
		{"exact match 3", fontsetFromStyles(font.StyleNormal, styleOblique, font.StyleItalic), font.StyleItalic, font.StyleItalic},
		{"approximate match oblique", fontsetFromStyles(font.StyleNormal, font.StyleItalic), styleOblique, font.StyleItalic},
		{"approximate match oblique", fontsetFromStyles(font.StyleNormal), styleOblique, font.StyleNormal},
		{"approximate match italic", fontsetFromStyles(font.StyleNormal, styleOblique), font.StyleItalic, styleOblique},
		{"approximate match italic", fontsetFromStyles(font.StyleNormal), font.StyleItalic, font.StyleNormal},
		{"approximate match normal", fontsetFromStyles(styleOblique, font.StyleItalic), font.StyleNormal, styleOblique},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.matchStyle(allIndices(tt.fs), tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.matchStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFontSet_matchWeight(t *testing.T) {
	tests := []struct {
		name string
		fs   fontSet
		args font.Weight
		want font.Weight
	}{
		{"exact", fontsetFromWeights(100, 200, 220, 300), 200, 200},
		{"approximate 430 1", fontsetFromWeights(100, 200, 220, 300, 470, 420), 430, 470},
		{"approximate 430 2", fontsetFromWeights(100, 200, 220, 300, 420), 430, 420},
		{"approximate 430 3", fontsetFromWeights(510, 600), 430, 510},
		{"approximate 300 1", fontsetFromWeights(280, 301, 600), 300, 280},
		{"approximate 300 2", fontsetFromWeights(301, 600), 300, 301},
		{"approximate 600 1", fontsetFromWeights(595, 650), 600, 650},
		{"approximate 600 2", fontsetFromWeights(550, 200, 580), 600, 580},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.matchWeight(allIndices(tt.fs), tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.matchWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fontsetFromAspects(as ...font.Aspect) (out fontSet) {
	for _, a := range as {
		out = append(out, Footprint{Aspect: a})
	}
	return out
}

func TestFontSet_selectBestMatch(t *testing.T) {
	defaultAspect := font.Aspect{Style: font.StyleNormal, Weight: font.WeightNormal, Stretch: font.StretchNormal}
	boldAspect := font.Aspect{Style: font.StyleNormal, Weight: font.WeightBold, Stretch: font.StretchNormal}
	boldItalicAspect := font.Aspect{Style: font.StyleItalic, Weight: font.WeightBold, Stretch: font.StretchNormal}
	narrowAspect := font.Aspect{Style: font.StyleItalic, Weight: font.WeightNormal, Stretch: font.StretchCondensed}

	tests := []struct {
		name string
		fs   fontSet
		args font.Aspect
		want Footprint
	}{
		{"exact match", fontsetFromAspects(defaultAspect, defaultAspect, boldAspect), defaultAspect, Footprint{Aspect: defaultAspect}},
		{"exact match", fontsetFromAspects(defaultAspect, defaultAspect, boldAspect), boldAspect, Footprint{Aspect: boldAspect}},
		{"exact match", fontsetFromAspects(defaultAspect, boldItalicAspect, boldAspect), boldItalicAspect, Footprint{Aspect: boldItalicAspect}},
		{"approximate match", fontsetFromAspects(defaultAspect, boldItalicAspect, boldAspect), font.Aspect{Style: styleOblique}, Footprint{Aspect: boldItalicAspect}},
		{"approximate match", fontsetFromAspects(defaultAspect, boldItalicAspect, boldAspect, narrowAspect), font.Aspect{Stretch: font.StretchExtraCondensed}, Footprint{Aspect: narrowAspect}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fs.retainsBestMatches(allIndices(tt.fs), tt.args)
			if got := tt.fs[result[0]]; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.selectBestMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
