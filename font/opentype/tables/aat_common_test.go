// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"reflect"
	"strings"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

func TestAATLookup4(t *testing.T) {
	// adapted from fontttools
	src := deHexStr(
		"0004 0006 0003 000C 0001 0006 " +
			"0002 0001 001E " + // glyph 1..2: mapping at offset 0x1E
			"0005 0004 001E " + // glyph 4..5: mapping at offset 0x1E
			"FFFF FFFF FFFF " + // end of search table
			"0007 0008")
	class, _, err := ParseAATLookup(src, 4)
	tu.AssertNoErr(t, err)
	gids := []GlyphID{1, 2, 4, 5}
	classes := []uint16{7, 8, 7, 8}
	for i, gid := range gids {
		c, ok := class.Class(gid)
		tu.Assert(t, ok)
		tu.Assert(t, c == classes[i])
	}
	_, found := class.Class(0xFF)
	tu.Assert(t, !found)

	// extracted from macos Tamil MN font
	src = []byte{0, 4, 0, 6, 0, 5, 0, 24, 0, 2, 0, 6, 0, 151, 0, 129, 0, 42, 0, 156, 0, 153, 0, 88, 0, 163, 0, 163, 0, 96, 1, 48, 1, 48, 0, 98, 255, 255, 255, 255, 0, 100, 0, 4, 0, 10, 0, 11, 0, 12, 0, 13, 0, 14, 0, 15, 0, 16, 0, 17, 0, 18, 0, 19, 0, 20, 0, 21, 0, 22, 0, 23, 0, 24, 0, 25, 0, 26, 0, 27, 0, 28, 0, 29, 0, 30, 0, 31, 0, 5, 0, 6, 0, 7, 0, 8, 0, 9, 0, 32}
	class, _, err = ParseAATLookup(src, 0xFFFF)
	tu.AssertNoErr(t, err)
	gids = []GlyphID{132, 129, 144, 145, 146, 140, 137, 130, 135, 138, 133, 139, 142, 143, 136, 134, 147, 141, 151, 132, 150, 148, 149, 304, 153, 154, 163, 155, 156}
	classes = []uint16{
		12, 4, 24, 25, 26, 20, 17, 10, 15, 18, 13, 19, 22, 23, 16, 14, 27, 21, 31, 12, 30, 28, 29, 32, 5, 6, 9, 7, 8,
	}
	for i, gid := range gids {
		c, ok := class.Class(gid)
		tu.Assert(t, ok)
		tu.Assert(t, c == classes[i])
	}
	_, found = class.Class(0xFF)
	tu.Assert(t, !found)
}

