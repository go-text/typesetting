package harfbuzz

import (
	"sort"
	"testing"

	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"
)

// ported from harfbuzz/test/api/test-aat-layout.c Copyright © 2018  Ebrahim Byagowi

func TestAATFeaturesSorted(t *testing.T) {
	var tags []int
	for _, f := range featureMappings {
		tags = append(tags, int(f.otFeatureTag))
	}
	if !sort.IntsAreSorted(tags) {
		t.Fatalf("expected sorted tags, got %v", tags)
	}
}

func aatLayoutGetFeatureTypes(feat tables.Feat) []aatLayoutFeatureType {
	out := make([]aatLayoutFeatureType, len(feat.Names))
	for i, f := range feat.Names {
		out[i] = f.Feature
	}
	return out
}

func aatLayoutFeatureTypeGetNameID(feat tables.Feat, feature uint16) int {
	if f := feat.GetFeature(feature); f != nil {
		return int(f.NameIndex)
	}
	return -1
}

func TestAatGetFeatureTypes(t *testing.T) {
	feat := openFontFile(t, "fonts/aat-feat.ttf").Feat

	features := aatLayoutGetFeatureTypes(feat)
	assertEqualInt(t, 11, len(feat.Names))

	assertEqualInt(t, 1, int(features[0]))
	assertEqualInt(t, 3, int(features[1]))
	assertEqualInt(t, 6, int(features[2]))

	assertEqualInt(t, 258, aatLayoutFeatureTypeGetNameID(feat, features[0]))
	assertEqualInt(t, 261, aatLayoutFeatureTypeGetNameID(feat, features[1]))
	assertEqualInt(t, 265, aatLayoutFeatureTypeGetNameID(feat, features[2]))
}

func TestAatHas(t *testing.T) {
	morx := openFontFile(t, "fonts/aat-morx.ttf")

	tu.Assert(t, len(morx.Morx) != 0)

	trak := openFontFile(t, "fonts/aat-trak.ttf")
	tu.Assert(t, !trak.Trak.IsEmpty())
}
