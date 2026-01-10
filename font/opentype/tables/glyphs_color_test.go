package tables

import (
	"testing"

	tu "github.com/unidoc/typesetting/testutils"
)

func TestCOLR(t *testing.T) {
	ft := readFontFile(t, "color/NotoColorEmoji-Regular.ttf")
	colr, err := ParseCOLR(readTable(t, ft, "COLR"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(colr.baseGlyphRecords) == 0)
	tu.Assert(t, len(colr.layerRecords) == 0)
	tu.Assert(t, len(colr.baseGlyphList.paintRecords) == 3845)
	tu.Assert(t, colr.baseGlyphList.paintRecords[0].Paint == PaintColrLayers{1, 3, 47625})
	tu.Assert(t, colr.ClipList.clips[0].ClipBox == ClipBoxFormat1{1, 480, 192, 800, 512})
	tu.Assert(t, colr.VarIndexMap == nil && colr.ItemVariationStore == nil)
	tu.Assert(t, len(colr.LayerList.paintTables) == 69264)

	clipBox, ok := colr.ClipList.Search(87)
	tu.Assert(t, ok && clipBox == ClipBoxFormat1{1, 64, -224, 1216, 928})

	// reference from fonttools
	paint := colr.LayerList.paintTables[6]
	transform, ok := paint.(PaintTransform)
	tu.Assert(t, ok)
	_, innerOK := transform.Paint.(PaintGlyph)
	tu.Assert(t, transform.Transform == Affine2x3{1, 0, 0, 1, 4.3119965, 0.375})
	tu.Assert(t, innerOK)

	_, ok = colr.Search(1)
	tu.Assert(t, !ok)
	_, ok = colr.Search(0xFFFF)
	tu.Assert(t, !ok)

	pt, ok := colr.Search(12)
	asColrLayers, ok2 := pt.(PaintColrLayers)
	tu.Assert(t, ok && ok2)
	tu.Assert(t, asColrLayers == PaintColrLayers{1, 9, 2427})

	for _, paint := range colr.baseGlyphList.paintRecords {
		if layers, ok := paint.Paint.(PaintColrLayers); ok {
			l, err := colr.LayerList.Resolve(layers)
			tu.AssertNoErr(t, err)
			tu.Assert(t, len(l) == int(layers.NumLayers))
		}
	}

	ft = readFontFile(t, "color/CoralPixels-Regular.ttf")
	colr, err = ParseCOLR(readTable(t, ft, "COLR"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(colr.baseGlyphRecords) == 335)
	tu.Assert(t, len(colr.layerRecords) == 5603)
	g1, g2 := colr.baseGlyphRecords[0], colr.baseGlyphRecords[1]
	tu.Assert(t, g1 == baseGlyph{0, 0, 11} && g2 == baseGlyph{2, 11, 18})
	tu.Assert(t, colr.layerRecords[0].PaletteIndex == 4)

	_, ok = colr.Search(1)
	tu.Assert(t, !ok)
	_, ok = colr.Search(0xFFFF)
	tu.Assert(t, !ok)

	pt, ok = colr.Search(0)
	asLayers, ok2 := pt.(PaintColrLayersResolved)
	tu.Assert(t, ok && ok2)
	tu.Assert(t, len(asLayers) == 11)
	tu.Assert(t, asLayers[0].PaletteIndex == 4)
	tu.Assert(t, asLayers[10].PaletteIndex == 11)
}

func TestCPAL(t *testing.T) {
	ft := readFontFile(t, "color/NotoColorEmoji-Regular.ttf")
	cpal, _, err := ParseCPAL(readTable(t, ft, "CPAL"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, cpal.Version == 0)
	tu.Assert(t, cpal.NumPaletteEntries == 5921)
	tu.Assert(t, cpal.numPalettes == 1 && len(cpal.ColorRecordIndices) == 1)

	ft = readFontFile(t, "color/CoralPixels-Regular.ttf")
	cpal, _, err = ParseCPAL(readTable(t, ft, "CPAL"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, cpal.Version == 0)
	tu.Assert(t, cpal.NumPaletteEntries == 32)
	tu.Assert(t, cpal.numPalettes == 2 && len(cpal.ColorRecordIndices) == 2)
}
