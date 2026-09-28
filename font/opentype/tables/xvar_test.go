// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

func deHexStr(s string) []byte {
	s = strings.Join(strings.Split(s, " "), "")
	if len(s)%2 != 0 {
		s += "0"
	}
	out, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return out
}

func TestParseTuple(t *testing.T) {
	// imported from fonttools

	data := deHexStr("DE AD C0 00 20 00 DE AD")
	got, _, err := ParseTuple(data[2:], 2)
	tu.AssertNoErr(t, err)
	expected := Tuple{Values: []Coord{NewCoord(-1), NewCoord(0.5)}}
	tu.AssertC(t, reflect.DeepEqual(got, expected), fmt.Sprintf("%v != %v", got, expected))

	// Shared tuples in the 'gvar' table of the Skia font, as printed
	// in Apple's TrueType specification.
	// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6gvar.html
	skiaGvarSharedTuplesData := deHexStr(
		"40 00 00 00 C0 00 00 00 00 00 40 00 00 00 C0 00 " +
			"C0 00 C0 00 40 00 C0 00 40 00 40 00 C0 00 40 00")

	skiaGvarSharedTuples := SharedTuples{[]Tuple{
		{Values: []Coord{NewCoord(1), NewCoord(0)}},
		{Values: []Coord{NewCoord(-1), NewCoord(0)}},
		{Values: []Coord{NewCoord(0), NewCoord(1)}},
		{Values: []Coord{NewCoord(0), NewCoord(-1)}},
		{Values: []Coord{NewCoord(-1), NewCoord(-1)}},
		{Values: []Coord{NewCoord(1), NewCoord(-1)}},
		{Values: []Coord{NewCoord(1), NewCoord(1)}},
		{Values: []Coord{NewCoord(-1), NewCoord(1)}},
	}}
	sharedTuples, _, err := ParseSharedTuples(skiaGvarSharedTuplesData, 8, 2)
	tu.AssertNoErr(t, err)
	tu.Assert(t, reflect.DeepEqual(sharedTuples, skiaGvarSharedTuples))
}

