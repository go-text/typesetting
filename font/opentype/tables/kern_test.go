// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

func TestParseKern(t *testing.T) {
	filepath := "common/FreeSerif.ttf"
	fp := readFontFile(t, filepath)
	_, _, err := ParseKern(readTable(t, fp, "kern"))
	tu.AssertNoErr(t, err)

	filepath = "toys/Kern2.ttf"
	fp = readFontFile(t, filepath)
	_, _, err = ParseKern(readTable(t, fp, "kern"))
	tu.AssertNoErr(t, err)

	for _, filepath := range []string{
		"toys/tables/kern0Exp.bin",
		"toys/tables/kern1.bin",
		"toys/tables/kern02.bin",
		"toys/tables/kern3.bin",
	} {
		table, err := td.Files.ReadFile(filepath)
		tu.AssertNoErr(t, err)

		kern, _, err := ParseKern(table)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(kern.Tables) > 0)

		for _, subtable := range kern.Tables {
			tu.Assert(t, subtable.Data() != nil)
		}
	}
}

func TestKern3(t *testing.T) {
	table, err := td.Files.ReadFile("toys/tables/kern3.bin")
	tu.AssertNoErr(t, err)

	kern, _, err := ParseKern(table)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(kern.Tables) == 5)

	expectedsLengths := [...][3]int{
		{570, 5688, 92},
		{570, 6557, 104},
		{570, 5832, 107},
		{570, 6083, 106},
		{570, 4828, 82},
	}
	for i := range kern.Tables {
		data, ok := kern.Tables[i].Data().(KernData3)
		tu.Assert(t, ok)
		exp := expectedsLengths[i]
		tu.Assert(t, len(data.LeftClass) == exp[0])
		tu.Assert(t, len(data.KernIndex) == exp[1])
		tu.Assert(t, len(data.Kernings) == exp[2])
	}
}
