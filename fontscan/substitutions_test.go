package fontscan

import (
	"reflect"
	"sort"
	"testing"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
)

func Test_familyList_insertStart(t *testing.T) {
	tests := []struct {
		start    []string
		families []string
		strong   bool
		want     familyList
	}{
		{nil, nil, false, familyList{}},
		{nil, []string{"aa", "bb"}, false, familyList{{"aa", false}, {"bb", false}}},
		{[]string{"a"}, []string{"aa", "bb"}, true, familyList{{"aa", true}, {"bb", true}, {"a", true}}},
		{[]string{"a"}, []string{"aa", "bb"}, false, familyList{{"aa", false}, {"bb", false}, {"a", true}}},
	}
	for _, tt := range tests {
		fl := newFamilyList(tt.start)
		fl.insertStart(tt.families, tt.strong)
		if !reflect.DeepEqual(fl, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, fl)
		}
	}
}

func Test_familyList_insertEnd(t *testing.T) {
	tests := []struct {
		start    []string
		families []string
		strong   bool
		want     familyList
	}{
		{nil, nil, false, familyList{}},
		{nil, []string{"aa", "bb"}, true, familyList{{"aa", true}, {"bb", true}}},
		{[]string{"a"}, []string{"aa", "bb"}, true, familyList{{"a", true}, {"aa", true}, {"bb", true}}},
		{[]string{"a"}, []string{"aa", "bb"}, false, familyList{{"a", true}, {"aa", false}, {"bb", false}}},
	}
	for _, tt := range tests {
		fl := newFamilyList(tt.start)
		fl.insertEnd(tt.families, tt.strong)
		if !reflect.DeepEqual(fl, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, fl)
		}
	}
}

func Test_familyList_insertAfter(t *testing.T) {
	tests := []struct {
		start    []string
		element  string
		families []string
		strong   bool
		want     familyList
	}{
		{[]string{"f1"}, "f1", []string{"aa", "bb"}, false, familyList{{"f1", true}, {"aa", false}, {"bb", false}}},
		{[]string{"f1", "f2"}, "f1", []string{"aa", "bb"}, false, familyList{{"f1", true}, {"aa", false}, {"bb", false}, {"f2", true}}},
		{[]string{"f1", "f2"}, "f1", []string{"aa", "bb"}, true, familyList{{"f1", true}, {"aa", true}, {"bb", true}, {"f2", true}}},
	}
	for _, tt := range tests {
		fl := newFamilyList(tt.start)
		mark := fl.elementEquals(tt.element)
		if mark < 0 {
			t.Fatalf("element %s not found in %v", tt.element, fl)
		}
		fl.insertAfter(mark, tt.families, tt.strong)
		if !reflect.DeepEqual(fl, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, fl)
		}
	}
}

func Test_familyList_insertBefore(t *testing.T) {
	tests := []struct {
		start    []string
		element  string
		families []string
		strong   bool
		want     familyList
	}{
		{[]string{"f1"}, "f1", []string{"aa", "bb"}, false, familyList{{"aa", false}, {"bb", false}, {"f1", true}}},
		{[]string{"f1", "f2"}, "f2", []string{"aa", "bb"}, false, familyList{{"f1", true}, {"aa", false}, {"bb", false}, {"f2", true}}},
	}
	for _, tt := range tests {
		fl := newFamilyList(tt.start)
		mark := fl.elementEquals(tt.element)
		if mark < 0 {
			t.Fatalf("element %s not found in %v", tt.element, fl)
		}
		fl.insertBefore(mark, tt.families, tt.strong)
		if !reflect.DeepEqual(fl, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, fl)
		}
	}
}

func Test_familyList_replace(t *testing.T) {
	tests := []struct {
		start    []string
		element  string
		families []string
		strong   bool
		want     familyList
	}{
		{[]string{"f1"}, "f1", []string{"aa", "bb"}, false, familyList{{"aa", false}, {"bb", false}}},
		{[]string{"f1", "f2"}, "f2", []string{"aa", "bb"}, false, familyList{{"f1", true}, {"aa", false}, {"bb", false}}},
	}
	for _, tt := range tests {
		fl := newFamilyList(tt.start)
		mark := fl.elementEquals(tt.element)
		if mark < 0 {
			t.Fatalf("element %s not found in %v", tt.element, fl)
		}
		fl.replace(mark, tt.families, tt.strong)
		if !reflect.DeepEqual(fl, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, fl)
		}
	}
}