func TestParseTrak(t *testing.T) {
	fp := readFontFile(t, "toys/Trak.ttf")
	trak, _, err := ParseTrak(readTable(t, fp, "trak"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(trak.Horiz.SizeTable) == 4)
	tu.Assert(t, len(trak.Vert.SizeTable) == 0)

	tu.Assert(t, reflect.DeepEqual(trak.Horiz.SizeTable, []float32{1, 2, 12, 96}))
	tu.Assert(t, reflect.DeepEqual(trak.Horiz.TrackTable[0].PerSizeTracking, []int16{200, 200, 0, -100}))
}

func TestParseFeat(t *testing.T) {
	fp := readFontFile(t, "toys/Feat.ttf")
	feat, _, err := ParseFeat(readTable(t, fp, "feat"))
	tu.AssertNoErr(t, err)

	expectedSettings := [...][]FeatureSettingName{
		{{2, 260}, {4, 259}, {10, 304}},
		{{0, 309}, {1, 263}, {3, 264}},
		{{0, 266}, {1, 267}},
		{{0, 271}, {2, 272}, {8, 273}},
		{{0, 309}, {1, 275}, {2, 277}, {3, 278}},
		{{0, 309}, {2, 280}},
		{{0, 283}},
		{{8, 308}},
		{{0, 309}, {3, 289}},
		{{0, 294}, {1, 295}, {2, 296}, {3, 297}},
		{{0, 309}, {1, 301}},
	}
	tu.Assert(t, len(feat.Names) == len(expectedSettings))
	for i, name := range feat.Names {
		exp := expectedSettings[i]
		got := name.SettingTable
		tu.Assert(t, reflect.DeepEqual(exp, got))
	}
}

func TestParseAnkr(t *testing.T) {
	table, err := td.Files.ReadFile("toys/tables/ankr.bin")
	tu.AssertNoErr(t, err)

	ankr, _, err := ParseAnkr(table, 1409)
	tu.AssertNoErr(t, err)

	_, isFormat4 := ankr.lookupTable.(AATLoopkup4)
	tu.Assert(t, isFormat4)
}

func TestParseMorx(t *testing.T) {
	files := tu.Filenames(t, "morx")
	files = append(files, "toys/Trak.ttf")
	for _, filename := range files {
		fp := readFontFile(t, filename)
		ng := numGlyphs(t, fp)

		table, _, err := ParseMorx(readTable(t, fp, "morx"), ng)
		tu.AssertNoErr(t, err)
		tu.Assert(t, int(table.nChains) == len(table.Chains))
		tu.Assert(t, int(table.nChains) == 1)

		for _, chain := range table.Chains {
			tu.AssertNoErr(t, err)
			tu.Assert(t, len(chain.Subtables) == int(chain.nSubtable))
			tu.Assert(t, chain.Flags == 1)
		}
	}
}

func TestMorxLigature(t *testing.T) {
	// imported from fonttools

	// Taken from “Example 2: A ligature table” in
	// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6morx.html
	// as retrieved on 2017-09-11.
	//
	// Compared to the example table in Apple’s specification, we’ve
	// made the following changes:
	//
	// * at offsets 0..35, we’ve prepended 36 bytes of boilerplate
	//   to make the data a structurally valid ‘morx’ table;
	//
	// * at offsets 88..91 (offsets 52..55 in Apple’s document), we’ve
	//   changed the range of the third segment from 23..24 to 26..28.
	//   The hexdump values in Apple’s specification are completely wrong;
	//   the values from the comments would work, but they can be encoded
	//   more compactly than in the specification example. For round-trip
	//   testing, we omit the ‘f’ glyph, which makes AAT lookup format 2
	//   the most compact encoding;
	//
	// * at offsets 92..93 (offsets 56..57 in Apple’s document), we’ve
	//   changed the glyph class of the third segment from 5 to 6, which
	//   matches the values from the comments to the spec (but not the
	//   Apple’s hexdump).
	morxLigatureData := deHexStr(
		"0002 0000 " + //  0: Version=2, Reserved=0
			"0000 0001 " + //  4: MorphChainCount=1
			"0000 0001 " + //  8: DefaultFlags=1
			"0000 00DA " + // 12: StructLength=218 (+8=226)
			"0000 0000 " + // 16: MorphFeatureCount=0
			"0000 0001 " + // 20: MorphSubtableCount=1
			"0000 00CA " + // 24: Subtable[0].StructLength=202 (+24=226)
			"80 " + // 28: Subtable[0].CoverageFlags=0x80
			"00 00 " + // 29: Subtable[0].Reserved=0
			"02 " + // 31: Subtable[0].MorphType=2/LigatureMorph
			"0000 0001 " + // 32: Subtable[0].SubFeatureFlags=0x1

			// State table header.
			"0000 0007 " + // 36: STXHeader.ClassCount=7
			"0000 001C " + // 40: STXHeader.ClassTableOffset=28 (+36=64)
			"0000 0040 " + // 44: STXHeader.StateArrayOffset=64 (+36=100)
			"0000 0078 " + // 48: STXHeader.EntryTableOffset=120 (+36=156)
			"0000 0090 " + // 52: STXHeader.LigActionsOffset=144 (+36=180)
			"0000 009C " + // 56: STXHeader.LigComponentsOffset=156 (+36=192)
			"0000 00AE " + // 60: STXHeader.LigListOffset=174 (+36=210)

			// Glyph class table.
			"0002 0006 " + // 64: ClassTable.LookupFormat=2, .UnitSize=6
			"0003 000C " + // 68:   .NUnits=3, .SearchRange=12
			"0001 0006 " + // 72:   .EntrySelector=1, .RangeShift=6
			"0016 0014 0004 " + // 76: GlyphID 20..22 [a..c] -> GlyphClass 4
			"0018 0017 0005 " + // 82: GlyphID 23..24 [d..e] -> GlyphClass 5
			"001C 001A 0006 " + // 88: GlyphID 26..28 [g..i] -> GlyphClass 6
			"FFFF FFFF 0000 " + // 94: <end of lookup>

			// State array.
			"0000 0000 0000 0000 0001 0000 0000 " + // 100: State[0][0..6]
			"0000 0000 0000 0000 0001 0000 0000 " + // 114: State[1][0..6]
			"0000 0000 0000 0000 0001 0002 0000 " + // 128: State[2][0..6]
			"0000 0000 0000 0000 0001 0002 0003 " + // 142: State[3][0..6]

			// Entry table.
			"0000 0000 " + // 156: Entries[0].NewState=0, .Flags=0
			"0000 " + // 160: Entries[0].ActionIndex=<n/a> because no 0x2000 flag
			"0002 8000 " + // 162: Entries[1].NewState=2, .Flags=0x8000 (SetComponent)
			"0000 " + // 166: Entries[1].ActionIndex=<n/a> because no 0x2000 flag
			"0003 8000 " + // 168: Entries[2].NewState=3, .Flags=0x8000 (SetComponent)
			"0000 " + // 172: Entries[2].ActionIndex=<n/a> because no 0x2000 flag
			"0000 A000 " + // 174: Entries[3].NewState=0, .Flags=0xA000 (SetComponent,Act)
			"0000 " + // 178: Entries[3].ActionIndex=0 (start at Action[0])

			// Ligature actions table.
			"3FFF FFE7 " + // 180: Action[0].Flags=0, .GlyphIndexDelta=-25
			"3FFF FFED " + // 184: Action[1].Flags=0, .GlyphIndexDelta=-19
			"BFFF FFF2 " + // 188: Action[2].Flags=<end of list>, .GlyphIndexDelta=-14

			// Ligature component table.
			"0000 0001 " + // 192: LigComponent[0]=0, LigComponent[1]=1
			"0002 0003 " + // 196: LigComponent[2]=2, LigComponent[3]=3
			"0000 0004 " + // 200: LigComponent[4]=0, LigComponent[5]=4
			"0000 0008 " + // 204: LigComponent[6]=0, LigComponent[7]=8
			"0010      " + // 208: LigComponent[8]=16

			// Ligature list.
			"03E8 03E9 " + // 210: LigList[0]=1000, LigList[1]=1001
			"03EA 03EB " + // 214: LigList[2]=1002, LigList[3]=1003
			"03EC 03ED " + // 218: LigList[4]=1004, LigList[3]=1005
			"03EE 03EF ") // 222: LigList[5]=1006, LigList[6]=1007

	tu.Assert(t, len(morxLigatureData) == 226)
	out, _, err := ParseMorx(morxLigatureData, 1515)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Chains) == 1)

	chain := out.Chains[0]
	tu.Assert(t, chain.Flags == 1)
	tu.Assert(t, len(chain.Subtables) == 1)
	subtable := chain.Subtables[0]

	const vertical uint8 = 0x80
	tu.Assert(t, subtable.Coverage == vertical)
	tu.Assert(t, subtable.SubFeatureFlags == 1)
	lig, ok := subtable.Data.(MorxSubtableLigature)
	tu.Assert(t, ok)
	machine := lig.AATStateTableExt
	tu.Assert(t, machine.StateSize == 7)

	class, ok := machine.Class.(AATLoopkup2)
	tu.Assert(t, ok)
	expMachineClassRecords := []LookupRecord2{
		{FirstGlyph: 20, LastGlyph: 22, Value: 4},
		{FirstGlyph: 23, LastGlyph: 24, Value: 5},
		{FirstGlyph: 26, LastGlyph: 28, Value: 6},
	}
	tu.Assert(t, reflect.DeepEqual(class.Records, expMachineClassRecords))

	expMachineStates := [][]uint16{
		{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000, 0x0000}, // State[0][0..6]
		{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000, 0x0000}, // State[1][0..6]
		{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0002, 0x0000}, // State[2][0..6]
		{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0002, 0x0003}, // State[3][0..6]
	}
	tu.Assert(t, reflect.DeepEqual(machine.States, expMachineStates))

	expMachineEntries := []AATStateEntry{
		{NewState: 0, Flags: 0},
		{NewState: 0x0002, Flags: 0x8000},
		{NewState: 0x0003, Flags: 0x8000},
		{NewState: 0, Flags: 0xA000},
	}
	tu.Assert(t, reflect.DeepEqual(machine.Entries, expMachineEntries))

	expLigActions := []uint32{
		0x3FFFFFE7,
		0x3FFFFFED,
		0xBFFFFFF2,
	}
	expComponents := []uint16{0, 1, 2, 3, 0, 4, 0, 8, 16}
	expLigatures := []GlyphID{
		1000, 1001, 1002, 1003, 1004, 1005, 1006, 1007,
	}
	tu.Assert(t, reflect.DeepEqual(lig.LigActions, expLigActions))
	tu.Assert(t, reflect.DeepEqual(lig.Components, expComponents))
	tu.Assert(t, reflect.DeepEqual(lig.Ligatures, expLigatures))
}

