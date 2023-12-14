package harfbuzz

import (
	"testing"

	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

// ported from harfbuzz/test/api/test-buffer.c Copyright © 2011  Google, Inc. Behdad Esfahbod

var utf32 = [7]rune{'a', 'b', 0x20000, 'd', 'e', 'f', 'g'}

const (
	bufferEmpty = iota
	bufferOneByOne
	bufferUtf32
	bufferNumTypes
)

func newTestBuffer(kind int) *Buffer {
	b := NewBuffer()

	switch kind {
	case bufferEmpty:

	case bufferOneByOne:
		for i := 1; i < len(utf32)-1; i++ {
			b.AddRune(utf32[i], i)
		}

	case bufferUtf32:
		b.AddRunes(utf32[:], 1, len(utf32)-2)

	}
	return b
}

func testBufferProperties(b *Buffer, t *testing.T) {
	/* test default properties */

	tu.Assert(t, b.Props.Direction == 0)
	tu.Assert(t, b.Props.Script == 0)
	tu.Assert(t, b.Props.Language == "")

	b.Props.Language = language.NewLanguage("fa")
	tu.Assert(t, b.Props.Language == language.NewLanguage("Fa"))

	// test Clear clears all properties
	b.Props.Direction = RightToLeft
	b.Props.Script = language.Arabic
	b.Props.Language = language.NewLanguage("fa")
	b.Flags = Bot
	b.Clear()

	tu.Assert(t, b.Props.Direction == 0)
	tu.Assert(t, b.Props.Script == 0)
	tu.Assert(t, b.Props.Language == "")
	tu.Assert(t, b.Flags == 0)
	tu.Assert(t, b.NotFound == 0)
}

func testBufferContents(b *Buffer, kind int, t *testing.T) {
	if kind == bufferEmpty {
		assertEqualInt(t, len(b.Info), 0)
		return
	}

	glyphs := b.Info
	L := len(glyphs)
	assertEqualInt(t, 5, L)
	assertEqualInt(t, 5, len(b.Pos))

	for _, g := range glyphs {
		assertEqualInt(t, int(g.Mask), 0)
		assertEqualInt(t, int(g.glyphProps), 0)
		assertEqualInt(t, int(g.ligProps), 0)
		assertEqualInt(t, int(g.syllable), 0)
		assertEqualInt(t, int(g.unicode), 0)
		assertEqualInt(t, int(g.complexAux), 0)
		assertEqualInt(t, int(g.complexCategory), 0)
	}

	for i, g := range glyphs {
		cluster := 1 + i
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
		assertEqualInt(t, g.Cluster, cluster)
	}

	/* reverse, test, and reverse back */

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[L-i]))
	}

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
	}

	// reverse_clusters works same as reverse for now since each codepoint is
	// in its own cluster

	b.reverseClusters()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[L-i]))
	}

	b.reverseClusters()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
	}

	/* now form a cluster and test again */
	glyphs[2].Cluster = glyphs[1].Cluster

	/* reverse, test, and reverse back */

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[L-i]))
	}

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
	}

	// reverse_clusters twice still should return the original string,
	// but when applied once, the 1-2 cluster should be retained.

	b.reverseClusters()
	for i, g := range glyphs {
		j := L - 1 - i
		if j == 1 {
			j = 2
		} else if j == 2 {
			j = 1
		}
		assertEqualInt(t, int(g.codepoint), int(utf32[1+j]))
	}

	b.reverseClusters()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
	}

	/* test reset clears content */

	b.Clear()
	assertEqualInt(t, len(b.Info), 0)
	assertEqualInt(t, len(b.Pos), 0)
}

func testBufferPositions(b *Buffer, t *testing.T) {
	/* Without shaping, positions should all be zero */
	assertEqualInt(t, len(b.Info), len(b.Pos))
	for _, pos := range b.Pos {
		assertEqualInt(t, 0, int(pos.XAdvance))
		assertEqualInt(t, 0, int(pos.YAdvance))
		assertEqualInt(t, 0, int(pos.XOffset))
		assertEqualInt(t, 0, int(pos.YOffset))
		assertEqualInt(t, 0, int(pos.attachChain))
		assertEqualInt(t, 0, int(pos.attachType))
	}

	//    /* test reset clears content */
	//    hb_buffer_reset (b);
	//    assertEqualInt (t, hb_buffer_get_length (b), ==, 0);
}

