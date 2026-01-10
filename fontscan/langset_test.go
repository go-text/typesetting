package fontscan

import (
	"os"
	"testing"

	ot "github.com/unidoc/typesetting/font/opentype"
	"github.com/unidoc/typesetting/language"
	tu "github.com/unidoc/typesetting/testutils"
)

func TestNewLangset(t *testing.T) {
	// trivial check
	for id, runes := range languagesRunes {
		ls := newLangsetFromCoverage(runes)
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
	tu.Assert(t, ls.Contains(language.LangEn) && ls.Contains(language.LangFr) && !ls.Contains(language.LangAr) && !ls.Contains(language.LangTa))
}
