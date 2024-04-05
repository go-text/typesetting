// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"
)

// ported from harfbuzz/test/api/test-var-coords.c Copyright © 2019 Ebrahim Byagowi

func TestVar(t *testing.T) {
	font := loadFont(t, "toys/CFF2-VF.otf")

	for _, test := range []struct {
		design     float32
		normalized VarCoord
	}{
		{200, -16384},
		{206, -16117},
		{400, 0},
		{450, 1529},
		{600, 6014},
	} {
		coords := font.NormalizeVariations([]float32{test.design})
		tu.AssertC(t, coords[0] == test.normalized, fmt.Sprintf("%d != %d", coords[0], test.normalized))
	}

	// test for crash
	for weight := float32(200); weight < 901; weight++ {
		font.NormalizeVariations([]float32{weight})
	}

	face := Face{Font: font}
	face.SetVariations([]Variation{{ot.MustNewTag("wght"), 206.}})
	tu.Assert(t, len(face.coords) == 1)
	tu.Assert(t, face.coords[0] == -16117)

	face.SetVariations(nil)
	tu.Assert(t, len(face.coords) == 0)

	font = loadFont(t, "common/NotoSansCJKjp-VF.otf")
	for _, test := range []struct {
		design     float32
		normalized VarCoord
	}{
		{200, 1311},
		{301, 2672},
		{400, 6390},
		{401.2, 6424},
	} {
		coords := font.NormalizeVariations([]float32{test.design})
		tu.AssertC(t, coords[0] == test.normalized, fmt.Sprintf("%d != %d", coords[0], test.normalized))
	}
}

func TestGlyphExtentsVar(t *testing.T) {
	font := loadFont(t, "common/SourceSans-VF-HVAR.ttf")

	for _, test := range []struct {
		coord    float32
		expected GlyphExtents
	}{
		{100, GlyphExtents{XBearing: 56, YBearing: 672, Width: 556, Height: -684}},
		{500, GlyphExtents{XBearing: 50, YBearing: 667, Width: 592, Height: -679}},
		{900, GlyphExtents{XBearing: 44, YBearing: 662, Width: 630, Height: -674}},
	} {
		coords := font.NormalizeVariations([]float32{test.coord})
		face := NewFace(font)
		face.coords = coords

		ext, _ := face.GlyphExtents(2)

		// the ref. values are extracted from harfbuzz, which round extents
		ext.XBearing = float32(math.Round(float64(ext.XBearing)))
		ext.YBearing = float32(math.Round(float64(ext.YBearing)))
		ext.Width = float32(math.Round(float64(ext.Width)))
		ext.Height = float32(math.Round(float64(ext.Height)))

		tu.Assert(t, ext == test.expected)
	}
}

func TestGetDefaultCoords(t *testing.T) {
	tf := fvar{
		{Tag: ot.MustNewTag("wght"), Minimum: 38, Default: 88, Maximum: 250},
		{Tag: ot.MustNewTag("wdth"), Minimum: 60, Default: 402, Maximum: 402},
		{Tag: ot.MustNewTag("opsz"), Minimum: 10, Default: 14, Maximum: 72},
	}

	vars := []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 60},
	}
	coords := tf.getDesignCoordsDefault(vars)
	tu.Assert(t, reflect.DeepEqual(coords, []float32{88, 60, 14}))
}

func TestNormalizeVar(t *testing.T) {
	tf := fvar{
		{Tag: ot.MustNewTag("wdth"), Minimum: 60, Default: 402, Maximum: 500},
	}

	vars := []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 60},
	}
	coords := tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []VarCoord{tables.NewCoord(-1)}))

	vars = []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 30},
	}
	coords = tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []VarCoord{tables.NewCoord(-1)}))

	vars = []Variation{
		{Tag: ot.MustNewTag("wdth"), Value: 700},
	}
	coords = tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []VarCoord{tables.NewCoord(1)}))
}