func TestBuffer(t *testing.T) {
	for i := 0; i < bufferNumTypes; i++ {
		buffer := newTestBuffer(i)

		testBufferContents(buffer, i, t)
		testBufferPositions(buffer, t)
		testBufferProperties(buffer, t) // clear the buffer
	}
}

/*
 * Comparing buffers.
 */

// Flags from comparing two buffers.
//
// For buffers with differing length, the per-glyph comparison is not
// attempted, though we do still scan reference buffer for dotted circle and
// `.notdef` glyphs.
//
// If the buffers have the same length, we compare them glyph-by-glyph and
// report which aspect(s) of the glyph info/position are different.
const (

	/* For buffers with differing length, the per-glyph comparison is not
	 * attempted, though we do still scan reference for dottedcircle / .notdef
	 * glyphs. */
	bdfLengthMismatch = 1 << iota

	/* We want to know if dottedcircle / .notdef glyphs are present in the
	 * reference, as we may not care so much about other differences in this
	 * case. */
	bdfNotdefPresent
	bdfDottedCirclePresent

	/* If the buffers have the same length, we compare them glyph-by-glyph
	 * and report which aspect(s) of the glyph info/position are different. */
	bdfCodepointMismatch
	bdfClusterMismatch
	bdfGlyphFlagsMismatch
	bdfPositionMismatch

	bufferDiffFlagEqual = 0x0000
)

/**
 * hb_buffer_diff:
 * @buffer: a buffer.
 * @reference: other buffer to compare to.
 * @dottedcircleGlyph: glyph id of U+25CC DOTTED CIRCLE, or (hb_codepont_t) -1.
 * @positionFuzz: allowed absolute difference in position values.
 *
 * If dottedcircleGlyph is (hb_codepoint_t) -1 then #bdfDottedCirclePresent
 * and #bdfNotdefPresent are never returned.  This should be used by most
 * callers if just comparing two buffers is needed.
 *
 * Since: 1.5.0
 **/

func bufferDiff(buffer, reference *Buffer, dottedcircleGlyph GID, positionFuzz int32) int {
	result := bufferDiffFlagEqual
	contains := dottedcircleGlyph != ^GID(0)

	count := len(reference.Info)

	if len(buffer.Info) != count {
		/*
		 * we can't compare glyph-by-glyph, but we do want to know if there
		 * are .notdef or dottedcircle glyphs present in the reference buffer
		 */
		info := reference.Info
		for i := 0; i < count; i++ {
			if contains && info[i].Glyph == dottedcircleGlyph {
				result |= bdfDottedCirclePresent
			}
			if contains && info[i].Glyph == 0 {
				result |= bdfNotdefPresent
			}
		}
		result |= bdfLengthMismatch
		return result
	}

	if count == 0 {
		return result
	}

	bufInfo := buffer.Info
	refInfo := reference.Info
	for i := 0; i < count; i++ {
		if bufInfo[i].codepoint != refInfo[i].codepoint {
			result |= bdfCodepointMismatch
		}
		if bufInfo[i].Cluster != refInfo[i].Cluster {
			result |= bdfClusterMismatch
		}
		if (bufInfo[i].Mask^refInfo[i].Mask)&glyphFlagDefined != 0 {
			result |= bdfGlyphFlagsMismatch
		}
		if contains && refInfo[i].Glyph == dottedcircleGlyph {
			result |= bdfDottedCirclePresent
		}
		if contains && refInfo[i].Glyph == 0 {
			result |= bdfNotdefPresent
		}
	}

	isDifferent := func(a, b int32) bool {
		d := a - b
		if d < 0 {
			d = -d
		}
		return d > positionFuzz
	}

	bufPos := buffer.Pos
	refPos := reference.Pos
	for i := 0; i < count; i++ {
		if isDifferent(bufPos[i].XAdvance, refPos[i].XAdvance) ||
			isDifferent(bufPos[i].YAdvance, refPos[i].YAdvance) ||
			isDifferent(bufPos[i].XOffset, refPos[i].XOffset) ||
			isDifferent(bufPos[i].YOffset, refPos[i].YOffset) {
			result |= bdfPositionMismatch
			break
		}
	}

	return result
}
