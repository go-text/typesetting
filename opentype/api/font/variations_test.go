// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"reflect"
	"testing"

	"github.com/go-text/typesetting/opentype/api"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

// ported from harfbuzz/test/api/test-var-coords.c Copyright © 2019 Ebrahim Byagowi

func TestVar(t *testing.T) {
	font := loadFont(t, "toys/CFF2-VF.otf")

	/* Design coords as input */
	designCoords := []float32{206.}
	coords := font.NormalizeVariations(designCoords)
	tu.Assert(t, coords[0]*(1<<14) == float32(-16116.88))

	// test for crash
	for weight := float32(200); weight < 901; weight++ {
		font.NormalizeVariations([]float32{weight})
	}

	face := Face{Font: font}
	face.SetVariations([]Variation{{loader.MustNewTag("wght"), 206.}})
	tu.Assert(t, len(face.Coords) == 1)
	tu.Assert(t, face.Coords[0] == -0.9836963)

	face.SetVariations(nil)
	tu.Assert(t, len(face.Coords) == 0)
}

func TestGlyphExtentsVar(t *testing.T) {
	font := loadFont(t, "common/SourceSans-VF-HVAR.ttf")

	coords := font.NormalizeVariations([]float32{500})
	face := Face{Font: font, Coords: coords}

	ext2, _ := face.GlyphExtents(2)

	tu.Assert(t, ext2 == api.GlyphExtents{XBearing: 50.192135, YBearing: 667.1601, Width: 591.8152, Height: -679.1601})
}

func TestGetDefaultCoords(t *testing.T) {
	tf := fvar{
		{Tag: loader.MustNewTag("wght"), Minimum: 38, Default: 88, Maximum: 250},
		{Tag: loader.MustNewTag("wdth"), Minimum: 60, Default: 402, Maximum: 402},
		{Tag: loader.MustNewTag("opsz"), Minimum: 10, Default: 14, Maximum: 72},
	}

	vars := []Variation{
		{Tag: loader.MustNewTag("wdth"), Value: 60},
	}
	coords := tf.getDesignCoordsDefault(vars)
	tu.Assert(t, reflect.DeepEqual(coords, []float32{88, 60, 14}))
}

func TestNormalizeVar(t *testing.T) {
	tf := fvar{
		{Tag: loader.MustNewTag("wdth"), Minimum: 60, Default: 402, Maximum: 500},
	}

	vars := []Variation{
		{Tag: loader.MustNewTag("wdth"), Value: 60},
	}
	coords := tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []float32{-1}))

	vars = []Variation{
		{Tag: loader.MustNewTag("wdth"), Value: 30},
	}
	coords = tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []float32{-1}))

	vars = []Variation{
		{Tag: loader.MustNewTag("wdth"), Value: 700},
	}
	coords = tf.normalizeCoordinates(tf.getDesignCoordsDefault(vars))
	tu.Assert(t, reflect.DeepEqual(coords, []float32{1}))
}

func TestAdvanceHVar(t *testing.T) {
	font := loadFont(t, "common/Commissioner-VF.ttf")
	coords := []float32{-0.4, 0, 0.8, 1}
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
		tu.Assert(t, exp == got)
	}
}

func TestAdvanceNoHVar(t *testing.T) {
	font := loadFont(t, "toys/GVAR-no-HVAR.ttf")

	tu.Assert(t, len(font.fvar) == 2)

	vars := []Variation{
		{Tag: loader.MustNewTag("wght"), Value: 600},
		{Tag: loader.MustNewTag("wght"), Value: 80},
	}
	face := Face{Font: font}
	face.SetVariations(vars)

	// 0 - 14 GIDs
	exps := [15]float32{600, 1164, 1170, 813, 741, 1164, 1170, 813, 741, 270, 270, 0, 0, 0, 0}

	for i, exp := range exps {
		got := face.HorizontalAdvance(api.GID(i))
		tu.Assert(t, exp == got)
	}
}
