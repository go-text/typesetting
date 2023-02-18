package harfbuzz

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strings"
	"testing"

	td "github.com/go-text/typesetting-utils/harfbuzz"
	"github.com/go-text/typesetting/opentype/api"
	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/loader"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

// This is the main test suite for harfbuzz, which parses and runs
// the test cases directly copied from harfbuzz/test/shaping
//
// Adapted from harfbuzz/util/hb-shape.cc, main-font-text.hh

func TestShapeExpected(t *testing.T) {
	tests := collectTests(t)
	fmt.Printf("Running %d tests...\n", len(tests))

	for _, testD := range tests {
		runShapingTest(t, testD, false)
	}
}

func TestDebug(t *testing.T) {
	// This test is a shortcut to inspect one specific tests
	// when debugging
	t.Skip()

	dir := "harfbuzz_reference/aots/tests"
	testString := `../fonts/gpos4_lookupflag_f1.otf;--features="test" --no-clusters --no-glyph-names --ned;U+0011,U+0012,U+0011,U+0013,U+0011;[17|18@1500,0|17@3000,0|19@4500,0|17@4500,0]`
	testD := newTestData(t, dir, testString)
	runShapingTest(t, testD, true)
}

// Generates gidDDD if glyph has no name.
func (f *Font) glyphToString(glyph GID) string {
	if name := f.face.GlyphName(glyph); name != "" {
		return name
	}

	return fmt.Sprintf("gid%d", glyph)
}

// return a compact representation of the buffer contents
func (b *Buffer) serialize(font *Font, opt formatOpts) string {
	if len(b.Info) == 0 {
		return "" //  the reference does not return []
	}
	gs := new(strings.Builder)
	gs.WriteByte('[')
	var x, y Position
	for i, glyph := range b.Info {
		if opt.hideGlyphNames {
			fmt.Fprintf(gs, "%d", glyph.Glyph)
		} else {
			gs.WriteString(font.glyphToString(glyph.Glyph))
		}

		if !opt.hideClusters {
			fmt.Fprintf(gs, "=%d", glyph.Cluster)
		}
		pos := b.Pos[i]

		if !opt.hidePositions {
			if x+pos.XOffset != 0 || y+pos.YOffset != 0 {
				fmt.Fprintf(gs, "@%d,%d", x+pos.XOffset, y+pos.YOffset)
			}
			if !opt.hideAdvances {
				fmt.Fprintf(gs, "+%d", pos.XAdvance)
				if pos.YAdvance != 0 {
					fmt.Fprintf(gs, ",%d", pos.YAdvance)
				}
			}
		}

		if opt.showExtents {
			extents, _ := font.GlyphExtents(glyph.Glyph)
			fmt.Fprintf(gs, "<%d,%d,%d,%d>", extents.XBearing, extents.YBearing, extents.Width, extents.Height)
		}

		if i != len(b.Info)-1 {
			gs.WriteByte('|')
		}

		if opt.hideAdvances {
			x += pos.XAdvance
			y += pos.YAdvance
		}
	}
	gs.WriteByte(']')
	return gs.String()
}

func (fo *fontOpts) loadFont(t *testing.T) *Font {
	// create the blob
	tu.Assert(t, fo.fontRef.File != "")
	f, err := td.Files.ReadFile(fo.fontRef.File)
	tu.AssertNoErr(t, err)

	fonts, err := loader.NewLoaders(bytes.NewReader(f))
	tu.AssertNoErr(t, err)

	tu.Assert(t, int(fo.fontRef.Index) < len(fonts))
	ft, err := font.NewFont(fonts[fo.fontRef.Index])
	tu.AssertNoErr(t, err)

	// create the face
	face := font.Face{Font: ft, XPpem: fo.xPpem, YPpem: fo.yPpem}
	face.SetVariations(fo.variations)

	font := NewFont(&face)

	if fo.fontSizeX == fontSizeUpem {
		fo.fontSizeX = int(font.faceUpem)
	}
	if fo.fontSizeY == fontSizeUpem {
		fo.fontSizeY = int(font.faceUpem)
	}

	font.Ptem = float32(fo.ptem)

	scaleX := scalbnf(float64(fo.fontSizeX), fo.subpixelBits)
	scaleY := scalbnf(float64(fo.fontSizeY), fo.subpixelBits)
	font.XScale, font.YScale = scaleX, scaleY

	return font
}

// returnns x * 2^exp
func scalbnf(x float64, exp int) int32 {
	return int32(x * (math.Pow(2, float64(exp))))
}