func Test_familyList_execute(t *testing.T) {
	tests := []struct {
		start []string
		args  substitution
		lang  LangID
		want  familyList
	}{
		{nil, substitution{familyEquals("f2"), []string{"aa", "bb"}, opReplace, 0}, 0, familyList{}},                                            // no match
		{[]string{"f1", "f2"}, substitution{familyEquals("f4"), []string{"aa", "bb"}, opReplace, 0}, 0, familyList{{"f1", true}, {"f2", true}}}, // no match

		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opReplace, 's'}, 0, familyList{{"f1", true}, {"aa", true}, {"bb", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opReplace, 'e'}, 0, familyList{{"f1", true}, {"aa", true}, {"bb", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opReplace, 'w'}, 0, familyList{{"f1", true}, {"aa", false}, {"bb", false}}},

		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opAppend, 's'}, 0, familyList{{"f1", true}, {"f2", true}, {"aa", true}, {"bb", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opAppend, 'e'}, 0, familyList{{"f1", true}, {"f2", true}, {"aa", true}, {"bb", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opAppend, 'w'}, 0, familyList{{"f1", true}, {"f2", true}, {"aa", false}, {"bb", false}}},

		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opAppendLast, 's'}, 0, familyList{{"f1", true}, {"f2", true}, {"aa", true}, {"bb", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opAppendLast, 'w'}, 0, familyList{{"f1", true}, {"f2", true}, {"aa", false}, {"bb", false}}},

		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opPrepend, 's'}, 0, familyList{{"f1", true}, {"aa", true}, {"bb", true}, {"f2", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opPrepend, 'e'}, 0, familyList{{"f1", true}, {"aa", true}, {"bb", true}, {"f2", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opPrepend, 'w'}, 0, familyList{{"f1", true}, {"aa", false}, {"bb", false}, {"f2", true}}},

		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opPrependFirst, 's'}, 0, familyList{{"aa", true}, {"bb", true}, {"f1", true}, {"f2", true}}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opPrependFirst, 'w'}, 0, familyList{{"aa", false}, {"bb", false}, {"f1", true}, {"f2", true}}},

		{[]string{"f1", "f2"}, substitution{langAndFamilyEqual{language.LangAr, "f2"}, []string{"aa", "bb"}, opPrependFirst, 'w'}, language.LangEn, familyList{{"f1", true}, {"f2", true}}},
		{[]string{"f1", "f2"}, substitution{langAndFamilyEqual{language.LangAr, "f2"}, []string{"aa", "bb"}, opPrependFirst, 'w'}, language.LangAr, familyList{{"aa", false}, {"bb", false}, {"f1", true}, {"f2", true}}},
		{[]string{"f1", "f2"}, substitution{langAndFamilyEqual{language.LangAr, "f7"}, []string{"aa", "bb"}, opPrependFirst, 'w'}, language.LangAr, familyList{{"f1", true}, {"f2", true}}},
		{[]string{"f1", "f2"}, substitution{langEqualsAndNoFamily{language.LangAr, "f2"}, []string{"aa", "bb"}, opPrependFirst, 'w'}, language.LangAr, familyList{{"f1", true}, {"f2", true}}},
		{[]string{"f1", "f2"}, substitution{langEqualsAndNoFamily{language.LangAr, "f7"}, []string{"aa", "bb"}, opPrependFirst, 'w'}, language.LangAr, familyList{{"aa", false}, {"bb", false}, {"f1", true}, {"f2", true}}},
	}
	for _, tt := range tests {
		fl := newFamilyList(tt.start)
		fl.execute(tt.args, tt.lang)
		if !reflect.DeepEqual(fl, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, fl)
		}
	}
}

func Test_familyList_compileTo(t *testing.T) {
	tests := []struct {
		fl   familyList
		want familyCrible
	}{
		{nil, familyCrible{}},
		{newFamilyList([]string{"a", "b"}), familyCrible{"a": scoreStrong{0, true}, "b": scoreStrong{1, true}}},
		{newFamilyList([]string{"a", "b", "a"}), familyCrible{"a": scoreStrong{0, true}, "b": scoreStrong{1, true}}},
		{familyList{{"a", false}, {"a", true}}, familyCrible{"a": scoreStrong{1, true}}},
		{familyList{{"a", false}, {"a", false}}, familyCrible{"a": scoreStrong{0, false}}},
		{familyList{{"a", false}, {"b", false}, {"a", true}}, familyCrible{"a": scoreStrong{2, true}, "b": scoreStrong{1, false}}},
	}
	for _, tt := range tests {
		got := make(familyCrible)
		tt.fl.compileTo(got)
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, got)
		}
	}
}