func TestParseGvar(t *testing.T) {
	// imported from fonttools

	gvarData := deHexStr("0001 0000 " + //   0: majorVersion=1 minorVersion=0
		"0002 0000 " + //   4: axisCount=2 sharedTupleCount=0
		"0000001C " + //   8: offsetToSharedTuples=28
		"0003 0000 " + //  12: glyphCount=3 flags=0
		"0000001C " + //  16: offsetToGlyphVariationData=28
		"0000 0000 000C 002F " + //  20: offsets=[0,0,12,47], times 2: [0,0,24,94],
		//                 //           +offsetToGlyphVariationData: [28,28,52,122]
		//
		// 28: Glyph variation data for glyph //0, ".notdef"
		// ------------------------------------------------
		// (no variation data for this glyph)
		//
		// 28: Glyph variation data for glyph //1, "space"
		// ----------------------------------------------
		"8001 000C " + //  28: tupleVariationCount=1|TUPLES_SHARE_POINT_NUMBERS, offsetToData=12(+28=40)
		"000A " + //  32: tvHeader[0].variationDataSize=10
		"8000 " + //  34: tvHeader[0].tupleIndex=EMBEDDED_PEAK
		"0000 2CCD " + //  36: tvHeader[0].peakTuple={wght:0.0, wdth:0.7}
		"00 " + //  40: all points
		"03 01 02 03 04 " + //  41: deltaX=[1, 2, 3, 4]
		"03 0b 16 21 2C " + //  46: deltaY=[11, 22, 33, 44]
		"00 " + //  51: padding
		//
		// 52: Glyph variation data for glyph //2, "I"
		// -----------------------------------------float32-
		"8002 001c " + //  52: tupleVariationCount=2|TUPLES_SHARE_POINT_NUMBERS, offsetToData=28(+52=80)
		"0012 " + //  56: tvHeader[0].variationDataSize=18
		"C000 " + //  58: tvHeader[0].tupleIndex=EMBEDDED_PEAK|INTERMEDIATE_REGION
		"2000 0000 " + //  60: tvHeader[0].peakTuple={wght:0.5, wdth:0.0}
		"0000 0000 " + //  64: tvHeader[0].intermediateStart={wght:0.0, wdth:0.0}
		"4000 0000 " + //  68: tvHeader[0].intermediateEnd={wght:1.0, wdth:0.0}
		"0016 " + //  72: tvHeader[1].variationDataSize=22
		"A000 " + //  74: tvHeader[1].tupleIndex=EMBEDDED_PEAK|PRIVATE_POINTS
		"C000 3333 " + //  76: tvHeader[1].peakTuple={wght:-1.0, wdth:0.8}
		"00 " + //  80: all points
		"07 03 01 04 01 " + //  81: deltaX.len=7, deltaX=[3, 1, 4, 1,
		"05 09 02 06 " + //  86:                       5, 9, 2, 6]
		"07 03 01 04 01 " + //  90: deltaY.len=7, deltaY=[3, 1, 4, 1,
		"05 09 02 06 " + //  95:                       5, 9, 2, 6]
		"06 " + //  99: 6 points
		"05 00 01 03 01 " + // 100: runLen=5(+1=6); delta-encoded run=[0, 1, 4, 5,
		"01 01 " + // 105:                                    6, 7]
		"05 f8 07 fc 03 fe 01 " + // 107: deltaX.len=5, deltaX=[-8,7,-4,3,-2,1]
		"05 a8 4d 2c 21 ea 0b " + // 114: deltaY.len=5, deltaY=[-88,77,44,33,-22,11]
		"00") // 121: padding
	tu.Assert(t, len(gvarData) == 122)

	gvarDataEmptyVariations := deHexStr("0001 0000 " + //  0: majorVersion=1 minorVersion=0
		"0002 0000 " + //  4: axisCount=2 sharedTupleCount=0
		"0000001c " + //  8: offsetToSharedTuples=28
		"0003 0000 " + // 12: glyphCount=3 flags=0
		"0000001c " + // 16: offsetToGlyphVariationData=28
		"0000 0000 0000 0000") // 20: offsets=[0, 0, 0, 0]
	tu.Assert(t, len(gvarDataEmptyVariations) == 28)

	sharedTuplesExpected := SharedTuples{}
	variationsHeadersExpected := [][]TupleVariationHeader{
		0: {},
		1: {
			{
				VariationDataSize: 0x000A,
				tupleIndex:        0x8000,
				PeakTuple:         Tuple{[]Coord{0, 0x2ccd}},
			},
		},
		2: {
			{
				VariationDataSize: 0x0012,
				tupleIndex:        0xC000,
				PeakTuple:         Tuple{[]Coord{NewCoord(0.5), 0}},
				IntermediateTuples: [2]Tuple{
					{[]Coord{0, 0}},
					{[]Coord{NewCoord(1), 0}},
				},
			},
			{
				VariationDataSize: 0x0016,
				tupleIndex:        0xA000,
				PeakTuple:         Tuple{[]Coord{NewCoord(-1), NewCoord(0.8)}},
			},
		},
	}

	gvarEmptyVariationsExpected := make([][]TupleVariationHeader, 3)

	out, _, err := ParseGvar(gvarData)
	tu.AssertNoErr(t, err)
	tu.Assert(t, reflect.DeepEqual(sharedTuplesExpected, out.SharedTuples))
	tu.Assert(t, len(variationsHeadersExpected) == len(out.GlyphVariationDatas))
	for i, exp := range variationsHeadersExpected {
		got := out.GlyphVariationDatas[i].TupleVariationHeaders
		tu.AssertC(t, fmt.Sprintf("%v", exp) == fmt.Sprintf("%v", got), fmt.Sprintf("%v != %v", exp, got))
	}

	out, _, err = ParseGvar(gvarDataEmptyVariations)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(gvarEmptyVariationsExpected) == len(out.GlyphVariationDatas))
	for i, exp := range gvarEmptyVariationsExpected {
		tu.Assert(t, reflect.DeepEqual(exp, out.GlyphVariationDatas[i].TupleVariationHeaders))
	}
}