func (so *shapeOpts) setupBuffer(buffer *Buffer) {
	buffer.Props = so.props
	var flags ShappingOptions
	if so.bot {
		flags |= Bot
	}
	if so.eot {
		flags |= Eot
	}
	if so.preserveDefaultIgnorables {
		flags |= PreserveDefaultIgnorables
	}
	if so.removeDefaultIgnorables {
		flags |= RemoveDefaultIgnorables
	}
	buffer.Flags = flags
	buffer.Invisible = so.invisibleGlyph
	buffer.ClusterLevel = so.clusterLevel
	buffer.GuessSegmentProperties()
}

func copyBufferProperties(dst, src *Buffer) {
	dst.Props = src.Props
	dst.Flags = src.Flags
	dst.ClusterLevel = src.ClusterLevel
}

func appendBuffer(dst, src *Buffer, start, end int) {
	origLen := len(dst.Info)

	dst.Info = append(dst.Info, src.Info[start:end]...)
	dst.Pos = append(dst.Pos, src.Pos[start:end]...)

	/* pre-context */
	if origLen == 0 && start+len(src.context[0]) > 0 {
		dst.clearContext(0)
		for start > 0 && len(dst.context[0]) < contextLength {
			start--
			dst.context[0] = append(dst.context[0], src.Info[start].codepoint)
		}

		for i := 0; i < len(src.context[0]) && len(dst.context[0]) < contextLength; i++ {
			dst.context[0] = append(dst.context[0], src.context[0][i])
		}
	}

	/* post-context */
	dst.clearContext(1)
	for end < len(src.Info) && len(dst.context[1]) < contextLength {
		dst.context[1] = append(dst.context[1], src.Info[end].codepoint)
		end++
	}
	for i := 0; i < len(src.context[1]) && len(dst.context[1]) < contextLength; i++ {
		dst.context[1] = append(dst.context[1], src.context[1][i])
	}
}

func (ti *testInput) populateBuffer() *Buffer {
	buffer := NewBuffer()

	if ti.textBefore != nil {
		buffer.AddRunes(ti.textBefore, len(ti.textBefore), 0)
	}

	buffer.AddRunes(ti.text, 0, len(ti.text))

	if ti.textAfter != nil {
		buffer.AddRunes(ti.textAfter, 0, 0)
	}

	ti.shaper.setupBuffer(buffer)

	return buffer
}

func (so *shapeOpts) shape(font *Font, buffer *Buffer, verify bool) error {
	var textBuffer *Buffer

	if verify {
		textBuffer = NewBuffer()
		appendBuffer(textBuffer, buffer, 0, len(buffer.Info))
	}

	features, err := so.parseFeatures()
	if err != nil {
		return err
	}
	buffer.Shape(font, features)

	if verify {
		if err := so.verifyBuffer(buffer, textBuffer, font); err != nil {
			return err
		}
	}

	return nil
}

func (so *shapeOpts) verifyBuffer(buffer, textBuffer *Buffer, font *Font) error {
	if err := so.verifyBufferMonotone(buffer); err != nil {
		return err
	}
	if err := so.verifyBufferSafeToBreak(buffer, textBuffer, font); err != nil {
		return err
	}
	if err := so.verifyValidGID(buffer, font); err != nil {
		log.Println(err)
	}
	return nil
}

func (so *shapeOpts) verifyValidGID(buffer *Buffer, font *Font) error {
	for _, glyph := range buffer.Info {
		_, ok := font.GlyphExtents(glyph.Glyph)
		if !ok {
			return fmt.Errorf("Unknow glyph %d in font", glyph.Glyph)
		}
	}
	return nil
}

/* Check that clusters are monotone. */
func (so *shapeOpts) verifyBufferMonotone(buffer *Buffer) error {
	if so.clusterLevel == MonotoneGraphemes || so.clusterLevel == MonotoneCharacters {
		isForward := buffer.Props.Direction.isForward()

		info := buffer.Info

		for i := 1; i < len(info); i++ {
			if info[i-1].Cluster != info[i].Cluster && (info[i-1].Cluster < info[i].Cluster) != isForward {
				return fmt.Errorf("cluster at index %d is not monotone", i)
			}
		}
	}

	return nil
}

