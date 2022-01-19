package fontscan

import (
	"fmt"
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

// "luximono"
// "DejaVu LGC Sans Mono"
//  "DejaVu Sans Mono"
//  "Bitstream Vera Sans Mono"
//  "Inconsolata"
//  "Andale Mono"
//  "Courier New"
//  "Cumberland AMT"
//  "Nimbus Mono L"
//  "Nimbus Mono PS"
//  "Nimbus Mono"
//  "Nimbus Mono PS"
//  "Nimbus Mono PS"
//  "Courier"
//  "Nimbus Mono PS"
//  "Nimbus Mono"
//  "Nimbus Mono L"
//  "Nimbus Mono PS"
//  "TlwgTypo"
//  "TlwgTypist"
//  "TlwgMono"
//  "Noto Sans Mono CJK JP"
//  "Noto Sans Mono CJK KR"
//  "Noto Sans Mono CJK SC"
//  "Noto Sans Mono CJK TC"
//  "Noto Sans Mono CJK HK"
//  "Khmer OS System"
//  "Miriam Mono"
//  "VL Gothic"
//  "IPAMonaGothic"
//  "IPAGothic"
//  "Sazanami Gothic"
//  "Kochi Gothic"
//  "AR PL KaitiM GB"
//  "MS Gothic"
//  "UmePlus Gothic"
//  "NSimSun"
//  "MingLiu"
//  "AR PL ShanHeiSun Uni"
//  "AR PL New Sung Mono"
//  "HanyiSong"
//  "AR PL SungtiL GB"
//  "AR PL Mingti2L Big5"
//  "ZYSong18030"
//  "NanumGothicCoding"
//  "NanumGothic"
//  "DejaVu Sans Mono"
//  "NanumGothic"
//  "DejaVu Sans"
//  "UnDotum"
//  "Baekmuk Dotum"
//  "Baekmuk Gulim"
//  "TlwgTypo"
//  "TlwgTypist"
//  "TlwgTypewriter"
//  "TlwgMono"
//  "Hasida"
//  "Mitra Mono"
//  "GF Zemen Unicode"
//  "Hapax Berbère"
//  "Lohit Bengali"
//  "Lohit Gujarati"
//  "Lohit Hindi"
//  "Lohit Marathi"
//  "Lohit Maithili"
//  "Lohit Kashmiri"
//  "Lohit Konkani"
//  "Lohit Nepali"
//  "Lohit Sindhi"
//  "Lohit Punjabi"
//  "Lohit Tamil"
//  "Meera"
//  "Lohit Malayalam"
//  "Lohit Kannada"
//  "Lohit Telugu"
//  "Lohit Oriya"
//  "LKLUG"
//  "FreeMono"
//  "monospace"
//  "Terafik"

//  "monospace"
//  "Courier"
//  "FreeSans"
//  "Arial Unicode MS"
//  "Arial Unicode"
//  "Code2000"
//  "Code2001"

func Test_newFamilyCrible(t *testing.T) {
	tests := []struct {
		name   string
		family string
		want   familyCrible
	}{
		{
			"no substitutions", "XXX", familyCrible{"xxx": 0},
		},
		// {
		// 	"one level substitution", "MingLiu", familyCrible{"mingliu": 0, "notoserifcjktc": 1, "arplumingtw": 2},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newFamilyCrible(tt.family); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFamilyCrible() = %v, want %v", got, tt.want)
			}
		})
	}

	fmt.Println(newFamilyCrible("luximono").families())
}

func fontsFromFamilies(families ...string) (out []Footprint) {
	for _, family := range families {
		out = append(out, Footprint{Family: family})
	}
	return out
}

func TestFontMap_selectByFamily(t *testing.T) {
	type fields struct {
		systemFootprint []Footprint
		userFootprint   []Footprint
	}
	tests := []struct {
		fields fields
		family string
		want   []Footprint
	}{
		{fields{nil, nil}, "", nil}, // no match on empty fontset
		// simple match
		{fields{fontsFromFamilies("arial"), nil}, "Arial", fontsFromFamilies("arial")},
		// blank and case
		{fields{fontsFromFamilies("ar Ial"), nil}, "Arial", fontsFromFamilies("ar Ial")},
		// two sources
		{fields{fontsFromFamilies("ar Ial"), fontsFromFamilies("emoji")}, "Arial", fontsFromFamilies("ar Ial")},
		// substitution
		{fields{fontsFromFamilies("arial"), nil}, "Helvetica", fontsFromFamilies("arial")},
		{fields{fontsFromFamilies("caladea", "XXX"), nil}, "cambria", fontsFromFamilies("caladea")},
		// substitution, with order
		{fields{fontsFromFamilies("arial", "Helvetica"), nil}, "Helvetica", fontsFromFamilies("Helvetica", "arial")},
		// substitution, with order, and no matching fonts
		{fields{fontsFromFamilies("arial", "Helvetica", "XXX"), nil}, "Helvetica", fontsFromFamilies("Helvetica", "arial")},
		// generic families
		{fields{fontsFromFamilies("cambria", "XXX"), nil}, "serif", fontsFromFamilies("cambria")},
		// default to generic families
		{fields{fontsFromFamilies("DEjaVuSerif", "XXX"), nil}, "cambria", fontsFromFamilies("DEjaVuSerif")},

		{
			fields{fontsFromFamilies("Nimbus Roman", "Nimbus Roman No9 L", "TeX Gyre Termes", "Tinos", "Liberation Serif", "DejaVu Serif"), nil},
			"Times", fontsFromFamilies("Nimbus Roman", "Nimbus Roman No9 L", "TeX Gyre Termes", "Tinos", "Liberation Serif", "DejaVu Serif"),
		},
	}
	for _, tt := range tests {
		fm := &FontSet{
			systemFootprint: tt.fields.systemFootprint,
			userFootprint:   tt.fields.userFootprint,
		}
		if got := fm.selectByFamily(tt.family); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FontMap.selectByFamily() = \n%v, want \n%v", got, tt.want)
		}
	}
}
