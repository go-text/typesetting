// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"

	td "github.com/go-text/typesetting-utils/opentype"
)

// check for crashes, return the number of glyphs
func loopThroughCmap(cmap Cmap) int {
	var nbGlyphs int
	iter := cmap.Iter()
	for iter.Next() {
		_, _ = iter.Char()
		nbGlyphs++
	}

	if cmap, ok := cmap.(CmapRuneRanger); ok {
		_ = cmap.RuneRanges(nil) // check for crashes
	}
	return nbGlyphs
}

func TestLegacyArabicIterationIncludesAliases(t *testing.T) {
	base := cmap12{
		{StartCharCode: 0x0627, EndCharCode: 0x0627, StartGlyphID: 9},
		{StartCharCode: 0xF000, EndCharCode: 0xF2FF, StartGlyphID: 1},
		{StartCharCode: 0x10000, EndCharCode: 0x10000, StartGlyphID: 7},
	}
	for name, cmap := range map[string]Cmap{"simplified": remaperPUASimp{base}, "traditional": remaperPUATrad{base}} {
		t.Run(name, func(t *testing.T) {
			seen := make(map[rune]GID)
			iter := cmap.Iter()
			for iter.Next() {
				r, gid := iter.Char()
				if _, ok := seen[r]; ok {
					t.Fatalf("duplicate rune %U", r)
				}
				seen[r] = gid
				if want, ok := cmap.Lookup(r); !ok || gid != want {
					t.Fatalf("%U: iteration returned %d, lookup returned %d, %v", r, gid, want, ok)
				}
			}
			if seen[0x0627] != 9 || seen[0x0628] == 0 || seen[0x10000] != 7 {
				t.Fatal("missing Arabic alias or original glyph")
			}
			if _, ok := seen[0x0629]; !ok {
				t.Fatal("missing additional Arabic alias")
			}
		})
	}
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

	fs, err := ot.NewLoaders(bytes.NewReader(file))
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

func TestRuneRanges(t *testing.T) {
	for _, filename := range append(tu.Filenames(t, "common"), tu.Filenames(t, "cmap")...) {
		fp := readFontFile(t, filename)
		cmapT, _, err := tables.ParseCmap(readTable(t, fp, "cmap"))
		tu.AssertNoErr(t, err)
		cmap, _, err := ProcessCmap(cmapT, tables.FPNone)
		tu.AssertNoErr(t, err)
		tu.Assert(t, cmap != nil)

		assertRuneRangesEqual(t, cmap)
	}
}

func assertRuneRangesEqual(t *testing.T, cm Cmap) {
	if _, ok := cm.(CmapRuneRanger); !ok {
		return
	}

	byRanges, byIter := make(map[rune]bool), make(map[rune]bool)

	iter := cm.Iter()
	for iter.Next() {
		r, _ := iter.Char()
		byIter[r] = true
	}

	for _, ran := range cm.(CmapRuneRanger).RuneRanges(nil) {
		for r := ran[0]; r <= ran[1]; r++ {
			byRanges[r] = true
		}
	}

	if !reflect.DeepEqual(byRanges, byIter) {
		t.Fatal("inconsistent rune ranges")
	}
}

func TestMacromanCmap(t *testing.T) {
	ld := readFontFile(t, "cmap/Brushstroke-Plain.otf")
	ft, err := NewFont(ld)
	tu.AssertNoErr(t, err)
	_, ok := ft.Cmap.(remaperMacroman)
	tu.Assert(t, ok)
}

func TestCmap4InvalidRangeOffset(t *testing.T) {
	// an idRangeOffset that points before the glyph array is an error, not a panic
	_, err := newCmap4(tables.CmapSubtable4{
		EndCode: []uint16{10, 0xFFFF}, StartCode: []uint16{10, 0xFFFF},
		IdDelta: []uint16{0, 1}, IdRangeOffsets: []uint16{2, 0},
	})
	tu.Assert(t, err != nil)

	// newCmap4 ignores a 0xFFFF offset but uses a real one on a segment starting at 0xFFFF
	cm, err := newCmap4(tables.CmapSubtable4{
		EndCode: []uint16{10, 0xFFFF}, StartCode: []uint16{10, 0xFFFF},
		IdDelta: []uint16{1, 0}, IdRangeOffsets: []uint16{0xFFFF, 2},
		GlyphIDArray: []byte{0, 5},
	})
	tu.AssertNoErr(t, err)
	g, _ := cm.Lookup(10)
	tu.Assert(t, g == 11)
	g, _ = cm.Lookup(0xFFFF)
	tu.Assert(t, g == 5)
}

func TestCmap4IteratorDelta(t *testing.T) {
	raw := make([]byte, 48)
	binary.BigEndian.PutUint16(raw[2:], 1)
	binary.BigEndian.PutUint16(raw[4:], 3)
	binary.BigEndian.PutUint16(raw[6:], 1)
	binary.BigEndian.PutUint32(raw[8:], 12)
	for i, v := range []uint16{4, 36, 0, 4, 4, 1, 0, 'A', 0xFFFF, 0, 'A', 0xFFFF, 0xFFFF, 1, 4, 0, 2} {
		binary.BigEndian.PutUint16(raw[12+2*i:], v)
	}
	tb, _, err := tables.ParseCmap(raw)
	if err != nil {
		t.Fatal(err)
	}
	cm, _, err := ProcessCmap(tb, tables.FPNone)
	if err != nil {
		t.Fatal(err)
	}
	gid, _ := cm.Lookup('A')
	iter := cm.Iter()
	if !iter.Next() {
		t.Fatal("empty")
	}
	_, iterGID := iter.Char()
	if gid != iterGID {
		t.Fatalf("Lookup = %d, Iter = %d", gid, iterGID)
	}
}

func TestCmap0MissingGlyph(t *testing.T) {
	raw := make([]byte, 274)
	binary.BigEndian.PutUint16(raw[2:], 1)
	binary.BigEndian.PutUint16(raw[4:], 1) // Mac platform, Roman encoding
	binary.BigEndian.PutUint32(raw[8:], 12)
	binary.BigEndian.PutUint16(raw[14:], 262)
	raw[18+'A'] = 1
	raw[18] = 2 // byte zero may have a real mapping
	tb, _, err := tables.ParseCmap(raw)
	if err != nil {
		t.Fatal(err)
	}
	cm, _, err := ProcessCmap(tb, tables.FPNone)
	if err != nil {
		t.Fatal(err)
	}
	if gid, ok := cm.Lookup(0); !ok || gid != 2 {
		t.Fatalf("byte zero mapping: got (%d, %v)", gid, ok)
	}
	if gid, ok := cm.Lookup('B'); ok {
		t.Fatalf("unsupported B reported present with GID %d", gid)
	}
}
