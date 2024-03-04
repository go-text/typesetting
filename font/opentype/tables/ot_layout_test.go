// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"fmt"
	"reflect"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

func TestParseOTLayout(t *testing.T) {
	for _, filename := range td.WithOTLayout {
		fp := readFontFile(t, filename)
		gsub, _, err := ParseLayout(readTable(t, fp, "GSUB"))
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(gsub.LookupList.Lookups) > 0)

		for _, lookup := range gsub.LookupList.Lookups {
			tu.Assert(t, lookup.lookupType > 0)
			lks, err := lookup.AsGSUBLookups()
			tu.AssertNoErr(t, err)
			for _, subtable := range lks {
				_, isExtension := subtable.(ExtensionSubs)
				tu.Assert(t, isExtension || subtable.Cov() != nil)
			}
		}

		gpos, _, err := ParseLayout(readTable(t, fp, "GPOS"))
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(gpos.LookupList.Lookups) > 0)

		for _, lookup := range gpos.LookupList.Lookups {
			tu.Assert(t, lookup.lookupType > 0)
			lks, err := lookup.AsGPOSLookups()
			tu.AssertNoErr(t, err)
			for _, subtable := range lks {
				_, isExtension := subtable.(ExtensionPos)
				tu.Assert(t, isExtension || subtable.Cov() != nil)
			}
		}

		_, _, err = ParseGDEF(readTable(t, fp, "GDEF"))
		tu.AssertNoErr(t, err)
	}
}

func TestGSUB(t *testing.T) {
	for _, filename := range tu.Filenames(t, "toys/gsub") {
		fp := readFontFile(t, filename)
		gsub, _, err := ParseLayout(readTable(t, fp, "GSUB"))
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(gsub.LookupList.Lookups) > 0)

		for _, lookup := range gsub.LookupList.Lookups {
			tu.Assert(t, lookup.lookupType > 0)
			lks, err := lookup.AsGSUBLookups()
			tu.AssertNoErr(t, err)
			for _, subtable := range lks {
				_, isExtension := subtable.(ExtensionSubs)
				tu.Assert(t, isExtension || subtable.Cov() != nil)
			}
		}
	}
}

func TestGPOS(t *testing.T) {
	for _, filename := range tu.Filenames(t, "toys/gpos") {
		fp := readFontFile(t, filename)
		gpos, _, err := ParseLayout(readTable(t, fp, "GPOS"))
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(gpos.LookupList.Lookups) > 0)

		for _, lookup := range gpos.LookupList.Lookups {
			tu.Assert(t, lookup.lookupType > 0)
			lks, err := lookup.AsGPOSLookups()
			tu.AssertNoErr(t, err)
			for _, subtable := range lks {
				_, isExtension := subtable.(ExtensionPos)
				tu.Assert(t, isExtension || subtable.Cov() != nil)
			}
		}
	}
}

func TestGSUBIndic(t *testing.T) {
	filepath := "toys/gsub/GSUBChainedContext2.ttf"
	fp := readFontFile(t, filepath)
	gsub, _, err := ParseLayout(readTable(t, fp, "GSUB"))
	tu.AssertNoErr(t, err)

	expectedScripts := []Script{
		{
			// Tag: MustNewTag("beng"),
			DefaultLangSys: &LangSys{
				RequiredFeatureIndex: 0xFFFF,
				FeatureIndices:       []uint16{0, 2},
			},
			LangSysRecords: []TagOffsetRecord{},
			LangSys:        []LangSys{},
		},
		{
			// Tag: MustNewTag("bng2"),
			DefaultLangSys: &LangSys{
				RequiredFeatureIndex: 0xFFFF,
				FeatureIndices:       []uint16{1, 2},
			},
			LangSysRecords: []TagOffsetRecord{},
			LangSys:        []LangSys{},
		},
	}

	expectedFeatures := []Feature{
		{
			// Tag: MustNewTag("init"),
			LookupListIndices: []uint16{0},
		},
		{
			// Tag: MustNewTag("init"),
			LookupListIndices: []uint16{1},
		},
		{
			// Tag: MustNewTag("blws"),
			LookupListIndices: []uint16{2},
		},
	}

	expectedLoopkup1 := []GSUBLookup{
		SingleSubs{
			Data: SingleSubstData1{
				format:       1,
				Coverage:     Coverage1{1, []GlyphID{6, 7}},
				DeltaGlyphID: 3,
			},
		},
	}

	expectedLoopkup2 := []GSUBLookup{
		SingleSubs{
			Data: SingleSubstData1{
				format:       1,
				Coverage:     Coverage1{1, []GlyphID{6, 7}},
				DeltaGlyphID: 3,
			},
		},
	}

	expectedLoopkup3 := []GSUBLookup{
		ChainedContextualSubs{
			Data: ChainedContextualSubs2{
				format:   2,
				coverage: Coverage1{1, []GlyphID{5}},
				BacktrackClassDef: ClassDef1{
					format:          1,
					StartGlyphID:    2,
					ClassValueArray: []uint16{1},
				},
				InputClassDef: ClassDef1{
					format:          1,
					StartGlyphID:    5,
					ClassValueArray: []uint16{1},
				},
				LookaheadClassDef: ClassDef2{
					format:            2,
					ClassRangeRecords: []ClassRangeRecord{},
				},
				ChainedClassSeqRuleSet: []ChainedClassSequenceRuleSet{
					{},
					{
						[]ChainedSequenceRule{
							{
								BacktrackSequence: []uint16{1},
								inputGlyphCount:   1,
								InputSequence:     []uint16{},
								LookaheadSequence: []uint16{},
								SeqLookupRecords: []SequenceLookupRecord{
									{SequenceIndex: 0, LookupListIndex: 3},
								},
							},
						},
					},
				},
			},
		},
	}

	expectedLoopkup4 := []GSUBLookup{
		SingleSubs{
			Data: SingleSubstData1{
				format:       1,
				Coverage:     Coverage1{1, []GlyphID{5}},
				DeltaGlyphID: 6,
			},
		},
	}
	expectedLookups := [][]GSUBLookup{
		expectedLoopkup1, expectedLoopkup2, expectedLoopkup3, expectedLoopkup4,
	}

	if exp, got := expectedScripts, gsub.ScriptList.Scripts; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	if exp, got := expectedFeatures, gsub.FeatureList.Features; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	for i, lk := range gsub.LookupList.Lookups {
		got, err := lk.AsGSUBLookups()
		tu.AssertNoErr(t, err)
		exp := expectedLookups[i]
		tu.AssertC(t, reflect.DeepEqual(got, exp), fmt.Sprintf("lookup %d expected \n%v\n, got \n%v\n", i, exp, got))
	}
}

