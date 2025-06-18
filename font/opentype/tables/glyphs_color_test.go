package tables

import (
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

func TestCOLR(t *testing.T) {
	ft := readFontFile(t, "common/NotoColorEmoji-Regular.ttf")
	colr, err := ParseCOLR(readTable(t, ft, "COLR"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(colr.BaseGlyphRecords) == 0)
	tu.Assert(t, len(colr.LayerRecords) == 0)
	tu.Assert(t, len(colr.BaseGlyphList.PaintRecords) == 3845)
	tu.Assert(t, colr.BaseGlyphList.PaintRecords[0].Paint == PaintColrLayers{1, 3, 47625})
	tu.Assert(t, len(colr.LayerList.PaintTables) == 69264)
	tu.Assert(t, colr.ClipList.Clips[0].ClipBox == ClipBoxFormat1{1, 480, 192, 800, 512})
	tu.Assert(t, colr.VarIndexMap == nil && colr.ItemVariationStore == nil)

	ft = readFontFile(t, "common/CoralPixels-Regular.ttf")
	colr, err = ParseCOLR(readTable(t, ft, "COLR"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(colr.BaseGlyphRecords) == 335)
	tu.Assert(t, len(colr.LayerRecords) == 5603)
	g1, g2 := colr.BaseGlyphRecords[0], colr.BaseGlyphRecords[1]
	tu.Assert(t, g1 == BaseGlyph{0, 0, 11} && g2 == BaseGlyph{2, 11, 18})
	tu.Assert(t, colr.LayerRecords[0].PaletteIndex == 4)
}
