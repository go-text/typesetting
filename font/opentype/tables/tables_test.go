// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"bytes"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	ot "github.com/go-text/typesetting/font/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

// wrap td.Files.ReadFile
func readFontFile(t testing.TB, filepath string) *ot.Loader {
	t.Helper()

	file, err := td.Files.ReadFile(filepath)
	tu.AssertNoErr(t, err)

	fp, err := ot.NewLoader(bytes.NewReader(file))
	tu.AssertNoErr(t, err)

	return fp
}

func readTable(t testing.TB, fl *ot.Loader, tag string) []byte {
	t.Helper()

	table, err := fl.RawTable(ot.MustNewTag(tag))
	tu.AssertNoErr(t, err)

	return table
}

func numGlyphs(t *testing.T, fp *ot.Loader) int {
	t.Helper()

	table := readTable(t, fp, "maxp")
	maxp, _, err := ParseMaxp(table)
	tu.AssertNoErr(t, err)

	return int(maxp.NumGlyphs)
}

func TestParseBasicTables(t *testing.T) {
	for _, filename := range append(tu.Filenames(t, "morx"), tu.Filenames(t, "common")...) {
		fp := readFontFile(t, filename)
		_, _, err := ParseOs2(readTable(t, fp, "OS/2"))
		tu.AssertNoErr(t, err)

		_, _, err = ParseHead(readTable(t, fp, "head"))
		tu.AssertNoErr(t, err)

		_, _, err = ParseMaxp(readTable(t, fp, "maxp"))
		tu.AssertNoErr(t, err)

		_, _, err = ParseName(readTable(t, fp, "name"))
		tu.AssertNoErr(t, err)
	}
}

func TestParseCmap(t *testing.T) {
	// general parsing
	for _, filename := range tu.Filenames(t, "common") {
		fp := readFontFile(t, filename)
		_, _, err := ParseCmap(readTable(t, fp, "cmap"))
		tu.AssertNoErr(t, err)
	}

	// specialized tests for each format
	for _, filename := range tu.Filenames(t, "cmap") {
		fp := readFontFile(t, filename)
		_, _, err := ParseCmap(readTable(t, fp, "cmap"))
		tu.AssertNoErr(t, err)
	}

	// tests through a single table
	for _, filepath := range tu.Filenames(t, "cmap/table") {
		table, err := td.Files.ReadFile(filepath)
		tu.AssertNoErr(t, err)

		cmap, _, err := ParseCmap(table)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(cmap.Records) > 0)
	}
}