func TestAdvanceHVar(t *testing.T) {
	font := loadFont(t, "common/Commissioner-VF.ttf")
	coords := []VarCoord{-6553, 0, 13108, tables.NewCoord(1)}

	// 0 - 99 GIDs
	exps := [100]float32{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6,
		1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 1.6, 3.2, 3.2, 3.2, 0, 0,
	}
	tu.Assert(t, font.hvar != nil)
	for i, exp := range exps {
		got := getAdvanceDeltaUnscaled(font.hvar, tables.GlyphID(i), coords)
		if math.Abs(float64(got-exp)) > 0.001 {
			t.Errorf("expected %f, got %f", exp, got)
		}
	}
}

func TestAdvanceNoHVar(t *testing.T) {
	font := loadFont(t, "toys/GVAR-no-HVAR.ttf")

	tu.Assert(t, len(font.fvar) == 2)

	vars := []Variation{
		{Tag: ot.MustNewTag("wght"), Value: 600},
		{Tag: ot.MustNewTag("wght"), Value: 80},
	}
	face := Face{Font: font}
	face.SetVariations(vars)

	// 0 - 14 GIDs
	exps := [15]float32{600, 1164, 1170, 813, 741, 1164, 1170, 813, 741, 270, 270, 0, 0, 0, 0}

	for i, exp := range exps {
		got := face.HorizontalAdvance(GID(i))
		tu.Assert(t, exp == got)
	}
}

func TestInvalidGVAR(t *testing.T) {
	// this file is build by subsetting the 'glyf' table
	// but keeping the variations tables
	f, err := os.Open("testdata/Selawik-VF-Subset.ttf")
	tu.AssertNoErr(t, err)
	defer f.Close()

	ld, err := ot.NewLoader(f)
	tu.AssertNoErr(t, err)

	head, _, _ := loadHeadTable(ld, nil)
	raw, _ := ld.RawTable(ot.MustNewTag("maxp"))
	maxp, _, _ := tables.ParseMaxp(raw)
	raw, _ = ld.RawTable(ot.MustNewTag("glyf"))
	locaRaw, _ := ld.RawTable(ot.MustNewTag("loca"))
	loca, _ := tables.ParseLoca(locaRaw, int(maxp.NumGlyphs), head.IndexToLocFormat == 1)
	glyf, _ := tables.ParseGlyf(raw, loca)

	raw, err = ld.RawTable(ot.MustNewTag("gvar"))
	tu.AssertNoErr(t, err)
	gvar, _, err := tables.ParseGvar(raw)
	tu.AssertNoErr(t, err)
	// check that newGvar does not crash..
	_, err = newGvar(gvar, glyf)
	// ... and reports an error
	tu.Assert(t, err != nil)
}

