// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"fmt"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"
)

func TestKern0(t *testing.T) {
	table, err := td.Files.ReadFile("toys/tables/kern0Exp.bin")
	tu.AssertNoErr(t, err)

	kern, _, err := tables.ParseKern(table)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(kern.Tables) == 1)

	kt := newKernSubtable(kern.Tables[0])
	data, ok := kt.Data.(Kern0)
	tu.Assert(t, ok)

	expecteds := []struct { // value extracted from harfbuzz run
		left, right GID
		kerning     int16
	}{
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 1059, 0},
		{1059, 65535, 0},
	}

	for _, exp := range expecteds {
		got := data.KernPair(exp.left, exp.right)
		tu.Assert(t, got == exp.kerning)
	}
}

func TestKern2(t *testing.T) {
	filepath := "toys/Kern2.ttf"
	fp := readFontFile(t, filepath)
	kern, _, err := tables.ParseKern(readTable(t, fp, "kern"))
	tu.AssertNoErr(t, err)

	tu.Assert(t, len(kern.Tables) == 3)

	k0, k1, k2 := newKernSubtable(kern.Tables[0]).Data, newKernSubtable(kern.Tables[1]).Data, newKernSubtable(kern.Tables[2]).Data
	_, is0 := k0.(Kern0)
	tu.Assert(t, is0)

	type2, ok := k1.(Kern2)
	tu.Assert(t, ok)
	expectedSubtable := map[[2]GID]int16{
		{67, 68}: 0,
		{68, 69}: 0,
		{69, 70}: -30,
		{70, 71}: 0,
		{71, 72}: 0,
		{72, 73}: -20,
		{73, 74}: 0,
		{74, 75}: 0,
		{75, 76}: 0,
		{76, 77}: 0,
		{77, 78}: 0,
		{78, 79}: 0,
		{79, 80}: 0,
		{80, 81}: 0,
		{81, 82}: 0,
		{36, 57}: 0,
	}
	for k, exp := range expectedSubtable {
		got := type2.KernPair(k[0], k[1])
		tu.AssertC(t, exp == got, fmt.Sprintf("invalid kern subtable : for (%d, %d) expected %d, got %d", k[0], k[1], exp, got))
	}

	type2, ok = k2.(Kern2)
	tu.Assert(t, ok)
	expectedSubtable = map[[2]GID]int16{
		{36, 57}: -80,
	}
	for k, exp := range expectedSubtable {
		got := type2.KernPair(k[0], k[1])
		tu.Assert(t, exp == got)
		tu.AssertC(t, exp == got, fmt.Sprintf("invalid kern subtable : for (%d, %d) expected %d, got %d", k[0], k[1], exp, got))
	}
}

func TestKerx6(t *testing.T) {
	table, err := td.Files.ReadFile("toys/tables/kerx6Exp-VF.bin")
	tu.AssertNoErr(t, err)

	kerx, _, err := tables.ParseKerx(table, 0xFF)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(kerx.Tables) == 1)

	k, ok := newKerxSubtable(kerx.Tables[0]).Data.(Kern6)
	tu.Assert(t, ok)

	expecteds := []struct { // value extracted from harfbuzz run
		left, right GID
		kerning     int16
	}{
		{283, 659, -270},
		{659, 3, 0},
		{3, 4, 0},
		{4, 333, -130},
		{333, 3, 0},
		{3, 283, 0},
		{283, 815, -230},
		{815, 3, 0},
		{3, 333, 0},
		{333, 573, -150},
		{573, 3, 0},
		{3, 815, 0},
		{815, 283, -170},
		{283, 3, 0},
		{3, 659, 0},
		{659, 283, -270},
		{283, 3, 0},
		{3, 283, 0},
		{283, 650, -270},
	}

	for _, exp := range expecteds {
		got := k.KernPair(exp.left, exp.right)
		tu.Assert(t, got == exp.kerning)
	}
}

func TestMorxLig(t *testing.T) {
	b, err := td.Files.ReadFile("toys/tables/morxLigature.bin")
	tu.AssertNoErr(t, err)
	mt, _, err := tables.ParseMorx(b, 2590)
	tu.AssertNoErr(t, err)
	morx := newMorx(mt)
	tu.Assert(t, len(morx) == 1)
	tu.Assert(t, len(morx[0].Subtables) == 16)

	expectedLigActionLength := []int{36, 60, 24, 36, 52, 78, 21, 39, 16, 42, 252, 2090, 1248, 168, 226}
	for i, st := range morx[0].Subtables[:14] {
		lig, ok := st.Data.(MorxLigatureSubtable)
		tu.Assert(t, ok)
		tu.Assert(t, expectedLigActionLength[i] == len(lig.LigatureAction))
	}
}

func TestKerx1(t *testing.T) {
	b, err := td.Files.ReadFile("toys/tables/kern1.bin")
	tu.AssertNoErr(t, err)
	kt, _, err := tables.ParseKern(b)
	tu.AssertNoErr(t, err)

	kerx := newKernxFromKern(kt)
	tu.Assert(t, len(kerx) == 12)

	expectedEntriesLength := []int{79, 31, 25, 14, 45, 32, 37, 30, 37, 16, 31, 16}
	for i, st := range kerx {
		kern1, ok := st.Data.(Kern1)
		tu.Assert(t, ok)
		tu.Assert(t, expectedEntriesLength[i] == len(kern1.Machine.entries))
	}
}
