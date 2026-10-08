// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"bytes"
	"encoding/binary"
	ot "github.com/go-text/typesetting/font/opentype"
	"os"
	"sort"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"
)

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

func TestEBDTFormat1(t *testing.T) {
	file, err := td.Files.ReadFile("bitmap/simsun.ttc")
	tu.AssertNoErr(t, err)

	faces, err := ParseTTC(bytes.NewReader(file))
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(faces) == 2)
	sizes := faces[0].BitmapSizes()
	tu.Assert(t, len(sizes) == 6)
	tu.Assert(t, sizes[0].XPpem == 12 && sizes[5].XPpem == 17)
}

func TestBitmapInvalidGlyphRange(t *testing.T) {
	_, err := newBitmapSubtable(tables.BitmapSubtable{
		FirstGlyph: 5, LastGlyph: 2,
		IndexSubHeader: tables.IndexSubHeader{IndexData: tables.IndexData2{}},
	}, nil)
	tu.Assert(t, err != nil)
}

func TestIndexSubTable5Offsets(t *testing.T) {
	idx, err := parseIndexSubTable5(tables.BitmapSubtable{IndexSubHeader: tables.IndexSubHeader{ImageFormat: 5}},
		tables.IndexData5{ImageSize: 2, GlyphIdArray: []tables.GlyphID{1, 2}}, []byte{1, 2, 3, 4})
	tu.AssertNoErr(t, err)
	tu.Assert(t, bytes.Equal(idx.glyphs[0], []byte{1, 2}))
	tu.Assert(t, bytes.Equal(idx.glyphs[1], []byte{3, 4}))
}

func TestSparseBitmapFallback(t *testing.T) {
	source, err := os.ReadFile("testdata/Roboto-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	base, err := ParseTTF(bytes.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	gid, ok := base.NominalGlyph('A')
	if !ok {
		t.Fatal("missing A")
	}
	ld, err := ot.NewLoader(bytes.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, indexFormat := range []uint16{1, 3} {
		cblc := make([]byte, 80)
		binary.BigEndian.PutUint16(cblc, 3)
		binary.BigEndian.PutUint32(cblc[4:], 1)   // one strike
		binary.BigEndian.PutUint32(cblc[8:], 56)  // index subtable array offset
		binary.BigEndian.PutUint32(cblc[12:], 24) // index tables size
		binary.BigEndian.PutUint32(cblc[16:], 1)  // one index subtable
		binary.BigEndian.PutUint16(cblc[48:], uint16(gid))
		binary.BigEndian.PutUint16(cblc[50:], uint16(gid))
		cblc[52], cblc[53], cblc[54], cblc[55] = 16, 16, 32, 1
		binary.BigEndian.PutUint16(cblc[56:], uint16(gid))
		binary.BigEndian.PutUint16(cblc[58:], uint16(gid))
		binary.BigEndian.PutUint32(cblc[60:], 8)           // relative index subtable offset
		binary.BigEndian.PutUint16(cblc[64:], indexFormat) // sparse index formats
		binary.BigEndian.PutUint16(cblc[66:], 17)          // PNG image format
		binary.BigEndian.PutUint32(cblc[68:], 4)           // skip CBDT version
		// Equal consecutive offsets mean this glyph has no bitmap.
		var ts []ot.Table
		for _, tag := range ld.Tables() {
			raw, err := ld.RawTable(tag)
			if err != nil {
				t.Fatal(err)
			}
			ts = append(ts, ot.Table{Tag: tag, Content: raw})
		}
		ts = append(ts, ot.Table{Tag: ot.MustNewTag("CBLC"), Content: cblc}, ot.Table{Tag: ot.MustNewTag("CBDT"), Content: []byte{0, 3, 0, 0}})
		sort.Slice(ts, func(i, j int) bool { return ts[i].Tag < ts[j].Tag })
		face, err := ParseTTF(bytes.NewReader(ot.WriteOpentype(ts, ot.TrueType)))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := face.GlyphDataOutline(gid); !ok {
			t.Fatal("missing outline")
		}
		if data := face.GlyphData(gid); data != nil {
			if _, ok := data.(GlyphOutline); !ok {
				t.Fatalf("missing bitmap suppresses usable outline: got %T", data)
			}
		} else {
			t.Fatal("no glyph data")
		}
		extents, ok := face.GlyphExtents(gid)
		want, wantOK := base.GlyphExtents(gid)
		if !ok || !wantOK || extents != want {
			t.Fatalf("bitmap index%d: got extents %v, want %v", indexFormat, extents, want)
		}
	}
}
