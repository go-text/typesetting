// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"bytes"
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
