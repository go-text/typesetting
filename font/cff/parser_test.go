// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package cff

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	psinterpreter "github.com/go-text/typesetting/font/cff/interpreter"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"
)

func TestParseCFF(t *testing.T) {
	for _, filepath := range tu.Filenames(t, "cff") {
		content, err := td.Files.ReadFile(filepath)
		tu.AssertNoErr(t, err)

		font, err := Parse(content)
		tu.AssertNoErr(t, err)

		tu.Assert(t, len(font.Charstrings) >= 12)

		if font.fdSelect != nil {
			for i := 0; i < len(font.Charstrings); i++ {
				_, err = font.fdSelect.fontDictIndex(tables.GlyphID(i))
				tu.AssertNoErr(t, err)
			}
		}

		for glyphIndex := range font.Charstrings {
			_, _, err := font.LoadGlyph(tables.GlyphID(glyphIndex))
			tu.AssertNoErr(t, err)
		}
	}
}

func TestGlyhName(t *testing.T) {
	content, err := td.Files.ReadFile("toys/NamesCFF.ttf")
	tu.AssertNoErr(t, err)

	ft, err := ot.NewLoader(bytes.NewReader(content))
	tu.AssertNoErr(t, err)

	table, err := ft.RawTable(ot.MustNewTag("CFF "))
	tu.AssertNoErr(t, err)

	cff, err := Parse(table)
	tu.AssertNoErr(t, err)

	expectedCharset := []uint16{
		0x0, 0x1, 0x187, 0x188, 0x189, 0x18a, 0x18b, 0x18c, 0x18d, 0x18e, 0x18f, 0x190, 0x191,
		0x192, 0x193, 0x194, 0x195, 0x196, 0x197, 0x198, 0x199, 0x19a, 0x19b, 0x19c, 0x19d, 0x19e, 0x19f, 0x1a0,
		0x1a1, 0x1a2, 0x1a3, 0x1a4, 0x1a5, 0x1a6, 0x1a7, 0x1a8, 0x1a9, 0x1aa, 0x1ab, 0x1ac, 0x1ad, 0x1ae, 0x1af,
		0x1b0, 0x1b1, 0x1b2, 0x1b3, 0x1b4, 0x1b5, 0x1b6, 0x1b7, 0x1b8, 0x1b9, 0x1ba, 0x1bb, 0x1bc, 0x1bd, 0x1be,
		0x1bf, 0x1c0, 0x1c1, 0x1c2, 0x1c3, 0x1c4, 0x1c5, 0x1c6, 0x1c7, 0x1c8, 0x1c9, 0x1ca, 0x1cb, 0x1cc, 0x1cd,
		0x1ce, 0x1cf, 0x1d0, 0x1d1,
	}
	tu.Assert(t, reflect.DeepEqual(expectedCharset, cff.charset))

	expectedUserStrings := []string{
		"uni0622", "uni0623", "uni0624", "uni0625", "uni0626", "uni0628", "uni06C0",
		"uni06C2", "uni06D3", "uni0625.fina", "uni0623.fina", "uni0622.fina", "uni0628.fina", "uni0626.init",
		"uni0628.init", "uni0626.medi", "uni0628.medi", "uni06C1.fina", "uni06D5.fina", "uni06C1.init", "uni06C1.medi",
		"uni0624.fina", "uni0626.fina", "uni0626.init_BaaBaaYaa", "uni0628.init_BaaBaaYaa", "uni0626.medi_BaaBaaYaa",
		"uni0628.medi_BaaBaaYaa", "uni0626.fina_BaaBaaYaa", "uni0626.medi_BaaBaaInit", "uni0628.medi_BaaBaaInit",
		"uni0626.init_BaaBaaIsol", "uni0628.init_BaaBaaIsol", "uni0628.fina_BaaBaaIsol", "uni0626.medi_BaaYaaFina",
		"uni0628.medi_BaaYaaFina", "uni0626.fina_BaaYaaFina", "uni0626.init_High", "uni0628.init_High",
		"uni0626.medi_High", "uni0628.medi_High", "uni0626.init_Wide", "uni0628.init_Wide", "uni0626.init_BaaYaaIsol",
		"uni0628.init_BaaYaaIsol", "uni06C1.init_HehYaaIsol", "uni0626.fina_KafYaaIsol", "uni0626.init_BaaHehInit",
		"uni0628.init_BaaHehInit", "uni0626.medi_BaaHehMedi", "uni0628.medi_BaaHehMedi", "uni06C1.medi_BaaHehMedi",
		"uni0625.LowHamza", "uni0628.init_LD", "uni0628.init_BaaBaaYaaLD", "uni0628.init_BaaBaaIsolLD", "uni0628.init_HighLD",
		"uni0628.init_WideLD", "uni0628.init_BaaYaaIsolLD", "uni0628.init_BaaHehInitLD", "uni06D2.fina", "uni0626.init_YaaBari",
		"uni0628.init_YaaBari", "uni06D2.fina_PostAscender", "uni06D2.fina_PostAyn", "uni06C1.init_YaaBari", "uni06C1.medi_HehYaaFina",
		"hamza.above", "uni0626.init_BaaBaaHeh", "uni0628.init_BaaBaaHeh", "uni0628.init_BaaBaaHehLD", "uni0626.medi_YaaBari",
		"uni0628.medi_YaaBari", "uni06D2.fina_PostToothFina", "uni0626.init_BaaBaaYaaBari", "uni0628.init_BaaBaaYaaBari",
		"0.114", "", "Copyright 2010-2021 The Amiri Project Authors https:github.comaliftypeamiri.", "Amiri",
	}
	gotUserStrings := make([]string, len(cff.userStrings))
	for i, b := range cff.userStrings {
		gotUserStrings[i] = string(b)
	}
	tu.Assert(t, reflect.DeepEqual(expectedUserStrings, gotUserStrings))
}

