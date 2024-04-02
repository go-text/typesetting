package fontscan

import (
	"reflect"
	"sort"
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

func (fc familyCrible) families() []string {
	var out []string
	for k := range fc {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		return fc[out[i]] < fc[out[j]]
	})
	return out
}

func Test_newFamilyCrible(t *testing.T) {
	tests := []struct {
		family string
		want   []string
	}{
		// these tests are extracted from the fontconfig reference implementation
		{"LuxiMono", []string{"luximono", "dejavulgcsansmono", "dejavusansmono", "bitstreamverasansmono", "inconsolata", "andalemono", "couriernew", "cumberlandamt", "nimbusmonol", "nimbusmonops", "nimbusmono", "courier", "tlwgtypo", "tlwgtypist", "tlwgmono", "notosansmonocjkjp", "notosansmonocjkkr", "notosansmonocjksc", "notosansmonocjktc", "notosansmonocjkhk", "khmerossystem", "miriammono", "vlgothic", "ipamonagothic", "ipagothic", "sazanamigothic", "kochigothic", "arplkaitimgb", "msgothic", "umeplusgothic", "nsimsun", "mingliu", "arplshanheisununi", "arplnewsungmono", "hanyisong", "arplsungtilgb", "arplmingti2lbig5", "zysong18030", "nanumgothiccoding", "nanumgothic", "dejavusans", "undotum", "baekmukdotum", "baekmukgulim", "tlwgtypewriter", "hasida", "mitramono", "gfzemenunicode", "hapaxberbère", "lohitbengali", "lohitgujarati", "lohithindi", "lohitmarathi", "lohitmaithili", "lohitkashmiri", "lohitkonkani", "lohitnepali", "lohitsindhi", "lohitpunjabi", "lohittamil", "meera", "lohitmalayalam", "lohitkannada", "lohittelugu", "lohitoriya", "lklug", "freemono", "monospace", "terafik", "freesans", "arialunicodems", "arialunicode", "code2000", "code2001", "sans-serif"}},
		{"Arial", []string{"arial", "arimo", "liberationsans", "albany", "albanyamt", "helvetica", "nimbussans", "nimbussansl", "texgyreheros", "dejavulgcsans", "dejavusans", "bitstreamverasans", "verdana", "luxisans", "lucidasansunicode", "bpgglahointernational", "tahoma", "urwgothic", "nimbussansnarrow", "loma", "waree", "garuda", "umpush", "laksaman", "notosanscjkjp", "notosanscjkkr", "notosanscjksc", "notosanscjktc", "notosanscjkhk", "lohitdevanagari", "droidsansfallback", "khmeros", "nachlieli", "yuditunicode", "kerkis", "armnethelvetica", "artsounk", "bpgutf8m", "saysetthaunicode", "jglaooldarial", "gfzemenunicode", "pigiarniq", "bdavat", "bcompset", "kacst-qr", "urdunastaliqunicode", "raghindi", "muktinarrow", "padmaa", "hapaxberbère", "msgothic", "umepluspgothic", "microsoftyahei", "microsoftjhenghei", "wenquanyizenhei", "wenquanyibitmapsong", "arplshanheisununi", "arplnewsung", "mgopenmoderna", "mgopenmodata", "mgopencosmetica", "vlgothic", "ipamonagothic", "ipagothic", "sazanamigothic", "kochigothic", "arplkaitimgb", "arplkaitimbig5", "arplsungtilgb", "arplmingti2lbig5", "ｍｓゴシック", "zysong18030", "nanumgothic", "undotum", "baekmukdotum", "baekmukgulim", "kacstqura", "lohitbengali", "lohitgujarati", "lohithindi", "lohitmarathi", "lohitmaithili", "lohitkashmiri", "lohitkonkani", "lohitnepali", "lohitsindhi", "lohitpunjabi", "lohittamil", "meera", "lohitmalayalam", "lohitkannada", "lohittelugu", "lohitoriya", "lklug", "freesans", "arialunicodems", "arialunicode", "code2000", "code2001", "sans-serif", "roya", "koodak", "terafik", "itcavantgardegothic", "helveticanarrow"}},
	}

	for _, tt := range tests {
		got := make(familyCrible)
		if got.fillWithSubstitutions(font.NormalizeFamily(tt.family)); !reflect.DeepEqual(got.families(), tt.want) {
			t.Errorf("newFamilyCrible() = %v, want %v", got.families(), tt.want)
		}
	}
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
		{nil, "", false, nil}, // no match on empty fontset
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
		if got := tt.fontset.selectByFamilyExact(tt.family, &scoredFootprints{}, make(familyCrible)); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FontMap.selectByFamily() = \n%v, want \n%v", got, tt.want)
		}
	}
}

func TestFontMap_selectByFamilyList(t *testing.T) {
	tests := []struct {
		fontset    fontSet
		family     string
		substitute bool
		want       []int
	}{
		{nil, "", true, nil}, // no match on empty fontset
		{fontsFromFamilies("arial"), "Helvetica", true, []int{0}},
		{fontsFromFamilies("caladea", "XXX"), "cambria", true, []int{0}},
		// substitution, with order
		{fontsFromFamilies("arial", "Helvetica"), "Helvetica", true, []int{1, 0}},
		// substitution, with order, and no matching fonts
		{fontsFromFamilies("arial", "Helvetica", "XXX"), "Helvetica", true, []int{1, 0}},
		// default to generic families
		{fontsFromFamilies("DEjaVuSerif", "XXX"), "cambria", true, []int{0}},
		// substitutions
		{
			fontsFromFamilies("Nimbus Roman", "Tinos", "Liberation Serif", "DejaVu Serif", "arial"),
			"Times", true,
			[]int{0, 1, 2, 3},
		},
	}
	for _, tt := range tests {
		if got := tt.fontset.selectByFamilyWithSubs([]string{tt.family}, &scoredFootprints{}, make(familyCrible)); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FontMap.selectByFamily() = \n%v, want \n%v", got, tt.want)
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