func (so *shapeOpts) verifyBufferSafeToBreak(buffer, textBuffer *Buffer, font *Font) error {
	if so.clusterLevel != MonotoneGraphemes && so.clusterLevel != MonotoneCharacters {
		/* Cannot perform this check without monotone clusters.
		 * Then again, unsafe-to-break flag is much harder to use without
		 * monotone clusters. */
		return nil
	}

	/* Check that breaking up shaping at safe-to-break is indeed safe. */

	fragment, reconstruction := NewBuffer(), NewBuffer()
	copyBufferProperties(reconstruction, buffer)

	info := buffer.Info
	text := textBuffer.Info

	/* Chop text and shape fragments. */
	forward := buffer.Props.Direction.isForward()
	start := 0
	textStart := len(textBuffer.Info)
	if forward {
		textStart = 0
	}
	textEnd := textStart
	for end := 1; end < len(buffer.Info)+1; end++ {
		offset := 1
		if forward {
			offset = 0
		}
		if end < len(buffer.Info) && (info[end].Cluster == info[end-1].Cluster ||
			info[end-offset].Mask&GlyphUnsafeToBreak != 0) {
			continue
		}

		/* Shape segment corresponding to glyphs start..end. */
		if end == len(buffer.Info) {
			if forward {
				textEnd = len(textBuffer.Info)
			} else {
				textStart = 0
			}
		} else {
			if forward {
				cluster := info[end].Cluster
				for textEnd < len(textBuffer.Info) && text[textEnd].Cluster < cluster {
					textEnd++
				}
			} else {
				cluster := info[end-1].Cluster
				for textStart != 0 && text[textStart-1].Cluster >= cluster {
					textStart--
				}
			}
		}
		if !(textStart < textEnd) {
			return fmt.Errorf("unexpected %d >= %d", textStart, textEnd)
		}

		if debugMode >= 1 {
			fmt.Println()
			fmt.Printf("VERIFY SAFE TO BREAK : start %d end %d text start %d end %d\n", start, end, textStart, textEnd)
			fmt.Println()
		}

		fragment.Clear()
		copyBufferProperties(fragment, buffer)

		flags := fragment.Flags
		if 0 < textStart {
			flags = (flags & ^Bot)
		}
		if textEnd < len(textBuffer.Info) {
			flags = (flags & ^Eot)
		}
		fragment.Flags = flags

		appendBuffer(fragment, textBuffer, textStart, textEnd)
		features, err := so.parseFeatures()
		if err != nil {
			return err
		}
		fragment.Shape(font, features)
		appendBuffer(reconstruction, fragment, 0, len(fragment.Info))

		start = end
		if forward {
			textStart = textEnd
		} else {
			textEnd = textStart
		}
	}

	diff := bufferDiff(reconstruction, buffer, ^GID(0), 0)
	if diff != bufferDiffFlagEqual {
		/* Return the reconstructed result instead so it can be inspected. */
		buffer.Info = nil
		buffer.Pos = nil
		appendBuffer(buffer, reconstruction, 0, len(reconstruction.Info))

		return fmt.Errorf("safe-to-break test failed: %d", diff)
	}

	return nil
}

// returns the serialized shaped output
// if `verify` is true, additional check on buffer contents is performed
func (mft testInput) shape(t *testing.T, verify bool) string {
	buffer := mft.populateBuffer()

	font := mft.fontOpts.loadFont(t)
	err := mft.shaper.shape(font, buffer, verify)
	tu.AssertNoErr(t, err)

	return buffer.serialize(font, mft.format)
}

// harfbuzz seems to be OK with an invalid font
// in pratice, it seems useless to do shaping without
// font, so we dont support it, meaning we skip this test
func skipInvalidFontIndex(t *testing.T, ft api.FontID) bool {
	f, err := td.Files.ReadFile(ft.File)
	tu.AssertNoErr(t, err)

	fonts, err := loader.NewLoaders(bytes.NewReader(f))
	tu.AssertNoErr(t, err)

	if int(ft.Index) >= len(fonts) {
		t.Logf("skipping invalid font index %d in font %s\n", ft.Index, ft.File)
		return true
	}
	return false
}

// skipVerify is true when debugging, to reduce stdout clutter
func runShapingTest(t *testing.T, test testData, skipVerify bool) {
	if skipInvalidFontIndex(t, test.input.fontOpts.fontRef) {
		return
	}

	verify := test.expected != "*"

	// actual does the shaping
	got := test.input.shape(t, !skipVerify && verify)
	got = strings.TrimSpace(got)

	if verify {
		tu.AssertC(t, test.expected == got, fmt.Sprintf("%s\n%s\n expected :\n%s\n got \n%s", test.originDir, test.originLine, test.expected, got))
	}
}