func (fc familyCrible) families() (strong, weak []string) {
	for k, v := range fc {
		if v.strong {
			strong = append(strong, k)
		} else {
			weak = append(weak, k)
		}
	}
	sort.Slice(strong, func(i, j int) bool { return fc[strong[i]].score < fc[strong[j]].score })
	sort.Slice(weak, func(i, j int) bool { return fc[weak[i]].score < fc[weak[j]].score })
	return
}

func Test_newFamilyCrible(t *testing.T) {
	tests := []struct {
		family               string
		wantStrong, wantWeak []string
	}{
		// these tests are extracted from the fontconfig reference implementation
		{
			"LuxiMono",
			[]string{"luximono"},
			[]string{"dejavulgcsansmono", "notosansmono", "dejavusansmono", "inconsolata", "andalemono", "couriernew", "cumberlandamt", "nimbusmonol", "nimbusmonops", "nimbusmono", "courier", "tlwgtypo", "tlwgtypist", "tlwgmono", "notosansmonocjkjp", "notosansmonocjkkr", "notosansmonocjksc", "notosansmonocjktc", "notosansmonocjkhk", "miriammono", "vlgothic", "ipamonagothic", "ipagothic", "sazanamigothic", "kochigothic", "arplkaitimgb", "msgothic", "umeplusgothic", "nsimsun", "mingliu", "arplshanheisununi", "arplnewsungmono", "hanyisong", "arplsungtilgb", "arplmingti2lbig5", "zysong18030", "nanumgothiccoding", "nanumgothic", "dejavusans", "undotum", "baekmukdotum", "baekmukgulim", "tlwgtypewriter", "hasida", "gfzemenunicode", "hapaxberbère", "lohitbengali", "lohitgujarati", "lohithindi", "lohitmarathi", "lohitmaithili", "lohitkashmiri", "lohitkonkani", "lohitnepali", "lohitsindhi", "lohitpunjabi", "lohittamil", "meera", "lohitmalayalam", "lohitkannada", "lohittelugu", "lohitoriya", "lklug", "freemono", "monospace", "terafik", "freesans", "arialunicodems", "arialunicode", "code2000", "code2001", "sans-serif"},
		},
		{
			"Arial",
			[]string{"arial", "arimo", "liberationsans", "albany", "albanyamt"},
			[]string{"helvetica", "nimbussans", "nimbussansl", "texgyreheros", "helveticaltstd", "dejavulgcsans", "notosans", "dejavusans", "verdana", "luxisans", "lucidasansunicode", "bpgglahointernational", "tahoma", "urwgothic", "nimbussansnarrow", "loma", "waree", "garuda", "umpush", "laksaman", "notosanscjkjp", "notosanscjkkr", "notosanscjksc", "notosanscjktc", "notosanscjkhk", "lohitdevanagari", "droidsansfallback", "khmeros", "nachlieli", "yuditunicode", "kerkis", "armnethelvetica", "artsounk", "bpgutf8m", "saysetthaunicode", "jglaooldarial", "gfzemenunicode", "pigiarniq", "bdavat", "bcompset", "kacst-qr", "urdunastaliqunicode", "raghindi", "muktinarrow", "malayalam", "sampige", "padmaa", "hapaxberbère", "msgothic", "umepluspgothic", "microsoftyahei", "microsoftjhenghei", "wenquanyizenhei", "wenquanyibitmapsong", "arplshanheisununi", "arplnewsung", "hiraginosans", "pingfangsc", "pingfangtc", "pingfanghk", "hiraginosanscns", "hiraginosansgb", "mgopenmodata", "vlgothic", "ipamonagothic", "ipagothic", "sazanamigothic", "kochigothic", "arplkaitimgb", "arplkaitimbig5", "arplsungtilgb", "arplmingti2lbig5", "ｍｓゴシック", "zysong18030", "tscu_paranar", "nanumgothic", "undotum", "baekmukdotum", "baekmukgulim", "applesdgothicneo", "kacstqura", "lohitbengali", "lohitgujarati", "lohithindi", "lohitmarathi", "lohitmaithili", "lohitkashmiri", "lohitkonkani", "lohitnepali", "lohitsindhi", "lohitpunjabi", "lohittamil", "meera", "lohitmalayalam", "lohitkannada", "lohittelugu", "lohitoriya", "lklug", "freesans", "arialunicodems", "arialunicode", "code2000", "code2001", "sans-serif", "roya", "koodak", "terafik", "itcavantgardegothic", "helveticanarrow"},
		},
		{
			"Helvetica",
			[]string{"helvetica", "nimbussans", "nimbussansl", "texgyreheros", "helveticaltstd"},
			[]string{"arial", "arimo", "liberationsans", "albany", "albanyamt", "dejavulgcsans", "notosans", "dejavusans", "verdana", "luxisans", "lucidasansunicode", "bpgglahointernational", "tahoma", "urwgothic", "nimbussansnarrow", "loma", "waree", "garuda", "umpush", "laksaman", "notosanscjkjp", "notosanscjkkr", "notosanscjksc", "notosanscjktc", "notosanscjkhk", "lohitdevanagari", "droidsansfallback", "khmeros", "nachlieli", "yuditunicode", "kerkis", "armnethelvetica", "artsounk", "bpgutf8m", "saysetthaunicode", "jglaooldarial", "gfzemenunicode", "pigiarniq", "bdavat", "bcompset", "kacst-qr", "urdunastaliqunicode", "raghindi", "muktinarrow", "malayalam", "sampige", "padmaa", "hapaxberbère", "msgothic", "umepluspgothic", "microsoftyahei", "microsoftjhenghei", "wenquanyizenhei", "wenquanyibitmapsong", "arplshanheisununi", "arplnewsung", "hiraginosans", "pingfangsc", "pingfangtc", "pingfanghk", "hiraginosanscns", "hiraginosansgb", "mgopenmodata", "vlgothic", "ipamonagothic", "ipagothic", "sazanamigothic", "kochigothic", "arplkaitimgb", "arplkaitimbig5", "arplsungtilgb", "arplmingti2lbig5", "ｍｓゴシック", "zysong18030", "tscu_paranar", "nanumgothic", "undotum", "baekmukdotum", "baekmukgulim", "applesdgothicneo", "kacstqura", "lohitbengali", "lohitgujarati", "lohithindi", "lohitmarathi", "lohitmaithili", "lohitkashmiri", "lohitkonkani", "lohitnepali", "lohitsindhi", "lohitpunjabi", "lohittamil", "meera", "lohitmalayalam", "lohitkannada", "lohittelugu", "lohitoriya", "lklug", "freesans", "arialunicodems", "arialunicode", "code2000", "code2001", "sans-serif", "roya", "koodak", "terafik", "itcavantgardegothic", "helveticanarrow"},
		},
	}

	for _, tt := range tests {
		got := make(familyCrible)
		got.fillWithSubstitutions(font.NormalizeFamily(tt.family), language.LangEn)
		strong, weak := got.families()
		if !(reflect.DeepEqual(strong, tt.wantStrong) && reflect.DeepEqual(weak, tt.wantWeak)) {
			t.Errorf("newFamilyCrible() = %v %v, want %v %v", strong, weak, tt.wantStrong, tt.wantWeak)
		}
	}
}

