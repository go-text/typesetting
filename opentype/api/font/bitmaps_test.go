// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"bytes"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

// wrap td.Files.ReadFile
func readFontFile(t testing.TB, filepath string) *loader.Loader {
	t.Helper()

	file, err := td.Files.ReadFile(filepath)
	tu.AssertNoErr(t, err)

	fp, err := loader.NewLoader(bytes.NewReader(file))
	tu.AssertNoErr(t, err)

	return fp
}

func readTable(t testing.TB, fl *loader.Loader, tag string) []byte {
	t.Helper()

	table, err := fl.RawTable(loader.MustNewTag(tag))
	tu.AssertNoErr(t, err)

	return table
}

func TestBloc(t *testing.T) {
	blocT, err := td.Files.ReadFile("toys/tables/bloc.bin")
	tu.AssertNoErr(t, err)
	bloc, _, err := tables.ParseCBLC(blocT)
	tu.AssertNoErr(t, err)

	bdatT, err := td.Files.ReadFile("toys/tables/bdat.bin")
	tu.AssertNoErr(t, err)

	bt, err := newBitmap(bloc, bdatT)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(bt) == 1)
	tu.Assert(t, len(bt[0].subTables) == 4)
}

func TestCBLC(t *testing.T) {
	for _, file := range td.WithCBLC {
		fp := readFontFile(t, file.Path)

		cblc, _, err := tables.ParseCBLC(readTable(t, fp, "CBLC"))
		tu.AssertNoErr(t, err)
		cbdt := readTable(t, fp, "CBDT")

		_, err = newBitmap(cblc, cbdt)
		tu.AssertNoErr(t, err)
	}
}

func TestEBLC(t *testing.T) {
	for _, file := range td.WithEBLC {
		fp := readFontFile(t, file.Path)

		eblc, _, err := tables.ParseCBLC(readTable(t, fp, "EBLC"))
		tu.AssertNoErr(t, err)
		ebdt := readTable(t, fp, "EBDT")

		_, err = newBitmap(eblc, ebdt)
		tu.AssertNoErr(t, err)
	}
}
