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
	ok := sort.SliceIsSorted(languagesRunes[:], func(i, j int) bool { return languagesRunes[i].lang < languagesRunes[j].lang })
	tu.Assert(t, ok)
}

func TestNewLanguageID(t *testing.T) {
	tests := []struct {
		l     language.Language
		want  LangID
		want1 bool
	}{
		{language.NewLanguage("a"), 0, false},
		{language.NewLanguage("af"), 2, true},
		{language.NewLanguage("af-xx"), 2, true}, // primary tag match
		{language.NewLanguage("az-az"), 14, true},
		{language.NewLanguage("az-ir"), 15, true},
		{language.NewLanguage("az-xx"), 0, false}, // no match
		{language.NewLanguage("BR"), 30, true},
		{language.NewLanguage("FR"), 69, true},
		{language.NewLanguage("fr-be"), 69, true},
		{language.NewLanguage("pa-pk"), 181, true}, // exact match
		{language.NewLanguage("pa-pr"), 180, true}, // primary tag match
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
	for id, lang := range languagesRunes {
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