func TestCFF2Var(t *testing.T) {
	type extent struct {
		GID     tables.GlyphID
		Extents [4]int
	}
	type data struct {
		Coord   float32
		Extents []extent
	}

	// these datas are generated using the C++ Harfbuzz implementation
	datas := []data{
		{
			Coord: 100,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{66, 540, 380, -553}},
				{70, [4]int{59, 540, 426, -553}},
				{77, [4]int{105, 794, 84, -807}},
				{78, [4]int{105, 540, 687, -540}},
				{81, [4]int{105, 540, 431, -783}},
				{84, [4]int{37, 540, 364, -553}},
				{85, [4]int{33, 681, 290, -694}},
				{89, [4]int{15, 527, 404, -527}},
				{1397, [4]int{64, 206, 256, -249}},
				{1398, [4]int{52, 240, 286, -286}},
				{1461, [4]int{127, 780, 759, -816}},
				{1463, [4]int{147, 683, 751, -679}},
				{1469, [4]int{129, 785, 803, -797}},
				{1472, [4]int{173, 779, 676, -813}},
				{1480, [4]int{173, 778, 678, -804}},
				{1484, [4]int{104, 778, 818, -811}},
				{1488, [4]int{114, 747, 755, -783}},
				{1490, [4]int{110, 779, 789, -799}},
				{1494, [4]int{178, 483, 652, -488}},
				{1495, [4]int{92, 626, 796, -594}},
				{1498, [4]int{88, 697, 831, -725}},
				{1499, [4]int{224, 759, 608, -771}},
				{1502, [4]int{146, 742, 742, -766}},
				{1505, [4]int{112, 693, 777, -708}},
				{1525, [4]int{103, 774, 746, -805}},
				{1532, [4]int{200, 763, 630, -808}},
				{1541, [4]int{118, 782, 744, -813}},
				{1557, [4]int{110, 756, 728, -773}},
				{1562, [4]int{200, 601, 650, -644}},
				{1593, [4]int{355, 757, 471, -774}},
				{1600, [4]int{101, 675, 796, -650}},
				{1637, [4]int{140, 829, 785, -856}},
				{1645, [4]int{112, 403, 775, -38}},
				{9496, [4]int{55, 757, 890, -826}},
				{10927, [4]int{47, 832, 905, -905}},
				{11016, [4]int{44, 826, 922, -896}},
				{11864, [4]int{47, 832, 907, -904}},
				{11922, [4]int{55, 828, 893, -903}},
				{14025, [4]int{47, 779, 913, -851}},
				{15364, [4]int{78, 834, 847, -903}},
				{15585, [4]int{47, 831, 905, -902}},
				{16915, [4]int{42, 830, 903, -906}},
				{17513, [4]int{40, 831, 908, -895}},
				{20106, [4]int{47, 820, 904, -890}},
				{20282, [4]int{92, 785, 823, -861}},
				{20333, [4]int{48, 782, 905, -853}},
				{20698, [4]int{43, 765, 844, -837}},
				{20791, [4]int{56, 831, 891, -899}},
				{26323, [4]int{59, 757, 883, -781}},
				{27699, [4]int{100, 839, 813, -887}},
				{40641, [4]int{60, 780, 868, -849}},
				{43650, [4]int{63, 831, 874, -904}},
				{44199, [4]int{47, 763, 904, -834}},
				{59058, [4]int{108, 228, 161, -304}},
				{63163, [4]int{118, 75, 826, -88}},
			},
		},
		{
			Coord: 200,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{64, 543, 390, -556}},
				{70, [4]int{57, 543, 434, -556}},
				{77, [4]int{102, 794, 98, -807}},
				{78, [4]int{102, 543, 700, -543}},
				{81, [4]int{102, 543, 440, -783}},
				{84, [4]int{36, 543, 371, -556}},
				{85, [4]int{32, 684, 299, -697}},
				{89, [4]int{15, 530, 417, -530}},
				{1397, [4]int{62, 210, 262, -256}},
				{1398, [4]int{50, 241, 290, -290}},
				{1461, [4]int{123, 783, 767, -822}},
				{1463, [4]int{143, 686, 760, -685}},
				{1469, [4]int{124, 788, 812, -804}},
				{1472, [4]int{170, 783, 681, -820}},
				{1480, [4]int{170, 782, 683, -811}},
				{1484, [4]int{102, 781, 821, -818}},
				{1488, [4]int{111, 751, 760, -791}},
				{1490, [4]int{108, 782, 793, -805}},
				{1494, [4]int{174, 488, 660, -497}},
				{1495, [4]int{88, 631, 804, -603}},
				{1498, [4]int{86, 702, 835, -732}},
				{1499, [4]int{219, 763, 615, -778}},
				{1502, [4]int{142, 746, 748, -772}},
				{1505, [4]int{107, 698, 786, -716}},
				{1525, [4]int{101, 778, 752, -812}},
				{1532, [4]int{195, 767, 640, -816}},
				{1541, [4]int{113, 785, 753, -819}},
				{1557, [4]int{105, 762, 738, -782}},
				{1562, [4]int{195, 603, 657, -650}},
				{1593, [4]int{350, 761, 480, -781}},
				{1600, [4]int{96, 680, 807, -658}},
				{1637, [4]int{135, 833, 794, -864}},
				{1645, [4]int{110, 409, 779, -50}},
				{9496, [4]int{53, 760, 895, -831}},
				{10927, [4]int{45, 834, 909, -909}},
				{11016, [4]int{42, 828, 926, -900}},
				{11864, [4]int{45, 834, 910, -908}},
				{11922, [4]int{52, 830, 899, -906}},
				{14025, [4]int{45, 782, 917, -856}},
				{15364, [4]int{77, 836, 849, -907}},
				{15585, [4]int{45, 833, 909, -906}},
				{16915, [4]int{40, 832, 907, -910}},
				{17513, [4]int{38, 833, 912, -898}},
				{20106, [4]int{45, 822, 908, -894}},
				{20282, [4]int{90, 787, 828, -865}},
				{20333, [4]int{45, 785, 910, -858}},
				{20698, [4]int{40, 769, 851, -843}},
				{20791, [4]int{54, 833, 894, -903}},
				{26323, [4]int{58, 760, 886, -786}},
				{27699, [4]int{97, 840, 819, -889}},
				{40641, [4]int{56, 782, 876, -853}},
				{43650, [4]int{62, 833, 876, -908}},
				{44199, [4]int{45, 768, 910, -841}},
				{59058, [4]int{114, 229, 168, -311}},
				{63163, [4]int{115, 86, 838, -99}},
			},
		},
		{
			Coord: 301,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{63, 547, 398, -560}},
				{70, [4]int{56, 547, 440, -560}},
				{77, [4]int{100, 795, 111, -808}},
				{78, [4]int{100, 547, 713, -547}},
				{81, [4]int{100, 547, 449, -784}},
				{84, [4]int{35, 547, 379, -560}},
				{85, [4]int{31, 687, 309, -700}},
				{89, [4]int{15, 534, 431, -534}},
				{1397, [4]int{59, 214, 270, -262}},
				{1398, [4]int{48, 242, 294, -294}},
				{1461, [4]int{119, 786, 775, -828}},
				{1463, [4]int{138, 690, 769, -691}},
				{1469, [4]int{119, 791, 820, -811}},
				{1472, [4]int{167, 787, 686, -827}},
				{1480, [4]int{167, 786, 689, -819}},
				{1484, [4]int{101, 784, 824, -826}},
				{1488, [4]int{109, 756, 763, -799}},
				{1490, [4]int{106, 786, 797, -813}},
				{1494, [4]int{171, 492, 668, -505}},
				{1495, [4]int{84, 636, 812, -612}},
				{1498, [4]int{84, 707, 839, -739}},
				{1499, [4]int{214, 767, 622, -786}},
				{1502, [4]int{137, 750, 755, -778}},
				{1505, [4]int{102, 703, 795, -725}},
				{1525, [4]int{99, 781, 759, -818}},
				{1532, [4]int{191, 772, 648, -826}},
				{1541, [4]int{108, 789, 762, -826}},
				{1557, [4]int{100, 767, 749, -790}},
				{1562, [4]int{189, 605, 665, -656}},
				{1593, [4]int{344, 764, 491, -787}},
				{1600, [4]int{90, 685, 818, -667}},
				{1637, [4]int{129, 837, 805, -873}},
				{1645, [4]int{108, 416, 783, -63}},
				{9496, [4]int{50, 762, 901, -835}},
				{10927, [4]int{44, 836, 911, -912}},
				{11016, [4]int{40, 830, 930, -904}},
				{11864, [4]int{42, 836, 914, -911}},
				{11922, [4]int{50, 831, 905, -907}},
				{14025, [4]int{43, 786, 921, -862}},
				{15364, [4]int{75, 837, 853, -910}},
				{15585, [4]int{43, 835, 913, -910}},
				{16915, [4]int{38, 834, 911, -913}},
				{17513, [4]int{36, 835, 917, -901}},
				{20106, [4]int{43, 824, 912, -898}},
				{20282, [4]int{87, 790, 834, -870}},
				{20333, [4]int{43, 789, 914, -864}},
				{20698, [4]int{38, 773, 857, -850}},
				{20791, [4]int{52, 835, 897, -907}},
				{26323, [4]int{56, 764, 889, -793}},
				{27699, [4]int{94, 841, 826, -894}},
				{40641, [4]int{52, 784, 885, -857}},
				{43650, [4]int{62, 835, 876, -912}},
				{44199, [4]int{42, 773, 916, -849}},
				{59058, [4]int{119, 231, 176, -320}},
				{63163, [4]int{113, 97, 849, -110}},
			},
		},
		{
			Coord: 400,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{59, 557, 424, -570}},
				{70, [4]int{52, 557, 460, -570}},
				{77, [4]int{92, 796, 149, -809}},
				{78, [4]int{92, 557, 749, -557}},
				{81, [4]int{92, 557, 475, -786}},
				{84, [4]int{32, 557, 399, -570}},
				{85, [4]int{27, 696, 336, -709}},
				{89, [4]int{15, 543, 468, -543}},
				{1397, [4]int{52, 224, 289, -280}},
				{1398, [4]int{42, 244, 305, -305}},
				{1461, [4]int{109, 794, 796, -844}},
				{1463, [4]int{126, 700, 794, -709}},
				{1469, [4]int{105, 799, 845, -831}},
				{1472, [4]int{160, 798, 699, -846}},
				{1480, [4]int{158, 798, 704, -840}},
				{1484, [4]int{96, 792, 833, -846}},
				{1488, [4]int{102, 768, 775, -822}},
				{1490, [4]int{101, 795, 807, -831}},
				{1494, [4]int{160, 506, 691, -529}},
				{1495, [4]int{73, 650, 835, -636}},
				{1498, [4]int{79, 722, 849, -759}},
				{1499, [4]int{201, 778, 640, -806}},
				{1502, [4]int{124, 760, 775, -793}},
				{1505, [4]int{88, 718, 820, -749}},
				{1525, [4]int{94, 792, 776, -838}},
				{1532, [4]int{178, 784, 674, -850}},
				{1541, [4]int{95, 798, 787, -843}},
				{1557, [4]int{86, 783, 777, -814}},
				{1562, [4]int{174, 612, 686, -674}},
				{1593, [4]int{329, 774, 518, -804}},
				{1600, [4]int{75, 699, 849, -690}},
				{1637, [4]int{114, 847, 832, -895}},
				{1645, [4]int{102, 433, 795, -98}},
				{9496, [4]int{44, 770, 916, -848}},
				{10927, [4]int{39, 843, 921, -924}},
				{11016, [4]int{35, 835, 940, -914}},
				{11864, [4]int{35, 841, 924, -921}},
				{11922, [4]int{43, 836, 921, -916}},
				{14025, [4]int{37, 795, 932, -876}},
				{15364, [4]int{71, 842, 861, -920}},
				{15585, [4]int{38, 840, 923, -920}},
				{16915, [4]int{31, 840, 924, -923}},
				{17513, [4]int{31, 840, 928, -909}},
				{20106, [4]int{38, 829, 922, -909}},
				{20282, [4]int{80, 797, 848, -884}},
				{20333, [4]int{35, 799, 928, -880}},
				{20698, [4]int{30, 785, 875, -868}},
				{20791, [4]int{47, 840, 905, -918}},
				{26323, [4]int{52, 773, 897, -808}},
				{27699, [4]int{87, 844, 842, -908}},
				{40641, [4]int{42, 790, 907, -869}},
				{43650, [4]int{59, 840, 882, -922}},
				{44199, [4]int{35, 786, 934, -868}},
				{59058, [4]int{135, 235, 195, -342}},
				{63163, [4]int{106, 126, 881, -139}},
			},
		},
		{
			Coord: 450,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{57, 560, 434, -573}},
				{70, [4]int{50, 560, 468, -573}},
				{77, [4]int{89, 796, 163, -809}},
				{78, [4]int{89, 560, 763, -560}},
				{81, [4]int{89, 560, 485, -786}},
				{84, [4]int{30, 560, 408, -573}},
				{85, [4]int{26, 700, 345, -713}},
				{89, [4]int{15, 547, 481, -547}},
				{1397, [4]int{49, 228, 297, -287}},
				{1398, [4]int{40, 245, 308, -309}},
				{1461, [4]int{105, 797, 804, -850}},
				{1463, [4]int{122, 703, 803, -715}},
				{1469, [4]int{100, 802, 853, -838}},
				{1472, [4]int{157, 802, 704, -854}},
				{1480, [4]int{154, 802, 710, -848}},
				{1484, [4]int{94, 795, 836, -853}},
				{1488, [4]int{99, 773, 780, -830}},
				{1490, [4]int{100, 798, 810, -837}},
				{1494, [4]int{157, 511, 699, -538}},
				{1495, [4]int{69, 655, 843, -645}},
				{1498, [4]int{77, 727, 852, -766}},
				{1499, [4]int{196, 782, 647, -813}},
				{1502, [4]int{119, 764, 782, -799}},
				{1505, [4]int{83, 723, 829, -757}},
				{1525, [4]int{92, 795, 783, -845}},
				{1532, [4]int{173, 789, 684, -859}},
				{1541, [4]int{89, 801, 797, -849}},
				{1557, [4]int{81, 789, 788, -823}},
				{1562, [4]int{169, 614, 693, -680}},
				{1593, [4]int{324, 778, 528, -811}},
				{1600, [4]int{70, 704, 860, -698}},
				{1637, [4]int{108, 851, 842, -904}},
				{1645, [4]int{100, 440, 799, -111}},
				{9496, [4]int{42, 773, 921, -853}},
				{10927, [4]int{37, 845, 925, -928}},
				{11016, [4]int{33, 837, 943, -918}},
				{11864, [4]int{33, 843, 927, -925}},
				{11922, [4]int{40, 838, 928, -921}},
				{14025, [4]int{35, 798, 936, -881}},
				{15364, [4]int{69, 844, 865, -924}},
				{15585, [4]int{36, 842, 927, -924}},
				{16915, [4]int{29, 842, 928, -927}},
				{17513, [4]int{30, 842, 931, -912}},
				{20106, [4]int{37, 831, 925, -913}},
				{20282, [4]int{78, 800, 853, -889}},
				{20333, [4]int{32, 802, 933, -885}},
				{20698, [4]int{27, 790, 882, -875}},
				{20791, [4]int{46, 842, 908, -922}},
				{26323, [4]int{51, 776, 900, -813}},
				{27699, [4]int{84, 845, 848, -913}},
				{40641, [4]int{38, 793, 916, -874}},
				{43650, [4]int{59, 842, 882, -926}},
				{44199, [4]int{33, 791, 939, -875}},
				{59058, [4]int{141, 237, 203, -351}},
				{63163, [4]int{103, 138, 894, -151}},
			},
		},
		{
			Coord: 710,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{51, 574, 470, -588}},
				{70, [4]int{44, 574, 496, -588}},
				{77, [4]int{79, 798, 215, -812}},
				{78, [4]int{79, 574, 813, -574}},
				{81, [4]int{79, 574, 519, -789}},
				{84, [4]int{26, 574, 437, -588}},
				{85, [4]int{21, 712, 383, -726}},
				{89, [4]int{16, 560, 532, -560}},
				{1397, [4]int{39, 242, 324, -311}},
				{1398, [4]int{31, 248, 324, -324}},
				{1461, [4]int{91, 808, 834, -872}},
				{1463, [4]int{105, 717, 838, -739}},
				{1469, [4]int{80, 813, 888, -865}},
				{1472, [4]int{146, 818, 724, -881}},
				{1480, [4]int{142, 818, 731, -877}},
				{1484, [4]int{87, 806, 848, -881}},
				{1488, [4]int{89, 791, 797, -863}},
				{1490, [4]int{88, 811, 830, -863}},
				{1494, [4]int{142, 529, 731, -570}},
				{1495, [4]int{53, 674, 875, -679}},
				{1498, [4]int{69, 747, 867, -793}},
				{1499, [4]int{177, 798, 674, -842}},
				{1502, [4]int{102, 778, 807, -821}},
				{1505, [4]int{63, 743, 865, -790}},
				{1525, [4]int{84, 810, 809, -872}},
				{1532, [4]int{155, 806, 720, -893}},
				{1541, [4]int{71, 814, 831, -873}},
				{1557, [4]int{62, 811, 827, -856}},
				{1562, [4]int{148, 622, 722, -703}},
				{1593, [4]int{303, 792, 567, -836}},
				{1600, [4]int{49, 723, 903, -730}},
				{1637, [4]int{87, 865, 880, -936}},
				{1645, [4]int{91, 464, 817, -159}},
				{9496, [4]int{33, 783, 942, -871}},
				{10927, [4]int{31, 853, 938, -942}},
				{11016, [4]int{26, 844, 957, -932}},
				{11864, [4]int{23, 850, 940, -939}},
				{11922, [4]int{30, 845, 951, -938}},
				{14025, [4]int{26, 811, 952, -901}},
				{15364, [4]int{64, 850, 875, -938}},
				{15585, [4]int{28, 849, 943, -939}},
				{16915, [4]int{21, 850, 944, -941}},
				{17513, [4]int{23, 850, 947, -923}},
				{20106, [4]int{26, 840, 944, -931}},
				{20282, [4]int{68, 810, 874, -907}},
				{20333, [4]int{22, 816, 951, -908}},
				{20698, [4]int{17, 806, 906, -900}},
				{20791, [4]int{39, 850, 919, -939}},
				{26323, [4]int{46, 789, 911, -835}},
				{27699, [4]int{73, 851, 872, -935}},
				{40641, [4]int{24, 807, 948, -897}},
				{43650, [4]int{56, 850, 888, -941}},
				{44199, [4]int{23, 810, 964, -903}},
				{59058, [4]int{163, 242, 230, -380}},
				{63163, [4]int{93, 179, 939, -193}},
			},
		},
		{
			Coord: 800,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{49, 578, 482, -592}},
				{70, [4]int{42, 578, 504, -592}},
				{77, [4]int{76, 798, 231, -812}},
				{78, [4]int{76, 578, 828, -578}},
				{81, [4]int{76, 578, 530, -789}},
				{84, [4]int{25, 578, 445, -592}},
				{85, [4]int{20, 716, 394, -730}},
				{89, [4]int{16, 564, 548, -564}},
				{1397, [4]int{36, 247, 332, -319}},
				{1398, [4]int{29, 249, 329, -329}},
				{1461, [4]int{86, 812, 844, -880}},
				{1463, [4]int{100, 721, 849, -746}},
				{1469, [4]int{74, 816, 898, -873}},
				{1472, [4]int{143, 823, 729, -890}},
				{1480, [4]int{138, 823, 738, -886}},
				{1484, [4]int{85, 810, 852, -890}},
				{1488, [4]int{86, 796, 802, -872}},
				{1490, [4]int{81, 815, 839, -871}},
				{1494, [4]int{138, 535, 740, -581}},
				{1495, [4]int{48, 680, 885, -690}},
				{1498, [4]int{67, 753, 872, -801}},
				{1499, [4]int{171, 803, 682, -851}},
				{1502, [4]int{96, 783, 816, -828}},
				{1505, [4]int{57, 749, 876, -800}},
				{1525, [4]int{82, 814, 816, -880}},
				{1532, [4]int{149, 811, 732, -903}},
				{1541, [4]int{65, 818, 842, -880}},
				{1557, [4]int{56, 818, 840, -866}},
				{1562, [4]int{141, 625, 731, -711}},
				{1593, [4]int{296, 796, 579, -843}},
				{1600, [4]int{42, 729, 916, -740}},
				{1637, [4]int{80, 870, 892, -946}},
				{1645, [4]int{89, 472, 821, -175}},
				{9496, [4]int{30, 786, 949, -876}},
				{10927, [4]int{28, 856, 943, -947}},
				{11016, [4]int{24, 846, 962, -936}},
				{11864, [4]int{20, 852, 945, -943}},
				{11922, [4]int{27, 847, 958, -943}},
				{14025, [4]int{24, 815, 956, -907}},
				{15364, [4]int{62, 852, 879, -942}},
				{15585, [4]int{26, 851, 947, -943}},
				{16915, [4]int{18, 852, 949, -945}},
				{17513, [4]int{20, 852, 953, -927}},
				{20106, [4]int{22, 844, 950, -938}},
				{20282, [4]int{65, 813, 880, -913}},
				{20333, [4]int{19, 820, 956, -914}},
				{20698, [4]int{14, 811, 914, -908}},
				{20791, [4]int{35, 852, 924, -944}},
				{26323, [4]int{44, 793, 915, -842}},
				{27699, [4]int{70, 853, 879, -942}},
				{40641, [4]int{19, 811, 958, -903}},
				{43650, [4]int{55, 852, 890, -945}},
				{44199, [4]int{20, 816, 971, -912}},
				{59058, [4]int{169, 244, 240, -390}},
				{63163, [4]int{90, 192, 953, -206}},
			},
		},
		{
			Coord: 900,
			Extents: []extent{
				{1, [4]int{0, 0, 0, 0}},
				{66, [4]int{47, 583, 494, -597}},
				{70, [4]int{40, 583, 514, -597}},
				{77, [4]int{72, 799, 250, -813}},
				{78, [4]int{72, 583, 846, -583}},
				{81, [4]int{72, 583, 543, -790}},
				{84, [4]int{23, 583, 456, -597}},
				{85, [4]int{18, 720, 407, -734}},
				{89, [4]int{16, 569, 566, -569}},
				{1397, [4]int{33, 252, 341, -328}},
				{1398, [4]int{26, 250, 334, -334}},
				{1461, [4]int{81, 816, 854, -888}},
				{1463, [4]int{94, 726, 861, -755}},
				{1469, [4]int{67, 820, 910, -882}},
				{1472, [4]int{139, 828, 736, -899}},
				{1480, [4]int{134, 829, 745, -897}},
				{1484, [4]int{83, 814, 856, -900}},
				{1488, [4]int{82, 802, 808, -883}},
				{1490, [4]int{73, 819, 850, -879}},
				{1494, [4]int{133, 541, 751, -592}},
				{1495, [4]int{43, 687, 895, -702}},
				{1498, [4]int{64, 760, 877, -811}},
				{1499, [4]int{165, 808, 691, -860}},
				{1502, [4]int{90, 788, 825, -836}},
				{1505, [4]int{50, 756, 888, -812}},
				{1525, [4]int{79, 819, 825, -889}},
				{1532, [4]int{143, 817, 744, -915}},
				{1541, [4]int{58, 822, 855, -888}},
				{1557, [4]int{49, 826, 854, -878}},
				{1562, [4]int{134, 628, 741, -719}},
				{1593, [4]int{289, 801, 592, -852}},
				{1600, [4]int{35, 736, 931, -752}},
				{1637, [4]int{73, 875, 905, -957}},
				{1645, [4]int{86, 480, 827, -191}},
				{9496, [4]int{27, 790, 956, -883}},
				{10927, [4]int{23, 859, 950, -952}},
				{11016, [4]int{21, 849, 967, -942}},
				{11864, [4]int{17, 855, 949, -948}},
				{11922, [4]int{23, 849, 967, -949}},
				{14025, [4]int{21, 819, 962, -914}},
				{15364, [4]int{60, 854, 883, -947}},
				{15585, [4]int{23, 854, 953, -949}},
				{16915, [4]int{15, 855, 955, -950}},
				{17513, [4]int{18, 855, 958, -931}},
				{20106, [4]int{17, 848, 958, -947}},
				{20282, [4]int{62, 816, 887, -919}},
				{20333, [4]int{15, 825, 963, -922}},
				{20698, [4]int{10, 817, 923, -917}},
				{20791, [4]int{29, 855, 932, -950}},
				{26323, [4]int{42, 798, 919, -850}},
				{27699, [4]int{66, 855, 887, -949}},
				{40641, [4]int{14, 816, 969, -911}},
				{43650, [4]int{54, 855, 892, -951}},
				{44199, [4]int{17, 822, 979, -921}},
				{59058, [4]int{177, 246, 249, -401}},
				{63163, [4]int{86, 207, 969, -221}},
			},
		},
	}

	b, err := td.Files.ReadFile("common/NotoSansCJKjp-VF.otf")
	tu.AssertNoErr(t, err)

	ft, err := ot.NewLoader(bytes.NewReader(b))
	tu.AssertNoErr(t, err)

	font, err := NewFont(ft)
	tu.AssertNoErr(t, err)

	for _, data := range datas[2:3] {
		coords := font.NormalizeVariations([]float32{data.Coord})

		for _, glyph := range data.Extents[1:2] {
			_, bounds, err := font.cff2.LoadGlyph(glyph.GID, coords)
			tu.AssertNoErr(t, err)

			got := bounds.ToExtents()
			// harfbuzz round values
			got.Width = float32(math.Round(float64(got.Width)))
			got.Height = float32(math.Round(float64(got.Height)))

			exp := GlyphExtents{
				XBearing: float32(glyph.Extents[0]),
				YBearing: float32(glyph.Extents[1]),
				Width:    float32(glyph.Extents[2]),
				Height:   float32(glyph.Extents[3]),
			}

			tu.Assert(t, exp == got)
		}
	}
}