func TestParseGvar2(t *testing.T) {
	for _, filepath := range []string{
		"common/Commissioner-VF.ttf",
		"common/Mada-VF.ttf",
	} {
		fp := readFontFile(t, filepath)
		_, _, err := ParseGvar(readTable(t, fp, "gvar"))
		tu.AssertNoErr(t, err)
	}
}

func TestParseHvar(t *testing.T) {
	for _, filepath := range []string{
		"common/Commissioner-VF.ttf",
		"common/Selawik-VF.ttf",
	} {
		fp := readFontFile(t, filepath)
		_, _, err := ParseHVAR(readTable(t, fp, "HVAR"))
		tu.AssertNoErr(t, err)
	}
}

// copied from harfbuzz/src/test-item-varstore.cc
func TestParseHvarData(t *testing.T) {
	// HVAR table data from SourceSerif4Variable-Roman_subset.otf
	const hvar_data = "\x00\x01\x00\x00\x00\x00\x00\x14\x00\x00\x00\xc4\x00\x00\x00\x00" +
		"\x00\x00\x00\x00\x00\x01\x00\x00\x00\x10\x00\x02\x00\x00\x00\x74\x00\x00\x00\x7a\x00" +
		"\x02\x00\x08\xc0\x00\xc0\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x40\x00\x40\x00" +
		"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xc0\x00\xc0\x00\x00\x00\x00\x00\x00" +
		"\x00\x00\x00\x00\x00\x40\x00\x40\x00\xc0\x00\xc0\x00\x00\x00\xc0\x00\xc0\x00\x00\x00" +
		"\xc0\x00\xc0\x00\x00\x00\x00\x00\x40\x00\x40\x00\x00\x00\x40\x00\x40\x00\xc0\x00\xc0" +
		"\x00\x00\x00\x00\x00\x40\x00\x40\x00\x00\x00\x40\x00\x40\x00\x00\x01\x00\x00\x00\x00" +
		"\x00\x04\x00\x00\x00\x08\x00\x00\x00\x01\x00\x02\x00\x03\x00\x04\x00\x05\x00\x06\x00\x07\xf9" +
		"\x0f\x2f\xbf\xfb\xfb\x35\xf9\x04\x04\xf3\xb4\xf2\xfb\x2e\xf3\x04\x04\x0e\xad\xfa\x01\x1a\x01\x15" +
		"\x22\x59\xd6\xe3\xf6\x06\xf5\x00\x01\x00\x05\x00\x04\x07\x05\x06"

	hvar, _, err := ParseHVAR([]byte(hvar_data))
	tu.AssertNoErr(t, err)
	tu.Assert(t, hvar.ItemVariationStore.AxisCount() == 2)
}

func TestParseAvar(t *testing.T) {
	for _, filepath := range td.WithAvar {
		fp := readFontFile(t, filepath)
		_, _, err := ParseAvar(readTable(t, fp, "avar"))
		tu.AssertNoErr(t, err)
	}
}

func TestParseMVAR(t *testing.T) {
	for _, filepath := range td.WithMVAR {
		fp := readFontFile(t, filepath)
		_, _, err := ParseMVAR(readTable(t, fp, "MVAR"))
		tu.AssertNoErr(t, err)
	}
}

func TestParseFvar(t *testing.T) {
	for _, item := range td.WithFvar {
		fp := readFontFile(t, item.Path)
		fvar, _, err := ParseFvar(readTable(t, fp, "fvar"))
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(fvar.Axis) == item.AxisCount)
	}
}

func TestCoordBits(t *testing.T) {
	tu.Assert(t, NewCoord(1) == Coord(1<<14))
	tu.Assert(t, NewCoord(-1) == -Coord(1<<14))

	tu.Assert(t, abs(NewCoord(1)) == abs(NewCoord(-1)))
	tu.Assert(t, abs(NewCoord(0.123)) < abs(NewCoord(0.56)))
	tu.Assert(t, abs(NewCoord(0.123)) < abs(NewCoord(-0.56)))
	tu.Assert(t, abs(NewCoord(-0.123)) < abs(NewCoord(-0.56)))
}

