// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"bytes"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"
)

// wrap td.Files.ReadFile
func readFontFile(t testing.TB, filepath string) *ot.Loader {
	t.Helper()

	file, err := td.Files.ReadFile(filepath)
	tu.AssertNoErr(t, err)

	fp, err := ot.NewLoader(bytes.NewReader(file))
	tu.AssertNoErr(t, err)

	return fp
}

func readTable(t testing.TB, fl *ot.Loader, tag string) []byte {
	t.Helper()

	table, err := fl.RawTable(ot.MustNewTag(tag))
	tu.AssertNoErr(t, err)

	return table
}

func TestCrashes(t *testing.T) {
	for _, filepath := range append(tu.Filenames(t, "common"), "toys/chromacheck-svg.ttf") {
		loadFont(t, filepath)
	}
}

func TestGlyphName(t *testing.T) {
	ft := loadFont(t, "toys/NamesCFF.ttf")
	tu.Assert(t, ft.post.names == nil)

	expected := [20]string{
		".notdef", "space", "uni0622", "uni0623", "uni0624", "uni0625", "uni0626", "uni0628",
		"uni06C0", "uni06C2", "uni06D3", "uni0625.fina", "uni0623.fina", "uni0622.fina", "uni0628.fina",
		"uni0626.init", "uni0628.init", "uni0626.medi", "uni0628.medi", "uni06C1.fina",
	}
	for i, exp := range expected {
		tu.Assert(t, ft.GlyphName(GID(i)) == exp)
	}
}

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, filepath := range tu.Filenames(b, "common") {
			loadFont(b, filepath)
		}
	}
}

func TestMaxpAndHmtx(t *testing.T) {
	const maxp_data = "\x00\x00\x50\x00" + // version
		"\x00\x05" // numGlyphs

	const hhea_data = "\x00\x01\x00\x00" + /* FixedVersion<>version;	 * 0x00010000u for version 1.0. */
		"\x02\x00" + /* FWORD		ascender;	 * Typographic ascent. */
		"\x00\x10" + /* FWORD		descender;	 * Typographic descent. */
		"\x00\x00" + /* FWORD		lineGap;	 * Typographic line gap. */
		"\x00\x00" + /* UFWORD	advanceMax;	 * Maximum advance width/height value in metrics table. */
		"\x00\x00" + /* FWORD		minLeadingBearing;  * Minimum left/top sidebearing value in metrics table. */
		"\x00\x00" + /* FWORD		minTrailingBearing;  * Minimum right/bottom sidebearing value; */
		"\x01\x00" + /* FWORD		maxExtent;	 * horizontal: Max(lsb + (xMax - xMin)), */
		"\x00\x00" + /* HBINT16	caretSlopeRise;	 * Used to calculate the slope of the,*/
		"\x00\x00" + /* HBINT16	caretSlopeRun;	 * 0 for vertical caret, 1 for horizontal. */
		"\x00\x00" + /* HBINT16	caretOffset;	 * The amount by which a slanted */
		"\x00\x00" + /* HBINT16	reserved1;	 * Set to 0. */
		"\x00\x00" + /* HBINT16	reserved2;	 * Set to 0. */
		"\x00\x00" + /* HBINT16	reserved3;	 * Set to 0. */
		"\x00\x00" + /* HBINT16	reserved4;	 * Set to 0. */
		"\x00\x00" + /* HBINT16	metricDataFormat; * 0 for current format. */
		"\x00\x02" /* HBUINT16	numberOfLongMetrics;  * Number of LongMetric entries in metric table. */

	const hmtx_data = "\x00\x01\x00\x02" + /* glyph 0 advance lsb */
		"\x00\x03\x00\x04" + /* glyph 1 advance lsb */
		"\x00\x05" + /* glyph 2         lsb */
		"\x00\x06" + /* glyph 3         lsb */
		"\x00\x07" + /* glyph 4         lsb */
		"\x00\x08" + /* glyph 5 advance */
		"\x00\x09" /* glyph 6 advance */

	maxp, _, err := tables.ParseMaxp([]byte(maxp_data))
	tu.AssertNoErr(t, err)
	_, hmtx, err := loadHVtmx([]byte(hhea_data), []byte(hmtx_data), int(maxp.NumGlyphs))
	tu.AssertNoErr(t, err)

	tu.Assert(t, hmtx.Advance(0) == 1)
	tu.Assert(t, hmtx.Advance(1) == 3)
	tu.Assert(t, hmtx.Advance(2) == 3)
	tu.Assert(t, hmtx.Advance(3) == 3)
	tu.Assert(t, hmtx.Advance(4) == 3)

	// spec expansion
	// tu.Assert(t, hmtx.Advance(5) == 8)
	// tu.Assert(t, hmtx.Advance(6) == 9)
	// tu.Assert(t, hmtx.Advance(7) == 9)

	tu.Assert(t, hmtx.Advance(8) == 0)
	tu.Assert(t, hmtx.Advance(9) == 0)
	tu.Assert(t, hmtx.Advance(10) == 0)
	tu.Assert(t, hmtx.Advance(11) == 0)
}

func TestLoadCFF2(t *testing.T) {
	b, err := td.Files.ReadFile("common/NotoSansCJKjp-VF.otf")
	tu.AssertNoErr(t, err)

	ft, err := ot.NewLoader(bytes.NewReader(b))
	tu.AssertNoErr(t, err)

	font, err := NewFont(ft)
	tu.AssertNoErr(t, err)

	tu.Assert(t, font.cff2 != nil)
	tu.Assert(t, font.cff2.VarStore.AxisCount() == 1)
}
