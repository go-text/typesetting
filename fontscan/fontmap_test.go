package fontscan

import (
	"reflect"
	"sort"
	"testing"
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
		if got := applySubstitutions(tt.family); !reflect.DeepEqual(got.families(), tt.want) {
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
		fontset FontSet
		family  string
		want    []Footprint
	}{
		{nil, "", nil}, // no match on empty fontset
		// simple match
		{fontsFromFamilies("arial"), "Arial", fontsFromFamilies("arial")},
		// blank and case
		{fontsFromFamilies("ar Ial"), "Arial", fontsFromFamilies("ar Ial")},
		// two fonts
		{fontsFromFamilies("ar Ial", "emoji"), "Arial", fontsFromFamilies("ar Ial")},
		// substitution
		{fontsFromFamilies("arial"), "Helvetica", fontsFromFamilies("arial")},
		{fontsFromFamilies("caladea", "XXX"), "cambria", fontsFromFamilies("caladea")},
		// substitution, with order
		{fontsFromFamilies("arial", "Helvetica"), "Helvetica", fontsFromFamilies("Helvetica", "arial")},
		// substitution, with order, and no matching fonts
		{fontsFromFamilies("arial", "Helvetica", "XXX"), "Helvetica", fontsFromFamilies("Helvetica", "arial")},
		// generic families
		{fontsFromFamilies("norasi", "XXX"), "serif", fontsFromFamilies("norasi")},
		// default to generic families
		{fontsFromFamilies("DEjaVuSerif", "XXX"), "cambria", fontsFromFamilies("DEjaVuSerif")},
		{
			fontsFromFamilies("Nimbus Roman", "Nimbus Roman No9 L", "TeX Gyre Termes", "Tinos", "Liberation Serif", "DejaVu Serif"),
			"Times", fontsFromFamilies("Nimbus Roman", "Nimbus Roman No9 L", "TeX Gyre Termes", "Tinos", "Liberation Serif", "DejaVu Serif"),
		},
	}
	for _, tt := range tests {
		if got := tt.fontset.selectByFamily(tt.family); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FontMap.selectByFamily() = \n%v, want \n%v", got, tt.want)
		}
	}
}

func BenchmarkNewFamilyCrible(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = applySubstitutions("Arial")
	}
}

func Test_FindFont(t *testing.T) {
	for _, family := range [...]string{
		"arial", "times", "deja vu",
	} {
		_, err := FindFont(family)
		if err != nil {
			t.Fatal(err)
		}
	}
}