func TestParseSTAT(t *testing.T) {
	for _, path := range td.WithAvar {
		fp := readFontFile(t, path)

		names, _, err := ParseName(readTable(t, fp, "name"))
		tu.AssertNoErr(t, err)

		stat, _, err := ParseSTAT(readTable(t, fp, "STAT"))
		tu.AssertNoErr(t, err)

		for _, axis := range stat.designAxes {
			tu.Assert(t, names.Name(axis.NameID) != "")
		}
	}
}

func TestItemVarStoreOutOfRange(t *testing.T) {
	store := ItemVarStore{
		format: 1,
		VariationRegionList: VariationRegionList{axisCount: 1, VariationRegions: []VariationRegion{
			{RegionAxes: []RegionAxisCoordinates{{StartCoord: -1, PeakCoord: 1, EndCoord: 1}}},
		}},
		ItemVariationDatas: []ItemVariationData{{RegionIndexes: []uint16{3}, DeltaSets: [][]int16{{10}}}},
	}
	// region index 3 is out of range
	tu.Assert(t, store.GetDelta(VariationStoreIndex{}, []Coord{1}) == 0)
	// more coordinates than region axes
	store.ItemVariationDatas[0].RegionIndexes[0] = 0
	tu.Assert(t, store.GetDelta(VariationStoreIndex{}, []Coord{1, 1}) == 10)
}

func TestFvarNamedInstanceStride(t *testing.T) {
	for _, size := range []int{8, 10} {
		src := make([]byte, 36+2*size)
		put16 := func(off int, v uint16) { binary.BigEndian.PutUint16(src[off:], v) }
		put32 := func(off int, v uint32) { binary.BigEndian.PutUint32(src[off:], v) }
		put16(0, 1)
		put16(4, 16)
		put16(6, 2)
		put16(8, 1)
		put16(10, 20)
		put16(12, 2)
		put16(14, uint16(size))
		copy(src[16:], "wght")
		put32(20, 1<<16)
		put32(24, 1<<16)
		put32(28, 2<<16)
		put16(34, 256)
		for i := 0; i < 2; i++ {
			start := 36 + i*size
			put16(start, uint16(257+i))
			put32(start+4, uint32(i+1)<<16)
			if size == 10 {
				put16(start+8, uint16(300+i))
			}
		}
		got, _, err := ParseFvar(src)
		if err != nil {
			t.Fatal(err)
		}
		for i, instance := range got.Instances {
			name := uint16(0)
			if size == 10 {
				name = uint16(300 + i)
			}
			if instance.SubfamilyNameID != uint16(257+i) || instance.Coordinates[0] != float32(i+1) || instance.PostScriptNameID != name {
				t.Fatalf("record size %d, instance %d: wrong record %+v", size, i, instance)
			}
		}
		if _, _, err := ParseFvar(src[:len(src)-1]); err == nil {
			t.Fatal("accepted truncated instances")
		}
		put16(14, 7)
		if _, _, err := ParseFvar(src); err == nil {
			t.Fatal("accepted undersized instance records")
		}
	}
}

func TestFvarInstanceBounds(t *testing.T) {
	src := make([]byte, 65535)
	for _, test := range []struct {
		name        string
		count, size int
	}{
		{"header product overflows int32", 65535, 65535},
		{"product overflows int", 2, int(^uint(0) >> 1)},
		{"negative count", -1, 8},
		{"negative size", 1, -1},
		{"zero size", 1, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := ParseFvarRecords(src, 0, test.count, test.size); err == nil {
				t.Fatal("accepted invalid instance dimensions")
			}
		})
	}
	for _, size := range []int{0, 8, 65535} {
		if _, _, err := ParseFvarRecords(nil, 0, 0, size); err != nil {
			t.Fatalf("empty instance list with size %d: %v", size, err)
		}
	}
}