func TestMorxInsertion(t *testing.T) {
	// imported from fonttools

	// Taken from the `morx` table of the second font in DevanagariSangamMN.ttc
	// on macOS X 10.12.6; manually pruned to just contain the insertion lookup.
	morxInsertionData := deHexStr(
		"0002 0000 " + //  0: Version=2, Reserved=0
			"0000 0001 " + //  4: MorphChainCount=1
			"0000 0001 " + //  8: DefaultFlags=1
			"0000 00A4 " + // 12: StructLength=164 (+8=172)
			"0000 0000 " + // 16: MorphFeatureCount=0
			"0000 0001 " + // 20: MorphSubtableCount=1
			"0000 0094 " + // 24: Subtable[0].StructLength=148 (+24=172)
			"00 " + // 28: Subtable[0].CoverageFlags=0x00
			"00 00 " + // 29: Subtable[0].Reserved=0
			"05 " + // 31: Subtable[0].MorphType=5/InsertionMorph
			"0000 0001 " + // 32: Subtable[0].SubFeatureFlags=0x1
			"0000 0006 " + // 36: STXHeader.ClassCount=6
			"0000 0014 " + // 40: STXHeader.ClassTableOffset=20 (+36=56)
			"0000 004A " + // 44: STXHeader.StateArrayOffset=74 (+36=110)
			"0000 006E " + // 48: STXHeader.EntryTableOffset=110 (+36=146)
			"0000 0086 " + // 52: STXHeader.InsertionActionOffset=134 (+36=170)
			// Glyph class table.
			"0002 0006 " + //  56: ClassTable.LookupFormat=2, .UnitSize=6
			"0006 0018 " + //  60:   .NUnits=6, .SearchRange=24
			"0002 000C " + //  64:   .EntrySelector=2, .RangeShift=12
			"00AC 00AC 0005 " + //  68: GlyphID 172..172 -> GlyphClass 5
			"01EB 01E6 0005 " + //  74: GlyphID 486..491 -> GlyphClass 5
			"01F0 01F0 0004 " + //  80: GlyphID 496..496 -> GlyphClass 4
			"01F8 01F6 0004 " + //  88: GlyphID 502..504 -> GlyphClass 4
			"01FC 01FA 0004 " + //  92: GlyphID 506..508 -> GlyphClass 4
			"0250 0250 0005 " + //  98: GlyphID 592..592 -> GlyphClass 5
			"FFFF FFFF 0000 " + // 104: <end of lookup>
			// State array.
			"0000 0000 0000 0000 0001 0000 " + // 110: State[0][0..5]
			"0000 0000 0000 0000 0001 0000 " + // 122: State[1][0..5]
			"0000 0000 0001 0000 0001 0002 " + // 134: State[2][0..5]
			// Entry table.
			"0000 0000 " + // 146: Entries[0].NewState=0, .Flags=0
			"FFFF " + // 150: Entries[0].CurrentInsertIndex=<None>
			"FFFF " + // 152: Entries[0].MarkedInsertIndex=<None>
			"0002 0000 " + // 154: Entries[1].NewState=0, .Flags=0
			"FFFF " + // 158: Entries[1].CurrentInsertIndex=<None>
			"FFFF " + // 160: Entries[1].MarkedInsertIndex=<None>
			"0000 " + // 162: Entries[2].NewState=0
			"2820 " + // 164:   .Flags=CurrentIsKashidaLike,CurrentInsertBefore
			//        .CurrentInsertCount=1, .MarkedInsertCount=0
			"0000 " + // 166: Entries[1].CurrentInsertIndex=0
			"FFFF " + // 168: Entries[1].MarkedInsertIndex=<None>
			// Insertion action table.
			"022F") // 170: InsertionActionTable[0]=GlyphID 559

	tu.Assert(t, len(morxInsertionData) == 172)

	out, _, err := ParseMorx(morxInsertionData, 910)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(out.Chains) == 1)

	chain := out.Chains[0]

	tu.Assert(t, chain.Flags == 1)
	tu.Assert(t, len(chain.Subtables) == 1)
	subtable := chain.Subtables[0]

	const vertical uint8 = 0
	tu.Assert(t, subtable.Coverage == vertical)
	tu.Assert(t, subtable.SubFeatureFlags == 1)
	insert, ok := subtable.Data.(MorxSubtableInsertion)
	tu.Assert(t, ok)
	machine := insert.AATStateTableExt
	tu.Assert(t, machine.StateSize == 6)

	class, ok := machine.Class.(AATLoopkup2)
	tu.Assert(t, ok)
	expMachineClassRecords := []LookupRecord2{
		{FirstGlyph: 172, LastGlyph: 172, Value: 5},
		{FirstGlyph: 486, LastGlyph: 491, Value: 5},
		{FirstGlyph: 496, LastGlyph: 496, Value: 4},
		{FirstGlyph: 502, LastGlyph: 504, Value: 4},
		{FirstGlyph: 506, LastGlyph: 508, Value: 4},
		{FirstGlyph: 592, LastGlyph: 592, Value: 5},
	}
	tu.Assert(t, reflect.DeepEqual(class.Records, expMachineClassRecords))

	expMachineStates := [][]uint16{
		{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000}, // 110: State[0][0..5]
		{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000}, // 122: State[1][0..5]
		{0x0000, 0x0000, 0x0001, 0x0000, 0x0001, 0x0002}, // 134: State[2][0..5]
	}
	tu.Assert(t, reflect.DeepEqual(machine.States, expMachineStates))

	expMachineEntries := []AATStateEntry{
		{NewState: 0, Flags: 0, data: [4]byte{0xff, 0xff, 0xff, 0xff}},
		{NewState: 0x0002, Flags: 0, data: [4]byte{0xff, 0xff, 0xff, 0xff}},
		{NewState: 0, Flags: 0x2820, data: [4]byte{0, 0, 0xff, 0xff}},
	}
	tu.Assert(t, reflect.DeepEqual(machine.Entries, expMachineEntries))

	tu.Assert(t, reflect.DeepEqual(insert.Insertions, []GlyphID{0x022f}))
}

