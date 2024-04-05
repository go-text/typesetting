package harfbuzz

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strings"
	"testing"

	td "github.com/go-text/typesetting-utils/harfbuzz"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

// This is the main test suite for harfbuzz, which parses and runs
// the test cases directly copied from harfbuzz/test/shaping
//
// Adapted from harfbuzz/util/hb-shape.cc, main-font-text.hh

func TestShapeExpected(t *testing.T) {
	tests := collectTests(t)

	// add tests based on the C++ binary
	tests = append(tests,
		// check we properly scale the offset values
		newTestData(t, "", "perf_reference/fonts/Roboto-Regular.ttf;--direction=ttb --font-size=2000;U+0061,U+0062;[gid70=0@-544,-1700+0,-2343|gid71=1@-562,-1912+0,-2343]"),
		newTestData(t, "", "perf_reference/fonts/Roboto-Regular.ttf;--direction=ttb --font-size=3000;U+0061,U+0062;[gid70=0@-816,-2550+0,-3515|gid71=1@-842,-2868+0,-3515]"),
	)

	fmt.Printf("Running %d tests...\n", len(tests))

	for _, testD := range tests {
		runShapingTest(t, testD, false)
	}
}

func TestDebug(t *testing.T) {
	// This test is a shortcut to inspect one specific test
	// when debugging
	t.Skip()

	// dir := "harfbuzz_reference/in-house/tests"
	testString := `fonts/AdobeBlank2.ttf;--no-glyph-names --no-positions;U+1F1E6,U+1F1E8;[1=0|1=0]`
	testD := newTestData(t, ".", testString)
	out := runShapingTest(t, testD, true)
	fmt.Println(out)
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

		if opt.showFlags {
			if mask := glyph.Mask & glyphFlagDefined; mask != 0 {
				fmt.Fprintf(gs, "#%d", mask)
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

	fonts, err := ot.NewLoaders(bytes.NewReader(f))
	tu.AssertNoErr(t, err)

	tu.Assert(t, int(fo.fontRef.Index) < len(fonts))
	ft, err := font.NewFont(fonts[fo.fontRef.Index])
	tu.AssertNoErr(t, err)

	// create the face
	face := font.NewFace(ft)
	face.SetPpem(fo.xPpem, fo.yPpem)
	face.SetVariations(fo.variations)

	font := NewFont(face)

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
	if so.unsafeToConcat {
		flags |= ProduceUnsafeToConcat
	}
	if so.safeToInsertTatweel {
		flags |= ProduceSafeToInsertTatweel
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
		err := so.verifyBuffer(buffer, textBuffer, font, features)
		return err
	}

	return nil
}

func (so *shapeOpts) verifyBuffer(buffer, textBuffer *Buffer, font *Font, features []Feature) error {
	if err := buffer.verifyMonotone(); err != nil {
		return err
	}
	if err := buffer.verifyUnsafeToBreak(textBuffer, font, features); err != nil {
		return err
	}
	if err := buffer.verifyValidGID(font); err != nil {
		log.Println(err)
	}
	return nil
}

// returns the serialized shaped output
// if `verify` is true, additional check on buffer contents is performed
func (mft testInput) shape(t *testing.T, verify bool) (string, error) {
	buffer := mft.populateBuffer()

	font := mft.fontOpts.loadFont(t)
	err := mft.shaper.shape(font, buffer, verify)
	if err != nil {
		return "", err
	}

	return buffer.serialize(font, mft.format), nil
}

// harfbuzz seems to be OK with an invalid font
// in pratice, it seems useless to do shaping without
// font, so we dont support it, meaning we skip this test
func skipInvalidFontIndex(t *testing.T, ft font.FontID) bool {
	f, err := td.Files.ReadFile(ft.File)
	tu.AssertNoErr(t, err)

	fonts, err := ot.NewLoaders(bytes.NewReader(f))
	tu.AssertNoErr(t, err)

	if int(ft.Index) >= len(fonts) {
		t.Logf("skipping invalid font index %d in font %s\n", ft.Index, ft.File)
		return true
	}
	return false
}

// skipVerify should be true when debugging, to reduce stdout clutter
// it returns the serialized output
func runShapingTest(t *testing.T, test testData, skipVerify bool) string {
	t.Helper()

	if skipInvalidFontIndex(t, test.input.fontOpts.fontRef) {
		return ""
	}

	verify := test.expected != "*"

	// actual does the shaping
	got, err := test.input.shape(t, !skipVerify && verify)
	if err != nil {
		t.Fatalf("for input %s: %s", test.originLine, err)
	}
	got = strings.TrimSpace(got)

	if verify {
		tu.AssertC(t, test.expected == got, fmt.Sprintf("%s\n%s\n expected :\n%s\n got \n%s", test.originDir, test.originLine, test.expected, got))
	}
	return got
}
