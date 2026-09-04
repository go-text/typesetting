// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package cff

import (
	"bytes"
	"encoding/binary"
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

// TestCFF2LocalSubrs covers a CFF2 font whose Private DICT declares local
// subroutines. The Subrs offset is relative to the Private DICT, not to the
// start of the table, so resolving it from the wrong base reads an arbitrary
// position: with this fixture that lands on a byte giving offSize 140 and the
// table is rejected outright, and on other fonts it can present a count of
// hundreds of millions, or an offSize of 0.
//
// Because NewFont discards the error from loadCff2, a font in that state loads
// with no CFF2 outlines at all rather than failing, which is why this went
// unnoticed: the only other CFF2 fixture, NotoSansCJKjp-VF.otf, declares no
// local subroutines and so never reaches the code.
func TestCFF2LocalSubrs(t *testing.T) {
	b, err := td.Files.ReadFile("toys/CFF2-VF.otf")
	tu.AssertNoErr(t, err)

	ft, err := ot.NewLoader(bytes.NewReader(b))
	tu.AssertNoErr(t, err)

	table, err := ft.RawTable(ot.MustNewTag("CFF2"))
	tu.AssertNoErr(t, err)

	out, err := ParseCFF2(table)
	tu.AssertNoErr(t, err)

	// the fixture declares local subroutines; if it stops doing so this test
	// no longer covers anything and should be pointed at another font
	nSubrs := 0
	for _, f := range out.fonts {
		nSubrs += len(f.localSubrs)
	}
	tu.Assert(t, nSubrs != 0)

	// every glyph must interpret, which is what actually exercises the subrs:
	// a charstring calling callsubr with a wrongly located INDEX fails here
	tu.Assert(t, len(out.Charstrings) != 0)
	for i := range out.Charstrings {
		_, _, err := out.LoadGlyph(uint16(i), nil)
		tu.AssertNoErr(t, err)
	}
}

// buildCFF2LocalSubrs assembles a minimal CFF2 table whose Private DICT
// declares local subroutines.
//
// It is synthetic rather than a font file so the case can be pinned exactly:
// the Subrs offset is placed so that resolving it from the start of the table,
// rather than from the Private DICT, lands inside the Top DICT and yields a
// byte that is not a legal offSize. Real fonts differ only in what the wrong
// address happens to contain.
func buildCFF2LocalSubrs() []byte {
	be32 := func(v uint32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, v); return b }
	dictInt := func(v int32) []byte { return append([]byte{29}, be32(uint32(v))...) }

	// a CFF2 INDEX: count uint32, offSize uint8, count+1 offsets, then data
	index := func(items ...[]byte) []byte {
		out := be32(uint32(len(items)))
		if len(items) == 0 {
			return out
		}
		out = append(out, 1) // offSize
		off := byte(1)
		offs := []byte{off}
		for _, it := range items {
			off += byte(len(it))
			offs = append(offs, off)
		}
		out = append(out, offs...)
		for _, it := range items {
			out = append(out, it...)
		}
		return out
	}

	const headerSize = 5
	const topDictLen = 6 + 7 // CharStrings, then FDArray
	const privDictLen = 6    // Subrs
	charStringsOff := headerSize + topDictLen
	charStrings := index([]byte{0x0b}) // one glyph: return
	fdArrayOff := charStringsOff + len(charStrings)

	fontDict := func(privOff int) []byte {
		fd := append(dictInt(privDictLen), dictInt(int32(privOff))...)
		return append(fd, 18) // Private [size, offset]
	}
	// the Font DICT's length does not depend on the value, so the offsets
	// stay put when it is filled in
	privDictOff := fdArrayOff + len(index(fontDict(0)))
	fdArray := index(fontDict(privDictOff))

	// "The local subrs offset is relative to the beginning of the Private
	// DICT data", so the INDEX sits immediately after it and the operand is
	// the DICT's own length
	privDict := append(dictInt(privDictLen), 19) // Subrs
	localSubrs := index([]byte{0x0b, 0x0b}, []byte{0x0b, 0x0b, 0x0b})

	out := []byte{2, 0, headerSize, 0, 0}
	binary.BigEndian.PutUint16(out[3:], topDictLen)
	top := append(dictInt(int32(charStringsOff)), 17)
	top = append(top, dictInt(int32(fdArrayOff))...)
	top = append(top, 12, 36)
	out = append(out, top...)
	out = append(out, charStrings...)
	out = append(out, fdArray...)
	out = append(out, privDict...)
	return append(out, localSubrs...)
}

// TestCFF2LocalSubrsOffset pins that a Private DICT's Subrs operand is resolved
// relative to the Private DICT and not to the start of the table.
//
// Reading it from the wrong base gives whatever happens to be at that address.
// Here it is a byte inside the Top DICT that is not a legal offSize, so the
// table is rejected; on the fonts that prompted this it was an offSize of 0 or
// a count of tens of millions. Because NewFont discards the error from
// loadCff2, the visible effect in all of those cases is a font that loads with
// no CFF2 outlines at all.
func TestCFF2LocalSubrsOffset(t *testing.T) {
	out, err := ParseCFF2(buildCFF2LocalSubrs())
	tu.AssertNoErr(t, err)

	tu.Assert(t, len(out.fonts) == 1)
	subrs := out.fonts[0].localSubrs
	tu.Assert(t, len(subrs) == 2)
	tu.Assert(t, len(subrs[0]) == 2 && len(subrs[1]) == 3)
}

// TestParseIndexContentBounds guards against a malformed INDEX header driving
// a huge allocation. The CFF2 INDEX header parser does not validate offSize
// (unlike CFF1), so a font can present offSize 0 — which zeroes the length
// check — or a 32-bit count far larger than the data; parseIndexContent must
// reject these instead of make([][]byte, count)-ing gigabytes.
func TestParseIndexContentBounds(t *testing.T) {
	// Minimal valid INDEX: count 1, offSize 1, offsets [1,2], one data byte.
	valid := []byte{0x01, 0x02, 0xAA}
	out, _, err := parseIndexContent(valid, indexStart{count: 1, offSize: 1})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out) == 1 && len(out[0]) == 1 && out[0][0] == 0xAA)

	// offSize 0 must be rejected before the allocation.
	_, _, err = parseIndexContent(valid, indexStart{count: 1 << 20, offSize: 0})
	tu.Assert(t, err != nil)

	// A count larger than the data can address must be rejected (also covers
	// the count+1 uint32 wrap at 0xFFFFFFFF).
	_, _, err = parseIndexContent(valid, indexStart{count: 0xFFFFFFFF, offSize: 1})
	tu.Assert(t, err != nil)

	// count exactly len(src)/offSize: the offset array still needs one more
	// offset than the data holds, so this must be an error rather than a
	// slice-out-of-range panic on src[offsetArraySize:].
	_, _, err = parseIndexContent(valid, indexStart{count: 3, offSize: 1})
	tu.Assert(t, err != nil)

	// offSize > 4 is out of spec and must be rejected.
	_, _, err = parseIndexContent(valid, indexStart{count: 1, offSize: 5})
	tu.Assert(t, err != nil)
}