func TestGSUBLigature(t *testing.T) {
	filepath := "toys/gsub/GSUBLigature.ttf"
	fp := readFontFile(t, filepath)
	gsub, _, err := ParseLayout(readTable(t, fp, "GSUB"))
	tu.AssertNoErr(t, err)

	lookups, err := gsub.LookupList.Lookups[0].AsGSUBLookups()
	tu.AssertNoErr(t, err)
	lookup := lookups[0]

	expected := LigatureSubs{
		substFormat: 1,
		Coverage:    Coverage1{1, []GlyphID{3, 4, 7, 8, 9}},
		LigatureSets: []LigatureSet{
			{ // glyph="3"
				[]Ligature{
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{5}, LigatureGlyph: 6},
				},
			},
			{ // glyph="4"
				[]Ligature{
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{4}, LigatureGlyph: 31},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{7}, LigatureGlyph: 32},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{11}, LigatureGlyph: 34},
				},
			},
			{ // glyph="7"
				[]Ligature{
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{4}, LigatureGlyph: 37},
					{componentCount: 3, ComponentGlyphIDs: []GlyphID{7, 7}, LigatureGlyph: 40},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{7}, LigatureGlyph: 8},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{8}, LigatureGlyph: 40},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{11}, LigatureGlyph: 38},
				},
			},
			{ // glyph="8"
				[]Ligature{
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{7}, LigatureGlyph: 40},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{11}, LigatureGlyph: 42},
				},
			},
			{ // glyph="9"
				[]Ligature{
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{4}, LigatureGlyph: 44},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{7}, LigatureGlyph: 45},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{9}, LigatureGlyph: 10},
					{componentCount: 2, ComponentGlyphIDs: []GlyphID{11}, LigatureGlyph: 46},
				},
			},
		},
	}

	tu.AssertC(t, reflect.DeepEqual(lookup, expected), fmt.Sprintf("expected %v, got %v", expected, lookup))
}

func TestGPOSCursive(t *testing.T) {
	filepath := "toys/gpos/GPOSCursive.ttf"
	fp := readFontFile(t, filepath)
	gpos, _, err := ParseLayout(readTable(t, fp, "GPOS"))
	tu.AssertNoErr(t, err)

	if len(gpos.LookupList.Lookups) != 4 || len(gpos.LookupList.Lookups[0].subtableOffsets) != 1 {
		t.Fatalf("invalid gpos lookups: %v", gpos.LookupList)
	}

	lookups, err := gpos.LookupList.Lookups[0].AsGPOSLookups()
	tu.AssertNoErr(t, err)

	cursive, ok := lookups[0].(CursivePos)
	tu.AssertC(t, ok, fmt.Sprintf("unexpected type for lookup %T", lookups[0]))

	got := cursive.EntryExits
	expected := []EntryExit{
		{AnchorFormat1{1, 405, 45}, AnchorFormat1{1, 0, 0}},
		{AnchorFormat1{1, 452, 500}, AnchorFormat1{1, 0, 0}},
	}
	tu.AssertC(t, reflect.DeepEqual(expected, got), fmt.Sprintf("expected %v, got %v", expected, got))
}

