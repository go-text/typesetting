package fontscan

import (
	"os"
	"sort"
	"testing"

	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

func TestAssertSorted(t *testing.T) {
	ok := sort.SliceIsSorted(languagesInfo[:], func(i, j int) bool { return languagesInfo[i].lang < languagesInfo[j].lang })
	tu.Assert(t, ok)
}

func TestNewLanguageID(t *testing.T) {
	tests := []struct {
		l     language.Language
		want  LangID
		want1 bool
	}{
		{language.NewLanguage("a"), 0, false},
		{language.NewLanguage("af"), langAf, true},
		{language.NewLanguage("af-xx"), langAf, true}, // primary tag match
		{language.NewLanguage("az-az"), langAz_Az, true},
		{language.NewLanguage("az-ir"), langAz_Ir, true},
		{language.NewLanguage("az-xx"), 0, false}, // no match
		{language.NewLanguage("BR"), langBr, true},
		{language.NewLanguage("FR"), langFr, true},
		{language.NewLanguage("fr-be"), langFr, true},
		{language.NewLanguage("pa-pk"), langPa_Pk, true}, // exact match
		{language.NewLanguage("pa-pr"), langPa, true},    // primary tag match
	}
	for _, tt := range tests {
		got, got1 := NewLangID(tt.l)
		if got != tt.want {
			t.Errorf("NewLanguageID() got = %v, want %v", got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("NewLanguageID() got1 = %v, want %v", got1, tt.want1)
		}
	}
}

func TestNewLangset(t *testing.T) {
	// trivial check
	for id, lang := range languagesInfo {
		ls := newLangsetFromCoverage(lang.runes)
		tu.Assert(t, ls.Contains(LangID(id)))
	}

	file2, err := os.Open("../font/testdata/UbuntuMono-R.ttf")
	tu.AssertNoErr(t, err)
	defer file2.Close()
	ld, err := ot.NewLoader(file2)
	tu.AssertNoErr(t, err)
	fp, _, err := newFootprintFromLoader(ld, true, scanBuffer{})
	tu.AssertNoErr(t, err)

	ls := newLangsetFromCoverage(fp.Runes)
	en, _ := NewLangID(language.NewLanguage("en"))
	fr, _ := NewLangID(language.NewLanguage("fr"))
	ar, _ := NewLangID(language.NewLanguage("ar"))
	ta, _ := NewLangID(language.NewLanguage("ta"))
	tu.Assert(t, ls.Contains(en) && ls.Contains(fr) && !ls.Contains(ar) && !ls.Contains(ta))
}