type lt = []weightedFamily

var (
	A = weightedFamily{"A", true}
	B = weightedFamily{"B", true}
	C = weightedFamily{"C", true}
	X = weightedFamily{"X", true}
	Y = weightedFamily{"Y", true}
	Z = weightedFamily{"Z", true}
)

func TestInsertAt(t *testing.T) {
	mkcap := func(cap int, strs ...weightedFamily) lt {
		return append(make(lt, 0, cap), strs...)
	}
	clone := func(s lt) lt {
		r := make(lt, len(s), cap(s))
		copy(r, s)
		return r
	}

	tests := []struct {
		start  lt
		at     int
		add    lt
		result lt
	}{
		{lt{}, 0, lt{A}, lt{A}},
		{lt{X}, 0, lt{A}, lt{A, X}},
		{lt{X}, 1, lt{A}, lt{X, A}},
		{mkcap(3, X), 0, lt{A}, lt{A, X}},
		{mkcap(3, X), 1, lt{A}, lt{X, A}},
		{mkcap(3, X), 0, lt{A, B}, lt{A, B, X}},
		{mkcap(3, X), 1, lt{A, B}, lt{X, A, B}},
		{mkcap(4, X, Y), 0, lt{A, B}, lt{A, B, X, Y}},
		{mkcap(4, X, Y), 1, lt{A, B}, lt{X, A, B, Y}},
		{mkcap(4, X, Y), 2, lt{A, B}, lt{X, Y, A, B}},
	}
	for _, tt := range tests {
		result := insertAt(clone(tt.start), tt.at, tt.add)
		if !reflect.DeepEqual(tt.result, result) {
			t.Fatalf("insertAt: expected %v, got %v", tt.result, result)
		}

		// replaceAt functions like insertAt when start == end
		result = replaceAt(clone(tt.start), tt.at, tt.at, tt.add)
		if !reflect.DeepEqual(tt.result, result) {
			t.Fatalf("replaceAt: expected %v, got %v", tt.result, result)
		}
	}
}

