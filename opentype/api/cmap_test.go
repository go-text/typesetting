// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package api

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
	tu "github.com/go-text/typesetting/opentype/testutils"

	td "github.com/go-text/typesetting-utils/opentype"
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

// check for crashes, return the number of glyphs
func loopThroughCmap(cmap Cmap) int {
	var nbGlyphs int
	iter := cmap.Iter()
	for iter.Next() {
		_, _ = iter.Char()
		nbGlyphs++
	}
	return nbGlyphs
}

func TestCmap(t *testing.T) {
	for _, filename := range append(tu.Filenames(t, "common"), tu.Filenames(t, "cmap")...) {
		fp := readFontFile(t, filename)
		cmapT, _, err := tables.ParseCmap(readTable(t, fp, "cmap"))
		tu.AssertNoErr(t, err)
		cmap, _, err := ProcessCmap(cmapT, tables.FPNone)
		tu.AssertNoErr(t, err)
		tu.Assert(t, cmap != nil)
		tu.Assert(t, loopThroughCmap(cmap) > 0)
	}

	for _, filename := range tu.Filenames(t, "cmap/table") {
		table, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)

		cmapT, _, err := tables.ParseCmap(table)
		tu.AssertNoErr(t, err)
		cmap, _, err := ProcessCmap(cmapT, tables.FPNone)
		tu.AssertNoErr(t, err)
		tu.Assert(t, cmap != nil)
		tu.Assert(t, loopThroughCmap(cmap) > 0)
	}
}

func TestCmap4(t *testing.T) {
	d1, d2, d3 := int16(-9), int16(-18), int16(-80)
	input := []uint16{
		0, 0, 0, // start of subtable
		8,
		8, 4, 0,
		20, 90, 480, 0xffff,
		0, // reserved pad
		10, 30, 153, 0xffff,
		uint16(d1), uint16(d2), uint16(d3), 1,
		0, 0, 0, 0,
	}
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, input)
	tu.AssertNoErr(t, err)

	cmapT, _, err := tables.ParseCmapSubtable4(buf.Bytes())
	tu.AssertNoErr(t, err)

	cmap, err := newCmap4(cmapT)
	tu.AssertNoErr(t, err)

	runes := [...]rune{10, 20, 30, 90, 153, 480, 0xFFFF}
	glyphs := [...]GID{1, 11, 12, 72, 73, 400, 0}
	for i, r := range runes {
		got, _ := cmap.Lookup(r)
		tu.Assert(t, got == glyphs[i])
	}
}

func TestBestEncoding(t *testing.T) {
	filename := "toys/3cmaps.ttc"
	file, err := td.Files.ReadFile(filename)
	tu.AssertNoErr(t, err)

	fs, err := loader.NewLoaders(bytes.NewReader(file))
	tu.AssertNoErr(t, err)

	font := fs[0]
	cmaps, _, err := tables.ParseCmap(readTable(t, font, "cmap"))
	tu.AssertNoErr(t, err)

	tu.Assert(t, len(cmaps.Records) == 3)
	cmap, _, err := ProcessCmap(cmaps, tables.FPNone)
	tu.AssertNoErr(t, err)

	_, ok := cmap.Lookup(0x2026)
	tu.Assert(t, ok)
	_, ok = cmap.Lookup(0xFFFFFFF)
	tu.Assert(t, !ok)
}

func TestCmap12(t *testing.T) {
	font := readFontFile(t, "cmap/CMAP12.otf")
	cmaps, _, err := tables.ParseCmap(readTable(t, font, "cmap"))
	tu.AssertNoErr(t, err)

	cmap, _, err := ProcessCmap(cmaps, tables.FPNone)
	tu.AssertNoErr(t, err)

	runes := [...]rune{
		0x0011, 0x0012, 0x0013, 0x0014, 0x0015, 0x0016, 0x0017, 0x0018,
	}
	gids := [...]GID{
		17, 18, 19, 20, 21, 22, 23, 24,
	}

	for i, r := range runes {
		got, _ := cmap.Lookup(r)
		tu.Assert(t, got == gids[i])
	}
}

func TestCmap14(t *testing.T) {
	font := readFontFile(t, "cmap/CMAP14.otf")
	cmaps, _, err := tables.ParseCmap(readTable(t, font, "cmap"))
	tu.AssertNoErr(t, err)

	_, uv, err := ProcessCmap(cmaps, tables.FPNone)
	tu.AssertNoErr(t, err)

	gid, flag := uv.GetGlyphVariant(33446, 917761)
	tu.Assert(t, flag == VariantFound)
	tu.Assert(t, gid == 2)

	_, flag = uv.GetGlyphVariant(33446, 0xF)
	tu.Assert(t, flag == VariantNotFound)
}