func TestParseKerx(t *testing.T) {
	for _, filepath := range []string{
		"toys/tables/kerx0.bin",
		"toys/tables/kerx2.bin",
		"toys/tables/kerx2bis.bin",
		"toys/tables/kerx24.bin",
		"toys/tables/kerx4-1.bin",
		"toys/tables/kerx4-2.bin",
		"toys/tables/kerx6Exp-VF.bin",
		"toys/tables/kerx6-VF.bin",
	} {
		table, err := td.Files.ReadFile(filepath)
		tu.AssertNoErr(t, err)

		kerx, _, err := ParseKerx(table, 0xFF)
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(kerx.Tables) > 0)

		for _, subtable := range kerx.Tables {
			tu.Assert(t, subtable.TupleCount > 0 == strings.Contains(filepath, "VF"))
			switch data := subtable.Data.(type) {
			case KerxData0:
				tu.Assert(t, len(data.Pairs) > 0)
			case KerxData2:
				tu.Assert(t, data.Left != nil)
				tu.Assert(t, data.Right != nil)
				tu.Assert(t, int(data.KerningStart) <= len(data.KerningData))
			case KerxData4:
				tu.Assert(t, data.Anchors != nil)
			}
		}
	}
}

func TestInvalidFeat(t *testing.T) {
	// this is an invalid feat table, comming from a real font table (huh...)
	table, err := td.Files.ReadFile("toys/tables/featInvalid.bin")
	tu.AssertNoErr(t, err)

	_, _, err = ParseFeat(table)
	tu.Assert(t, err != nil)
}

func TestParseLtag(t *testing.T) {
	table, err := td.Files.ReadFile("toys/tables/ltag.bin")
	tu.AssertNoErr(t, err)

	ltag, _, err := ParseLtag(table)
	tu.AssertNoErr(t, err)

	tu.Assert(t, len(ltag.tagRange) == 1)
	tu.Assert(t, ltag.Language(0) == "pl")
}