func TestReplaceAt(t *testing.T) {
	mkcap := func(cap int, strs ...weightedFamily) lt {
		return append(make(lt, 0, cap), strs...)
	}
	clone := func(s lt) lt {
		r := make(lt, len(s), cap(s))
		copy(r, s)
		return r
	}

	tests := []struct {
		start  lt
		at, to int
		add    lt
		result lt
	}{
		{mkcap(4, X, Y, Z), 0, 2, lt{A}, lt{A, Z}},

		{mkcap(4, X, Y, Z), 0, 1, lt{A}, lt{A, Y, Z}},
		{mkcap(4, X, Y, Z), 0, 1, lt{A, B}, lt{A, B, Y, Z}},
		{mkcap(4, X, Y, Z), 0, 1, lt{A, B, C}, lt{A, B, C, Y, Z}},
		{mkcap(4, X, Y, Z), 0, 2, lt{A}, lt{A, Z}},
		{mkcap(4, X, Y, Z), 0, 2, lt{A, B}, lt{A, B, Z}},
		{mkcap(4, X, Y, Z), 0, 2, lt{A, B, C}, lt{A, B, C, Z}},

		{mkcap(4, X, Y, Z), 1, 2, lt{A}, lt{X, A, Z}},
		{mkcap(4, X, Y, Z), 1, 2, lt{A, B}, lt{X, A, B, Z}},
		{mkcap(4, X, Y, Z), 1, 2, lt{A, B, C}, lt{X, A, B, C, Z}},
		{mkcap(4, X, Y, Z), 1, 3, lt{A}, lt{X, A}},
		{mkcap(4, X, Y, Z), 1, 3, lt{A, B}, lt{X, A, B}},
		{mkcap(4, X, Y, Z), 1, 3, lt{A, B, C}, lt{X, A, B, C}},
	}
	for _, tt := range tests {
		result := replaceAt(clone(tt.start), tt.at, tt.to, tt.add)
		if !reflect.DeepEqual(tt.result, result) {
			t.Fatalf("expected %v, got %v", tt.result, result)
		}
	}
}

func BenchmarkNewFamilyCrible(b *testing.B) {
	c := make(familyCrible)
	for i := 0; i < b.N; i++ {
		c.fillWithSubstitutions("Arial", language.LangEn)
	}
}

func TestSubstituteHelveticaOrder(t *testing.T) {
	c := make(familyCrible)
	c.fillWithSubstitutionsList([]string{font.NormalizeFamily("BlinkMacSystemFont"), font.NormalizeFamily("Helvetica")}, language.LangEn)
	// BlinkMacSystemFont is not known by the library, so it is expanded with generic sans-serif,
	// but with lower priority then Helvetica
	expected := []string{"blinkmacsystemfont", "helvetica", "nimbussans", "nimbussansl", "texgyreheros", "helveticaltstd"}
	ls, _ := c.families()
	if !reflect.DeepEqual(ls, expected) {
		t.Fatalf("expected %v, got %v", expected, ls)
	}
}

func TestLanguageSubstitutions(t *testing.T) {
	c := make(familyCrible)
	c.fillWithSubstitutions(font.NormalizeFamily("NimbusSans"), language.LangOr)
	if _, has := c["lohitoriya"]; !has {
		t.Fatal("missing Lohit Oriya")
	}
	c.reset()
	c.fillWithSubstitutions(font.NormalizeFamily("NimbusSans"), language.LangGu)
	if _, has := c["lohitgujarati"]; !has {
		t.Fatal("missing Lohit Gujarati")
	}
	c.reset()
	c.fillWithSubstitutions(font.NormalizeFamily("NimbusSans"), language.LangPa)
	if _, has := c["lohitgurmukhi"]; !has {
		t.Fatal("missing Lohit Gurmukhi")
	}
}