func TestBits(t *testing.T) {
	var buf8 [8]int8
	uint16As2Bits(buf8[:], 0x123F)
	if exp := [8]int8{0, 1, 0, -2, 0, -1, -1, -1}; buf8 != exp {
		t.Fatalf("expected %v, got %v", exp, buf8)
	}

	var buf4 [4]int8
	uint16As4Bits(buf4[:], 0x123F)
	if exp := [4]int8{1, 2, 3, -1}; buf4 != exp {
		t.Fatalf("expected %v, got %v", exp, buf4)
	}

	var buf2 [2]int8
	uint16As8Bits(buf2[:], 0x123F)
	if exp := [2]int8{18, 63}; buf2 != exp {
		t.Fatalf("expected %v, got %v", exp, buf2)
	}
}

func TestGDEFCaretList3(t *testing.T) {
	filepath := "toys/GDEFCaretList3.ttf"
	fp := readFontFile(t, filepath)
	gdef, _, err := ParseGDEF(readTable(t, fp, "GDEF"))
	tu.AssertNoErr(t, err)

	expectedLigGlyphs := [][]CaretValue3{ //  LigGlyphCount=4
		// CaretCount=1
		{
			CaretValue3{Coordinate: 620, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 205}},
		},
		// CaretCount=1
		{
			CaretValue3{Coordinate: 675, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 193}},
		},
		// CaretCount=2
		{
			CaretValue3{Coordinate: 696, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 173}},
			CaretValue3{Coordinate: 1351, Device: DeviceVariation{DeltaSetOuter: 6, DeltaSetInner: 14}},
		},
		// CaretCount=2
		{
			CaretValue3{Coordinate: 702, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 179}},
			CaretValue3{Coordinate: 1392, Device: DeviceVariation{DeltaSetOuter: 6, DeltaSetInner: 11}},
		},
	}

	for i := range expectedLigGlyphs {
		expL, gotL := expectedLigGlyphs[i], gdef.LigCaretList.LigGlyphs[i]
		tu.Assert(t, len(expL) == len(gotL.CaretValues))
		for j := range expL {
			exp, got := expL[j], gotL.CaretValues[j]
			asFormat3, ok := got.(CaretValue3)
			tu.Assert(t, ok)
			tu.Assert(t, exp.Coordinate == asFormat3.Coordinate)
			tu.Assert(t, exp.Device == asFormat3.Device)
		}
	}
}

func TestGDEFVarStore(t *testing.T) {
	filepath := "common/Commissioner-VF.ttf"
	fp := readFontFile(t, filepath)
	gdef, _, err := ParseGDEF(readTable(t, fp, "GDEF"))
	tu.AssertNoErr(t, err)

	tu.Assert(t, len(gdef.ItemVarStore.VariationRegionList.VariationRegions) == 15)
	tu.Assert(t, len(gdef.ItemVarStore.ItemVariationDatas) == 52)
}

func TestGPOS2_1(t *testing.T) {
	fp := readFontFile(t, "toys/gpos/gpos2_1_font6.otf")

	gposT, _, err := ParseLayout(readTable(t, fp, "GPOS"))
	tu.AssertNoErr(t, err)

	tu.Assert(t, len(gposT.LookupList.Lookups) == 1)
	subtables, err := gposT.LookupList.Lookups[0].AsGPOSLookups()
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(subtables) == 1)
	pairpos, ok := subtables[0].(PairPos)
	tu.Assert(t, ok)
	data, ok := pairpos.Data.(PairPosData1)
	tu.Assert(t, ok)
	tu.Assert(t, data.ValueFormat1 == 1 && data.ValueFormat2 == 2)
	tu.Assert(t, len(data.PairSets) == 1)
	tu.Assert(t, data.PairSets[0].pairValueCount == 2)

	v1 := PairValueRecord{
		19,
		ValueRecord{-200, 0, 0, 0, nil, nil, nil, nil},
		ValueRecord{0, -100, 0, 0, nil, nil, nil, nil},
	}
	v2 := PairValueRecord{
		20,
		ValueRecord{-300, 0, 0, 0, nil, nil, nil, nil},
		ValueRecord{0, -400, 0, 0, nil, nil, nil, nil},
	}
	v1g, err := data.PairSets[0].data.get(0)
	tu.AssertNoErr(t, err)
	v2g, err := data.PairSets[0].data.get(1)
	tu.AssertNoErr(t, err)

	tu.Assert(t, reflect.DeepEqual(v1, v1g))
	tu.Assert(t, reflect.DeepEqual(v2, v2g))
}