func TestType2Extent(t *testing.T) {
	// regression test for a bug discovered in https://github.com/go-text/render/pull/8

	content, err := td.Files.Open("toys/tables/cff_with_fixed.json")
	tu.AssertNoErr(t, err)

	var data struct {
		Charstring  []byte
		LocalSubrs  [][]byte
		GlobalSubrs [][]byte
	}
	err = json.NewDecoder(content).Decode(&data)
	tu.AssertNoErr(t, err)

	var (
		loader type2CharstringHandler
		psi    psinterpreter.Machine
	)
	err = psi.Run(data.Charstring, data.LocalSubrs, data.GlobalSubrs, &loader)
	tu.AssertNoErr(t, err)

	extents := loader.cs.Bounds.ToExtents()
	tu.Assert(t, 0 <= extents.Width && extents.Width <= 1000 && -1000 <= extents.Height && extents.Height <= 0)
}

func TestParseCFF2(t *testing.T) {
	b, err := td.Files.ReadFile("common/NotoSansCJKjp-VF.otf")
	tu.AssertNoErr(t, err)

	ft, err := ot.NewLoader(bytes.NewReader(b))
	tu.AssertNoErr(t, err)

	table, err := ft.RawTable(ot.MustNewTag("CFF2"))
	tu.AssertNoErr(t, err)

	out, err := ParseCFF2(table)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Charstrings) == 0xFFFF)
	tu.Assert(t, len(out.VarStore.ItemVariationDatas) == 1)
	tu.Assert(t, len(out.VarStore.VariationRegionList.VariationRegions[0].RegionAxes) == 1)

	for i := range out.Charstrings {
		_, _, err := out.LoadGlyph(uint16(i), []tables.Coord{tables.NewCoord(0.5)})
		tu.AssertNoErr(t, err)

		_, _, err = out.LoadGlyph(uint16(i), nil) // with no variation activated
		tu.AssertNoErr(t, err)
	}
}

func TestIssue122(t *testing.T) {
	b, err := td.Files.ReadFile("common/NotoSansCJKjp-VF.otf")
	tu.AssertNoErr(t, err)

	ft, err := ot.NewLoader(bytes.NewReader(b))
	tu.AssertNoErr(t, err)

	table, err := ft.RawTable(ot.MustNewTag("CFF2"))
	tu.AssertNoErr(t, err)

	out, err := ParseCFF2(table)
	tu.AssertNoErr(t, err)

	// check that the correct number of segments are
	// computed, even if the user has not activated variations
	segments, _, _ := out.LoadGlyph(38, nil)
	tu.Assert(t, len(segments) == 12)
}
