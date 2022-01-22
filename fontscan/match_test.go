package fontscan

import (
	"reflect"
	"sort"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

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
		if got := applySubstitutions(ignoreBlanksAndCase(tt.family)); !reflect.DeepEqual(got.families(), tt.want) {
			t.Errorf("newFamilyCrible() = %v, want %v", got.families(), tt.want)
		}
	}
}

func fontsFromFamilies(families ...string) (out FontSet) {
	for _, family := range families {
		out = append(out, Footprint{Family: family})
	}
	return out
}

func TestFontMap_selectByFamily(t *testing.T) {
	tests := []struct {
		fontset    FontSet
		family     string
		substitute bool
		want       FontSet
	}{
		{nil, "", false, nil}, // no match on empty fontset
		// simple match
		{fontsFromFamilies("arial"), "Arial", false, fontsFromFamilies("arial")},
		// blank and case
		{fontsFromFamilies("ar Ial"), "Arial", false, fontsFromFamilies("ar Ial")},
		// two fonts
		{fontsFromFamilies("ar Ial", "emoji"), "Arial", false, fontsFromFamilies("ar Ial")},
		// substitution
		{fontsFromFamilies("arial"), "Helvetica", false, nil},
		{fontsFromFamilies("arial"), "Helvetica", true, fontsFromFamilies("arial")},
		{fontsFromFamilies("caladea", "XXX"), "cambria", true, fontsFromFamilies("caladea")},
		// substitution, with order
		{fontsFromFamilies("arial", "Helvetica"), "Helvetica", true, fontsFromFamilies("Helvetica", "arial")},
		// substitution, with order, and no matching fonts
		{fontsFromFamilies("arial", "Helvetica", "XXX"), "Helvetica", true, fontsFromFamilies("Helvetica", "arial")},
		// generic families
		{fontsFromFamilies("norasi", "XXX"), "serif", false, fontsFromFamilies("norasi")},
		// default to generic families
		{fontsFromFamilies("DEjaVuSerif", "XXX"), "cambria", true, fontsFromFamilies("DEjaVuSerif")},
		// substitutions
		{
			fontsFromFamilies("Nimbus Roman", "Tinos", "Liberation Serif", "DejaVu Serif", "arial"),
			"Times", true, fontsFromFamilies("Nimbus Roman", "Tinos", "Liberation Serif", "DejaVu Serif"),
		},
	}
	for _, tt := range tests {
		if got := tt.fontset.selectByFamily(tt.family, tt.substitute); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FontMap.selectByFamily() = \n%v, want \n%v", got, tt.want)
		}
	}
}

func BenchmarkNewFamilyCrible(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = applySubstitutions("Arial")
	}
}

func fontsetFromStretches(sts ...Stretch) (out FontSet) {
	for _, stretch := range sts {
		out = append(out, Footprint{Aspect: Aspect{Stretch: stretch}})
	}
	return out
}

func fontsetFromStyles(sts ...Style) (out FontSet) {
	for _, style := range sts {
		out = append(out, Footprint{Aspect: Aspect{Style: style}})
	}
	return out
}

func fontsetFromWeights(sts ...Weight) (out FontSet) {
	for _, weight := range sts {
		out = append(out, Footprint{Aspect: Aspect{Weight: weight}})
	}
	return out
}

func TestFontSet_matchStretch(t *testing.T) {
	tests := []struct {
		name string
		fs   FontSet
		args Stretch
		want Stretch
	}{
		{"exact match", fontsetFromStretches(1, 2, 3), 1, 1},
		{"approximate match narow1", fontsetFromStretches(0.5, 1.1, 3), 0.9, 0.5},
		{"approximate match narow2", fontsetFromStretches(1.2, 1.1, 3), 0.9, 1.1},
		{"approximate match wide1", fontsetFromStretches(1.2, 1.1, 3), 1.11, 1.2},
		{"approximate match wide2", fontsetFromStretches(1.2, 1.1, 3), 4, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.matchStretch(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.matchStretch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFontSet_matchStyle(t *testing.T) {
	tests := []struct {
		name string
		fs   FontSet
		args Style
		want Style
	}{
		{"exact match 1", fontsetFromStyles(fonts.StyleNormal, fonts.StyleOblique, fonts.StyleItalic), fonts.StyleNormal, fonts.StyleNormal},
		{"exact match 2", fontsetFromStyles(fonts.StyleNormal, fonts.StyleOblique, fonts.StyleItalic), fonts.StyleOblique, fonts.StyleOblique},
		{"exact match 3", fontsetFromStyles(fonts.StyleNormal, fonts.StyleOblique, fonts.StyleItalic), fonts.StyleItalic, fonts.StyleItalic},
		{"approximate match oblique", fontsetFromStyles(fonts.StyleNormal, fonts.StyleItalic), fonts.StyleOblique, fonts.StyleItalic},
		{"approximate match oblique", fontsetFromStyles(fonts.StyleNormal), fonts.StyleOblique, fonts.StyleNormal},
		{"approximate match italic", fontsetFromStyles(fonts.StyleNormal, fonts.StyleOblique), fonts.StyleItalic, fonts.StyleOblique},
		{"approximate match italic", fontsetFromStyles(fonts.StyleNormal), fonts.StyleItalic, fonts.StyleNormal},
		{"approximate match normal", fontsetFromStyles(fonts.StyleOblique, fonts.StyleItalic), fonts.StyleNormal, fonts.StyleOblique},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.matchStyle(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.matchStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFontSet_matchWeight(t *testing.T) {
	tests := []struct {
		name string
		fs   FontSet
		args Weight
		want Weight
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
			if got := tt.fs.matchWeight(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.matchWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fontsetFromAspects(as ...Aspect) (out FontSet) {
	for _, a := range as {
		out = append(out, Footprint{Aspect: a})
	}
	return out
}

func TestFontSet_selectBestMatch(t *testing.T) {
	defaultAspect := Aspect{Style: fonts.StyleNormal, Weight: fonts.WeightNormal, Stretch: fonts.StretchNormal}
	boldAspect := Aspect{Style: fonts.StyleNormal, Weight: fonts.WeightBold, Stretch: fonts.StretchNormal}
	boldItalicAspect := Aspect{Style: fonts.StyleItalic, Weight: fonts.WeightBold, Stretch: fonts.StretchNormal}
	narrowAspect := Aspect{Style: fonts.StyleItalic, Weight: fonts.WeightNormal, Stretch: fonts.StretchCondensed}

	tests := []struct {
		name string
		fs   FontSet
		args Aspect
		want Footprint
	}{
		{"exact match", fontsetFromAspects(defaultAspect, defaultAspect, boldAspect), defaultAspect, Footprint{Aspect: defaultAspect}},
		{"exact match", fontsetFromAspects(defaultAspect, defaultAspect, boldAspect), boldAspect, Footprint{Aspect: boldAspect}},
		{"exact match", fontsetFromAspects(defaultAspect, boldItalicAspect, boldAspect), boldItalicAspect, Footprint{Aspect: boldItalicAspect}},
		{"approximate match", fontsetFromAspects(defaultAspect, boldItalicAspect, boldAspect), Aspect{Style: fonts.StyleOblique}, Footprint{Aspect: boldItalicAspect}},
		{"approximate match", fontsetFromAspects(defaultAspect, boldItalicAspect, boldAspect, narrowAspect), Aspect{Stretch: fonts.StretchExtraCondensed}, Footprint{Aspect: narrowAspect}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fs.retainsBestMatches(tt.args)
			if got := tt.fs[0]; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FontSet.selectBestMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
