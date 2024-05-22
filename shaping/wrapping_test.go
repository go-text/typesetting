package shaping

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/segmenter"
	tu "github.com/go-text/typesetting/testutils"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// glyphs returns a slice of glyphs with clusters from start to
// end. If start is greater than end, the glyphs will be returned
// with descending cluster values.
func glyphs(start, end int) []Glyph {
	inc := 1
	if start > end {
		inc = -inc
	}
	num := max(start, end) - min(start, end) + 1
	g := make([]Glyph, 0, num)
	for i := 0; i < num; i++ {
		g = append(g, simpleGlyph(start))
		start += inc
	}
	return g
}

func TestMapRunesToClusterIndices(t *testing.T) {
	type testcase struct {
		name     string
		dir      di.Direction
		runes    Range
		glyphs   []Glyph
		expected []int
	}
	for _, tc := range []testcase{
		{
			name:  "simple",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(0),
				simpleGlyph(1),
				simpleGlyph(2),
				simpleGlyph(3),
				simpleGlyph(4),
			},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:  "simple offset",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5, Offset: 5},
			glyphs: []Glyph{
				simpleGlyph(5),
				simpleGlyph(6),
				simpleGlyph(7),
				simpleGlyph(8),
				simpleGlyph(9),
			},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:  "simple rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(4),
				simpleGlyph(3),
				simpleGlyph(2),
				simpleGlyph(1),
				simpleGlyph(0),
			},
			expected: []int{4, 3, 2, 1, 0},
		},
		{
			name:  "simple offset rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5, Offset: 5},
			glyphs: []Glyph{
				simpleGlyph(9),
				simpleGlyph(8),
				simpleGlyph(7),
				simpleGlyph(6),
				simpleGlyph(5),
			},
			expected: []int{4, 3, 2, 1, 0},
		},
		{
			name:  "fused clusters",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				complexGlyph(0, 2, 2),
				complexGlyph(0, 2, 2),
				simpleGlyph(2),
				complexGlyph(3, 2, 2),
				complexGlyph(3, 2, 2),
			},
			expected: []int{0, 0, 2, 3, 3},
		},
		{
			name:  "fused clusters rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				complexGlyph(3, 2, 2),
				complexGlyph(3, 2, 2),
				simpleGlyph(2),
				complexGlyph(0, 2, 2),
				complexGlyph(0, 2, 2),
			},
			expected: []int{3, 3, 2, 0, 0},
		},
		{
			name:  "ligatures",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				ligatureGlyph(0, 2),
				simpleGlyph(2),
				ligatureGlyph(3, 2),
			},
			expected: []int{0, 0, 1, 2, 2},
		},
		{
			name:  "ligatures rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				ligatureGlyph(3, 2),
				simpleGlyph(2),
				ligatureGlyph(0, 2),
			},
			expected: []int{2, 2, 1, 0, 0},
		},
		{
			name:  "expansion",
			dir:   di.DirectionLTR,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(0),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				simpleGlyph(2),
				simpleGlyph(3),
				simpleGlyph(4),
			},
			expected: []int{0, 1, 4, 5, 6},
		},
		{
			name:  "expansion rtl",
			dir:   di.DirectionRTL,
			runes: Range{Count: 5},
			glyphs: []Glyph{
				simpleGlyph(4),
				simpleGlyph(3),
				simpleGlyph(2),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				expansionGlyph(1, 3),
				simpleGlyph(0),
			},
			expected: []int{6, 3, 2, 1, 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mapping := mapRunesToClusterIndices(tc.dir, tc.runes, tc.glyphs, nil)
			if !reflect.DeepEqual(tc.expected, mapping) {
				t.Errorf("expected %v, got %v", tc.expected, mapping)
			}
			mapping = mapRunesToClusterIndices2(tc.dir, tc.runes, tc.glyphs, nil)
			if !reflect.DeepEqual(tc.expected, mapping) {
				t.Errorf("expected %v, got %v", tc.expected, mapping)
			}
			mapping = mapRunesToClusterIndices3(tc.dir, tc.runes, tc.glyphs, nil)
			if !reflect.DeepEqual(tc.expected, mapping) {
				t.Errorf("expected %v, got %v", tc.expected, mapping)
			}
			for runeIdx, glyphIdx := range tc.expected {
				g := mapRuneToClusterIndex(tc.dir, tc.runes, tc.glyphs, runeIdx)
				if g != glyphIdx {
					t.Errorf("map single, expected rune %d to yield %d, got %d", runeIdx, glyphIdx, g)
				}
			}
		})
	}
}

func TestInclusiveRange(t *testing.T) {
	type testcase struct {
		name string
		// inputs
		dir         di.Direction
		start       int
		breakAfter  int
		runeToGlyph []int
		numGlyphs   int
		// expected outputs
		gs, ge int
	}
	for _, tc := range []testcase{
		{
			name:        "simple at start",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          0,
			ge:          2,
		},
		{
			name:        "simple in middle",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          1,
			ge:          3,
		},
		{
			name:        "simple at end",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          2,
			ge:          4,
		},
		{
			name:        "simple at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          2,
			ge:          4,
		},
		{
			name:        "simple in middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          1,
			ge:          3,
		},
		{
			name:        "simple at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters at start",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       0,
			breakAfter:  1,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          0,
			ge:          1,
		},
		{
			name:        "fused clusters start and middle",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters middle and end",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          2,
			ge:          4,
		},
		{
			name:        "fused clusters at end",
			dir:         di.DirectionLTR,
			numGlyphs:   5,
			start:       3,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          3,
			ge:          4,
		},
		{
			name:        "fused clusters at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       0,
			breakAfter:  1,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          3,
			ge:          4,
		},
		{
			name:        "fused clusters start and middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          2,
			ge:          4,
		},
		{
			name:        "fused clusters middle and end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   5,
			start:       3,
			breakAfter:  4,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          0,
			ge:          1,
		},
		{
			name:        "ligatures at start",
			dir:         di.DirectionLTR,
			numGlyphs:   3,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          0,
			ge:          1,
		},
		{
			name:        "ligatures in middle",
			dir:         di.DirectionLTR,
			numGlyphs:   3,
			start:       2,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          1,
			ge:          1,
		},
		{
			name:        "ligatures at end",
			dir:         di.DirectionLTR,
			numGlyphs:   3,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          1,
			ge:          2,
		},
		{
			name:        "ligatures at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   3,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          1,
			ge:          2,
		},
		{
			name:        "ligatures in middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   3,
			start:       2,
			breakAfter:  2,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          1,
			ge:          1,
		},
		{
			name:        "ligatures at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   3,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          0,
			ge:          1,
		},
		{
			name:        "expansion at start",
			dir:         di.DirectionLTR,
			numGlyphs:   7,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          0,
			ge:          4,
		},
		{
			name:        "expansion in middle",
			dir:         di.DirectionLTR,
			numGlyphs:   7,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          1,
			ge:          5,
		},
		{
			name:        "expansion at end",
			dir:         di.DirectionLTR,
			numGlyphs:   7,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          4,
			ge:          6,
		},
		{
			name:        "expansion at start rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   7,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          2,
			ge:          6,
		},
		{
			name:        "expansion in middle rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   7,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          1,
			ge:          5,
		},
		{
			name:        "expansion at end rtl",
			dir:         di.DirectionRTL,
			numGlyphs:   7,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          0,
			ge:          2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gs, ge := inclusiveGlyphRange(tc.dir, tc.start, tc.breakAfter, tc.runeToGlyph, tc.numGlyphs)
			if gs != tc.gs {
				t.Errorf("glyphStart mismatch, got %d, expected %d", gs, tc.gs)
			}
			if ge != tc.ge {
				t.Errorf("glyphEnd mismatch, got %d, expected %d", ge, tc.ge)
			}
		})
	}
}

func withRange(output Output, runes Range) Output {
	output.Runes = runes
	return output
}

var (
	oneWord       = "aaaa"
	shapedOneWord = Output{
		Advance: fixed.I(10 * len([]rune(oneWord))),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: glyphs(0, len(oneWord)-1),
		Runes: Range{
			Count: len([]rune(oneWord)),
		},
	}
	// Assume the simple case of 1:1:1 glyph:rune:byte for this input.
	text1       = "text one is ltr"
	shapedText1 = Output{
		Advance: fixed.I(10 * len([]rune(text1))),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: glyphs(0, 14),
		Runes: Range{
			Count: len([]rune(text1)),
		},
	}
	text1Trailing       = text1 + " "
	shapedText1Trailing = func() Output {
		out := shapedText1
		out.Runes.Count++
		out.Glyphs = append(out.Glyphs, simpleGlyph(len(out.Glyphs)))
		out.RecalculateAll()
		return out
	}()
	// Test M:N:O glyph:rune:byte for this input.
	// The substring `lig` is shaped as a ligature.
	// The substring `DROP` is not shaped at all.
	text2       = "안П你 ligDROP 안П你 ligDROP"
	shapedText2 = Output{
		// There are 11 glyphs shaped for this string.
		Advance: fixed.I(10 * 11),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: []Glyph{
			0: simpleGlyph(0),      // 안        - 4 bytes
			1: simpleGlyph(1),      // П         - 3 bytes
			2: simpleGlyph(2),      // 你        - 4 bytes
			3: simpleGlyph(3),      // <space>   - 1 byte
			4: ligatureGlyph(4, 7), // lig       - 3 runes, 3 bytes
			// DROP                   - 4 runes, 4 bytes (included in ligature rune count)
			5:  simpleGlyph(11),      // <space> - 1 byte
			6:  simpleGlyph(12),      // 안      - 4 bytes
			7:  simpleGlyph(13),      // П       - 3 bytes
			8:  simpleGlyph(14),      // 你      - 4 bytes
			9:  simpleGlyph(15),      // <space> - 1 byte
			10: ligatureGlyph(16, 7), // lig     - 3 runes, 3 bytes
			// DROP                   - 4 runes, 4 bytes (included in ligature rune count)
		},
		Runes: Range{
			Count: len([]rune(text2)),
		},
	}
	// Test RTL languages.
	text3       = "שלום أهلا שלום أهلا"
	shapedText3 = Output{
		// There are 15 glyphs shaped for this string.
		Advance: fixed.I(10 * 15),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: []Glyph{
			0: ligatureGlyph(16, 3), // LIGATURE of three runes:
			//                         ا - 3 bytes
			//                         ل - 3 bytes
			//                         ه - 3 bytes
			1: simpleGlyph(15),     // أ - 3 bytes
			2: simpleGlyph(14),     // <space> - 1 byte
			3: simpleGlyph(13),     // ם - 3 bytes
			4: simpleGlyph(12),     // ו - 3 bytes
			5: simpleGlyph(11),     // ל - 3 bytes
			6: simpleGlyph(10),     // ש - 3 bytes
			7: simpleGlyph(9),      // <space> - 1 byte
			8: ligatureGlyph(6, 3), // LIGATURE of three runes:
			//                         ا - 3 bytes
			//                         ل - 3 bytes
			//                         ه - 3 bytes
			9:  simpleGlyph(5), //     أ - 3 bytes
			10: simpleGlyph(4), //     <space> - 1 byte
			11: simpleGlyph(3), //     ם - 3 bytes
			12: simpleGlyph(2), //     ו - 3 bytes
			13: simpleGlyph(1), //     ל - 3 bytes
			14: simpleGlyph(0), //     ש - 3 bytes
		},
		Direction: di.DirectionRTL,
		Runes: Range{
			Count: len([]rune(text3)),
		},
	}
	multiInputText1       = "aa aa aa"
	shapedMultiInputText1 = Output{
		Advance: fixed.I(10 * len([]rune(multiInputText1))),
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
		Glyphs: glyphs(0, len([]rune(multiInputText1))-1),
		Runes:  Range{Count: len([]rune(multiInputText1))},
	}
	splitShapedMultiInput1 = splitShapedAt(shapedMultiInputText1, 4, 6)

	bidiText1       = "hello أهلا שלום test"
	shapedBidiText1 = []Output{
		{
			// LTR initial segment
			Advance:   fixed.I(10 * len([]rune("hello "))),
			Direction: di.DirectionLTR,
			Runes: Range{
				Count: len([]rune("hello ")),
			},
			LineBounds: Bounds{
				Ascent:  fixed.I(10),
				Descent: fixed.I(5),
			},
			GlyphBounds: Bounds{
				Ascent: fixed.I(10),
			},
			Glyphs: glyphs(0, len([]rune("hello "))-1),
		},
		{
			// RTL middle segment
			Advance:   fixed.I(10 * len([]rune("أهلا שלום "))),
			Direction: di.DirectionRTL,
			Runes: Range{
				Offset: len([]rune("hello ")),
				Count:  len([]rune("أهلا שלום ")),
			},
			LineBounds: Bounds{
				Ascent:  fixed.I(10),
				Descent: fixed.I(5),
			},
			GlyphBounds: Bounds{
				Ascent: fixed.I(10),
			},
			Glyphs: glyphs(len([]rune("hello أهلا שלום "))-1, len([]rune("hello "))),
		},
		{
			// LTR final segment
			Advance:   fixed.I(10 * len([]rune("test"))),
			Direction: di.DirectionLTR,
			Runes: Range{
				Offset: len([]rune("hello أهلا שלום ")),
				Count:  len([]rune("test")),
			},
			LineBounds: Bounds{
				Ascent:  fixed.I(10),
				Descent: fixed.I(5),
			},
			GlyphBounds: Bounds{
				Ascent: fixed.I(10),
			},
			Glyphs: glyphs(len([]rune("hello أهلا שלום ")), len([]rune("hello أهلا שלום test"))-1),
		},
	}
)

// splitShapedAt splits a single shaped output into multiple. It splits
// on each provided glyph index in indices, with the index being the end of
// a slice range (so it's exclusive). You can think of the index as the
// first glyph of the next output.
func splitShapedAt(shaped Output, indices ...glyphIndex) []Output {
	numOut := len(indices) + 1
	outputs := make([]Output, 0, numOut)
	start := 0
	runeOffset := shaped.Runes.Offset
	for _, i := range indices {
		newOut := shaped
		newOut.Glyphs = newOut.Glyphs[start:i]
		newOut.Runes.Offset = runeOffset
		newOut.Runes.Count = 0
		cluster := -1
		for _, g := range newOut.Glyphs {
			if cluster == g.ClusterIndex {
				continue
			}
			cluster = g.ClusterIndex
			newOut.Runes.Count += g.RuneCount
		}
		runeOffset += newOut.Runes.Count
		newOut.RecalculateAll()
		outputs = append(outputs, newOut)
		start = i
	}
	newOut := shaped
	newOut.Glyphs = newOut.Glyphs[start:]
	newOut.Runes.Offset = runeOffset
	newOut.Runes.Count = shaped.Runes.Count + shaped.Runes.Offset - newOut.Runes.Offset
	newOut.RecalculateAll()
	outputs = append(outputs, newOut)
	return outputs
}

func TestWrapLine(t *testing.T) {
	type expected struct {
		line     Line
		nextLine int
		done     bool
	}
	type testcase struct {
		name      string
		shaped    []Output
		paragraph []rune
		maxWidth  int
		expected  []expected
	}
	for _, tc := range []testcase{
		{
			name:      "simple",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  40,
			expected: []expected{
				{
					line:     []Output{splitShapedAt(shapedText1, 5)[0]},
					nextLine: 5,
					done:     false,
				},
				{
					line:     []Output{splitShapedAt(shapedText1, 5, 9)[1]},
					nextLine: 9,
					done:     false,
				},
				{
					line:     []Output{splitShapedAt(shapedText1, 9, 12)[1]},
					nextLine: 12,
					done:     false,
				},
				{
					line:     []Output{splitShapedAt(shapedText1, 12)[1]},
					nextLine: 15,
					done:     true,
				},
			},
		},
		{
			// This test uses the same input text as the previous, but chops
			// every glyph into its own run. This simulates text that changes
			// font or style every glyph.
			name:      "simple in pieces 1",
			shaped:    splitShapedAt(shapedText1, 1, 2, 3, 4, 5),
			paragraph: []rune(text1),
			maxWidth:  40,
			expected: []expected{
				{
					line:     splitShapedAt(shapedText1, 1, 2, 3, 4, 5)[:5],
					nextLine: 5,
					done:     false,
				},
			},
		},
		{
			// This test uses the same test strategy as the previous, but divides
			// the run into segments that do not align evenly with line break
			// candidates. This forces the wrapper to break one run across lines.
			name:      "simple in pieces 2",
			shaped:    splitShapedAt(shapedText1, 3, 6),
			paragraph: []rune(text1),
			maxWidth:  40,
			expected: []expected{
				{
					line:     splitShapedAt(shapedText1, 3, 5)[:2],
					nextLine: 5,
					done:     false,
				},
				{
					line:     splitShapedAt(shapedText1, 5, 6, 9)[1:3],
					nextLine: 9,
					done:     false,
				},
			},
		},
		{
			name:      "simple rtl",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  40,
			expected: []expected{
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 10)[1],
							Range{Count: 5}),
					},
					nextLine: 5,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 7, 10)[1],
							Range{Offset: 5, Count: 5}),
					},
					nextLine: 10,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 2, 7)[1],
							Range{Offset: 10, Count: 5}),
					},
					nextLine: 15,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedText3, 2)[0],
							Range{Offset: 15, Count: 4}),
					},
					nextLine: 19,
					done:     true,
				},
			},
		},
		{
			name:      "simple bidi",
			shaped:    shapedBidiText1,
			paragraph: []rune(bidiText1),
			maxWidth:  110,
			expected: []expected{
				{
					line: []Output{
						shapedBidiText1[0],
						withRange(splitShapedAt(shapedBidiText1[1], 5)[1],
							Range{Offset: 6, Count: 5}),
					},
					nextLine: 11,
					done:     false,
				},
				{
					line: []Output{
						withRange(splitShapedAt(shapedBidiText1[1], 5)[0],
							Range{Offset: 11, Count: 5}),
						shapedBidiText1[2],
					},
					nextLine: 20,
					done:     true,
				},
			},
		},
		{
			// This is regression test. One of the benchmarks would hit this case in which
			// the line wrap candidate was exactly at the end of the available runs and
			// would increment the currentRun off of the end of the available runs.
			name:      "one word",
			shaped:    []Output{shapedOneWord},
			paragraph: []rune(oneWord),
			maxWidth:  10,
			expected: []expected{
				{
					line: []Output{
						shapedOneWord,
					},
					nextLine: 4,
					done:     true,
				},
			},
		},
		{
			name:      "simple fast-path",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  200,
			expected: []expected{
				{
					line:     Line{shapedText1},
					nextLine: 15,
					done:     true,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				line WrappedLine
				done bool
				l    LineWrapper
			)
			l.Prepare(WrapConfig{BreakPolicy: Never}, tc.paragraph, NewSliceIterator(tc.shaped))
			// Iterate every line declared in the test case expectations. This
			// allows test cases to be exhaustive if they need to wihtout forcing
			// every case to wrap entire paragraphs.
			for lineNumber, expected := range tc.expected {
				line, done = l.WrapNextLine(tc.maxWidth)
				compareLines(t, lineNumber, expected.line, line.Line)
				if done != expected.done {
					t.Errorf("done mismatch! expected %v, got %v", expected.done, done)
				}

				tu.AssertC(t, line.NextLine == expected.nextLine, fmt.Sprintf("expected %d, got %d", expected.nextLine, line.NextLine))

				if expected.done { // check WrapNextLine is now a no-op
					line, done = l.WrapNextLine(200)
					if line.Line != nil || !done {
						t.Errorf("expect nil output for WrapNextLine, got %p and %v", line.Line, done)
					}
				}
			}
		})
	}
}

func compareLines(t *testing.T, lineNumber int, expected, actual Line) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("line %d: expected %d runs, got %d", lineNumber, len(expected), len(actual))
		return
	}
	for i := range expected {
		expected := expected[i]
		actual := actual[i]
		if len(expected.Glyphs) != len(actual.Glyphs) {
			t.Errorf("line %d: run %d: expected %d glyphs, got %d", lineNumber, i, len(expected.Glyphs), len(actual.Glyphs))
			return
		}
		for ii := range expected.Glyphs {
			eg := expected.Glyphs[ii]
			ag := actual.Glyphs[ii]
			if eg != ag {
				t.Errorf("line: %d: run %d: glyph %d: expected\n%#+v\ngot\n%#+v", lineNumber, i, ii, eg, ag)
			}
		}
		expected.Glyphs = nil
		actual.Glyphs = nil
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("line %d: run %d: expected\n%#+v\ngot\n%#+v", lineNumber, i, expected, actual)
		}
	}
}

func TestLineWrap(t *testing.T) {
	emptyString := []Output{{
		Size:       fixed.I(100),
		LineBounds: Bounds{Ascent: fixed.I(-100), Descent: fixed.I(50)},
	}}
	type testcase struct {
		name      string
		shaped    []Output
		paragraph []rune
		maxWidth  int
		expected  []Line
	}
	for _, tc := range []testcase{
		{
			// Wrapping the empty string should produce a run with no glyphs/runes but
			// with valid line bounds.
			name:      "empty string",
			shaped:    emptyString,
			paragraph: []rune{},
			maxWidth:  1000,
			expected:  []Line{emptyString},
		},
		{
			// Wrapping the empty string should produce a run with no glyphs/runes but
			// with valid line bounds (even if the available width is zero).
			name:      "empty string zero width",
			shaped:    emptyString,
			paragraph: []rune{},
			maxWidth:  0,
			expected:  []Line{emptyString},
		},
		{
			// This test case verifies that no line breaks occur if they are not
			// necessary, and that the proper Offsets are reported in the output.
			name:      "all one line",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(shapedText1, Range{
						Count: len([]rune(text1)),
					}),
				},
			},
		},
		{
			// This test case verifies that trailing whitespace characters on a
			// line do not just disappear if it's the first line.
			name:      "trailing whitespace",
			shaped:    []Output{shapedText1Trailing},
			paragraph: []rune(text1Trailing),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(shapedText1Trailing, Range{
						Count: len([]rune(text1)) + 1,
					}),
				},
			},
		},
		{
			// This test case verifies that the line wrapper rejects line break
			// candidates that would split a glyph cluster.
			name: "reject mid-cluster line breaks",
			shaped: []Output{
				{
					Advance: fixed.I(10 * 3),
					LineBounds: Bounds{
						Ascent:  fixed.I(10),
						Descent: fixed.I(5),
						// No line gap.
					},
					GlyphBounds: Bounds{
						Ascent: fixed.I(10),
						// No glyphs descend.
					},
					Glyphs: []Glyph{
						simpleGlyph(0),
						complexGlyph(1, 2, 2),
						complexGlyph(1, 2, 2),
					},
					Runes: Range{Count: 3},
				},
			},
			// This unicode data was discovered in a testing/quick failure
			// for widget.Editor. It has the property that the middle two
			// runes form a harfbuzz cluster but also have a legal UAX#14
			// segment break between them.
			paragraph: []rune{0xa8e58, 0x3a4fd, 0x119dd},
			maxWidth:  20,
			expected: []Line{
				[]Output{
					withRange(
						Output{
							Direction: di.DirectionLTR,
							Advance:   fixed.I(10),
							LineBounds: Bounds{
								Ascent:  fixed.I(10),
								Descent: fixed.I(5),
							},
							GlyphBounds: Bounds{
								Ascent: fixed.I(10),
							},
							Glyphs: []Glyph{
								simpleGlyph(0),
							},
						},
						Range{
							Count: 1,
						},
					),
				},
				[]Output{
					withRange(
						Output{
							Direction: di.DirectionLTR,
							Advance:   fixed.I(20),
							LineBounds: Bounds{
								Ascent:  fixed.I(10),
								Descent: fixed.I(5),
							},
							GlyphBounds: Bounds{
								Ascent: fixed.I(10),
							},
							Glyphs: []Glyph{
								complexGlyph(1, 2, 2),
								complexGlyph(1, 2, 2),
							},
						},
						Range{
							Count:  2,
							Offset: 1,
						},
					),
				},
			},
		},
		{
			// This test case verifies that the line wrapper rejects line break
			// candidates that would split a glyph cluster at non-zero offsets
			// within the shaped text.
			name: "reject mid-cluster line breaks at non-zero offsets",
			shaped: []Output{
				{
					Advance:     fixed.I(10),
					LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{Ascent: fixed.I(10)},
					Glyphs: []Glyph{
						simpleGlyph(0),
					},
					Runes: Range{Count: 1},
				},
				{
					Advance:     fixed.I(10),
					LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{Ascent: fixed.I(10)},
					Glyphs: []Glyph{
						simpleGlyph(1),
					},
					Runes: Range{Count: 1, Offset: 1},
				},
				{
					Advance:     fixed.I(10 * 2),
					LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{Ascent: fixed.I(10)},
					Glyphs: []Glyph{
						complexGlyph(2, 2, 2),
						complexGlyph(2, 2, 2),
					},
					Runes: Range{Count: 2, Offset: 2},
				},
				{
					Advance:    fixed.I(10),
					LineBounds: Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
					GlyphBounds: Bounds{
						Ascent: fixed.I(10),
					},
					Glyphs: []Glyph{
						simpleGlyph(4),
					},
					Runes: Range{Count: 1, Offset: 4},
				},
			},
			// This unicode data was discovered in a fuzz test failure
			// for Gio's text shaper.
			paragraph: []rune{1593, 48, 32, 1474, 48},
			maxWidth:  40,
			expected: []Line{
				[]Output{
					{
						Advance:     fixed.I(10),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							simpleGlyph(0),
						},
						Runes: Range{Count: 1},
					},
					{
						Advance:     fixed.I(10),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							simpleGlyph(1),
						},
						Runes: Range{Count: 1, Offset: 1},
					},
					{
						Advance:     fixed.I(10 * 2),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							complexGlyph(2, 2, 2),
							complexGlyph(2, 2, 2),
						},
						Runes: Range{Count: 2, Offset: 2},
					},
					{
						Advance:     fixed.I(10),
						LineBounds:  Bounds{Ascent: fixed.I(10), Descent: fixed.I(5)},
						GlyphBounds: Bounds{Ascent: fixed.I(10)},
						Glyphs: []Glyph{
							simpleGlyph(4),
						},
						Runes: Range{Count: 1, Offset: 4},
					},
				},
			},
		},
		{
			// This test case verifies that line breaking does occur, and that
			// all lines have proper offsets.
			name:      "line break on last word",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  120,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 12)[0],
						Range{
							Count: len([]rune(text1)) - 3,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 12)[1],
						Range{
							Offset: len([]rune(text1)) - 3,
							Count:  3,
						},
					),
				},
			},
		},
		{
			// This test case verifies that many line breaks still result in
			// correct offsets. This test also ensures that leading whitespace
			// is correctly hidden on lines after the first.
			name:      "line break several times",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  70,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 5)[0],
						Range{
							Count: 5,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 5, 12)[1],
						Range{
							Offset: 5,
							Count:  7,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, 12)[1],
						Range{
							Offset: 12,
							Count:  3,
						},
					),
				},
			},
		},
		{
			// This test case verifies baseline offset math for more complicated input.
			name:      "all one line 2",
			shaped:    []Output{shapedText2},
			paragraph: []rune(text2),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(
						shapedText2,
						Range{
							Count: len([]rune(text2)),
						},
					),
				},
			},
		},
		{
			// This test case verifies that offset accounting correctly handles complex
			// input across line breaks. It is legal to line-break within words composed
			// of more than one script, so this test expects that to occur.
			name:      "line break several times 2",
			shaped:    []Output{shapedText2},
			paragraph: []rune(text2),
			maxWidth:  40,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText2, 4)[0],
						Range{
							Count: len([]rune("안П你 ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText2, 4, 8)[1],
						Range{
							Count:  len([]rune("ligDROP 안П")),
							Offset: len([]rune("안П你 ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText2, 8, 11)[1],
						Range{
							Count:  len([]rune("你 ligDROP")),
							Offset: len([]rune("안П你 ligDROP 안П")),
						},
					),
				},
			},
		},
		{
			// This test case verifies baseline offset math for complex RTL input.
			name:      "all one line 3",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  1000,
			expected: []Line{
				[]Output{
					withRange(
						shapedText3,
						Range{
							Count: len([]rune(text3)),
						},
					),
				},
			},
		},
		{
			// This test case verifies line wrapping logic in RTL mode.
			name:      "line break once [RTL]",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  100,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 7)[1],
						Range{
							Count: len([]rune("שלום أهلا ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 7)[0],
						Range{
							Count:  len([]rune("שלום أهلا")),
							Offset: len([]rune("שלום أهلا ")),
						},
					),
				},
			},
		},
		{
			// This test case verifies line wrapping logic in RTL mode.
			name:      "line break several times [RTL]",
			shaped:    []Output{shapedText3},
			paragraph: []rune(text3),
			maxWidth:  50,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 10)[1],
						Range{
							Count: len([]rune("שלום ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 7, 10)[1],
						Range{
							Count:  len([]rune("أهلا ")),
							Offset: len([]rune("שלום ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 2, 7)[1],
						Range{
							Count:  len([]rune("שלום ")),
							Offset: len([]rune("שלום أهلا ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, 2)[0],
						Range{
							Count:  len([]rune("أهلا")),
							Offset: len([]rune("שלום أهلا שלום ")),
						},
					),
				},
			},
		},
		{
			// This test case verifies the behavior of the line wrapper for multi-run
			// shaped input.
			name:      "multiple input runs 1",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  20,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:3],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 2",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  30,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:3],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 3",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  40,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:3],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 4",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  50,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 3)[:1],
				splitShapedAt(shapedMultiInputText1, 3, 4, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 5",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  60,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 4, 6)[:2],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 6",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  70,
			expected: []Line{
				splitShapedAt(shapedMultiInputText1, 4, 6)[:2],
				splitShapedAt(shapedMultiInputText1, 6)[1:],
			},
		},
		{
			name:      "multiple input runs 7",
			shaped:    splitShapedMultiInput1,
			paragraph: []rune(multiInputText1),
			maxWidth:  80,
			expected: []Line{
				splitShapedMultiInput1,
			},
		},
		{
			// This is regression test. One of the benchmarks would hit this case in which
			// the line wrap candidate was exactly at the end of the available runs and
			// would increment the currentRun off of the end of the available runs.
			name:      "one word",
			shaped:    []Output{shapedOneWord},
			paragraph: []rune(oneWord),
			maxWidth:  10,
			expected: []Line{
				{
					shapedOneWord,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var l LineWrapper
			outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, tc.maxWidth, tc.paragraph, NewSliceIterator(tc.shaped))

			if len(tc.expected) != len(outs) {
				t.Errorf("expected %d lines, got %d", len(tc.expected), len(outs))
				return
			}
			for lineNum := range tc.expected {
				expectedLine := tc.expected[lineNum]
				actualLine := outs[lineNum]
				if len(expectedLine) != len(actualLine) {
					t.Errorf("expected %d runs in line %d, got %d", len(expectedLine), lineNum, len(actualLine))
					continue
				}
				for runNum := range expectedLine {
					expectedRun := expectedLine[runNum]
					actualRun := actualLine[runNum]
					lenE := len(expectedRun.Glyphs)
					lenO := len(actualRun.Glyphs)
					if lenE != lenO {
						t.Errorf("line %d run %d: expected %d glyphs, got %d", lineNum, runNum, lenE, lenO)
						continue
					}
					for k := range expectedRun.Glyphs {
						e := expectedRun.Glyphs[k]
						o := actualRun.Glyphs[k]
						if !reflect.DeepEqual(e, o) {
							t.Errorf("line %d: glyph mismatch at index %d, expected: %#v, got %#v", runNum, k, e, o)
						}
					}
					if expectedRun.Runes != actualRun.Runes {
						t.Errorf("line %d: expected %#v offsets, got %#v", runNum, expectedRun.Runes, actualRun.Runes)
					}
					if expectedRun.Direction != actualRun.Direction {
						t.Errorf("line %d: expected %v direction, got %v", runNum, expectedRun.Direction, actualRun.Direction)
					}
					// Reduce the verbosity of the reflect mismatch since we already
					// compared the glyphs.
					expectedRun.Glyphs = nil
					actualRun.Glyphs = nil
					if !reflect.DeepEqual(expectedRun, actualRun) {
						t.Errorf("line %d: expected: %#v, got %#v", runNum, expectedRun, actualRun)
					}
				}
			}
		})
	}
}

// simpleGlyph returns a simple square glyph with the provided cluster
// value.
func simpleGlyph(cluster int) Glyph {
	return complexGlyph(cluster, 1, 1)
}

// ligatureGlyph returns a simple square glyph with the provided cluster
// value and number of runes.
func ligatureGlyph(cluster, runes int) Glyph {
	return complexGlyph(cluster, runes, 1)
}

// expansionGlyph returns a simple square glyph with the provided cluster
// value and number of glyphs.
func expansionGlyph(cluster, glyphs int) Glyph {
	return complexGlyph(cluster, 1, glyphs)
}

// complexGlyph returns a simple square glyph with the provided cluster
// value, number of associated runes, and number of glyphs in the cluster.
func complexGlyph(cluster, runes, glyphs int) Glyph {
	return Glyph{
		Width:        fixed.I(10),
		Height:       fixed.I(10),
		XAdvance:     fixed.I(10),
		YAdvance:     fixed.I(10),
		YBearing:     fixed.I(10),
		ClusterIndex: cluster,
		GlyphCount:   glyphs,
		RuneCount:    runes,
	}
}

func checkOptions(t *testing.T, runes []rune, options []breakOption) bool {
	t.Helper()
	// Ensure breaks are in valid range.
	for _, o := range options {
		if o.breakAtRune < 0 || o.breakAtRune > len(runes)-1 {
			t.Errorf("breakAtRune out of bounds: %d when len(runes)=%d", o.breakAtRune, len(runes))
			return false
		}
	}
	// Ensure breaks are sorted.
	if !sort.SliceIsSorted(options, func(i, j int) bool {
		return options[i].breakAtRune < options[j].breakAtRune
	}) {
		t.Errorf("breaks are not sorted: %#+v", options)
		return false
	}

	// Ensure breaks are unique.
	m := make([]bool, len(runes))
	for _, o := range options {
		if m[o.breakAtRune] {
			t.Errorf("breaks are not unique: %v is repeated in %#+v", o, m)
			return false
		} else {
			m[o.breakAtRune] = true
		}
	}

	return true
}

func TestRawBreakOptions(t *testing.T) {
	if err := quick.Check(func(runes []rune) bool {
		breaker := newBreaker(&segmenter.Segmenter{}, runes)
		var wordOptions []breakOption
		for b, ok := breaker.nextWordRaw(); ok; b, ok = breaker.nextWordRaw() {
			wordOptions = append(wordOptions, b)
		}
		var graphemeOptions []breakOption
		for b, ok := breaker.nextGraphemeRaw(); ok; b, ok = breaker.nextGraphemeRaw() {
			graphemeOptions = append(graphemeOptions, b)
		}
		return checkOptions(t, runes, wordOptions) && checkOptions(t, runes, graphemeOptions)
	}, nil); err != nil {
		t.Errorf("generated invalid break options: %v", err)
	}
}

// TestWrappingLatinE2E actually performs both text shaping and line wrapping
// on a selection of latin text.
func TestWrappingLatinE2E(t *testing.T) {
	textInput := []rune("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	face := benchEnFace
	var shaper HarfbuzzShaper
	out := []Output{shaper.Shape(Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})}
	var l LineWrapper
	outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, NewSliceIterator(out))
	if len(outs) < 3 {
		t.Errorf("expected %d lines, got %d", 3, len(outs))
	}
}

// TestWrappingBidiRegression checks a specific regression discovered within the Gio test suite.
func TestWrappingBidiRegression(t *testing.T) {
	type testcase struct {
		name     string
		text     []rune
		inputs   []Output
		maxWidth int
	}
	ltrText := []rune("The quick brown fox jumps over the lazy dog.")
	ltrTextRuns := []Input{
		{RunStart: 0, RunEnd: len(ltrText), Script: language.Latin, Direction: di.DirectionLTR},
	}
	bidiText := []rune("الحب سماء brown привет fox تمط jumps привет over غير الأحلام")
	bidiTextRuns := []Input{
		{RunStart: 0, RunEnd: 10, Script: language.Arabic, Direction: di.DirectionRTL},
		{RunStart: 10, RunEnd: 16, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 16, RunEnd: 23, Script: language.Cyrillic, Direction: di.DirectionLTR},
		{RunStart: 23, RunEnd: 26, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 26, RunEnd: 31, Script: language.Arabic, Direction: di.DirectionRTL},
		{RunStart: 31, RunEnd: 37, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 37, RunEnd: 44, Script: language.Cyrillic, Direction: di.DirectionLTR},
		{RunStart: 44, RunEnd: 48, Script: language.Latin, Direction: di.DirectionLTR},
		{RunStart: 48, RunEnd: 60, Script: language.Arabic, Direction: di.DirectionRTL},
	}
	bidiText2 := []rune("د عرمثال dstي met لم aqل جدmوpمg lرe dرd  لو عل ميrةsdiduntut lab renنيتذدagلaaiua.ئPocttأior رادرsاي mيrbلmnonaيdتد ماةعcلخ.")
	bidiTextRuns2 := []Input{
		{Text: bidiText2, RunStart: 0, RunEnd: 9, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 9, RunEnd: 12, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 12, RunEnd: 14, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 14, RunEnd: 17, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 17, RunEnd: 21, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 21, RunEnd: 23, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 23, RunEnd: 27, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 27, RunEnd: 28, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 28, RunEnd: 29, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 29, RunEnd: 30, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 30, RunEnd: 31, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 31, RunEnd: 34, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 34, RunEnd: 35, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 35, RunEnd: 38, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 38, RunEnd: 39, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 39, RunEnd: 40, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 40, RunEnd: 50, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 50, RunEnd: 51, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 51, RunEnd: 52, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 52, RunEnd: 69, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 69, RunEnd: 74, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 74, RunEnd: 76, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 76, RunEnd: 77, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 77, RunEnd: 82, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 82, RunEnd: 84, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 84, RunEnd: 89, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 89, RunEnd: 90, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 90, RunEnd: 93, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 93, RunEnd: 98, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 98, RunEnd: 99, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 99, RunEnd: 102, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 102, RunEnd: 103, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 103, RunEnd: 104, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 104, RunEnd: 106, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 106, RunEnd: 107, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 107, RunEnd: 112, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 112, RunEnd: 113, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 113, RunEnd: 114, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 114, RunEnd: 121, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 121, RunEnd: 122, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 122, RunEnd: 125, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
	}
	truncator := (&HarfbuzzShaper{}).Shape(Input{
		Text:      []rune("…"),
		RunStart:  0,
		RunEnd:    len([]rune("…")),
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})
	shapeInputs := func(inputs []Input) []Output {
		var shaper HarfbuzzShaper
		out := make([]Output, len(inputs))
		for i, input := range inputs {
			out[i] = shaper.Shape(input)
		}
		return out
	}
	applyDefaultsAndShape := func(textInput []rune, runs []Input) []Output {
		enFace := benchEnFace
		arFace := benchArFace

		ppem := fixed.I(16)
		arLang := language.NewLanguage("AR")

		for i := range runs {
			runs[i].Text = textInput
			runs[i].Size = ppem
			if runs[i].Direction == di.DirectionRTL {
				runs[i].Face = arFace
			} else {
				runs[i].Face = enFace
			}
			// Even though the text sample is mixed, the overall document language is arabic.
			runs[i].Language = arLang
		}
		return shapeInputs(runs)
	}
	for _, tc := range []testcase{
		{
			name:     "simple ltr",
			text:     ltrText,
			inputs:   applyDefaultsAndShape(ltrText, ltrTextRuns),
			maxWidth: 100,
		},
		{
			name:     "many-run bidi",
			text:     bidiText,
			inputs:   applyDefaultsAndShape(bidiText, bidiTextRuns),
			maxWidth: 100,
		},
		{
			name:     "complex bidi",
			text:     bidiText2,
			inputs:   shapeInputs(bidiTextRuns2),
			maxWidth: 200,
		},
		{
			// This test case was observed failing Gio's text shaping fuzzer. It causes our line
			// wrapper to drop an entire run of text for some reason.
			name:     "bidi regression",
			text:     []rune("000000000000000 00000000 ٰ00000"),
			inputs:   regressionRuns,
			maxWidth: 121,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, policy := range []LineBreakPolicy{Always, Never, WhenNecessary} {
				for _, truncation := range []bool{true, false} {
					t.Run(fmt.Sprintf("graphemeBreak=%s_truncate=%v", policy, truncation), func(t *testing.T) {
						var l LineWrapper
						truncateAfter := 0
						if truncation {
							truncateAfter = 1
						}
						lines, truncated := l.WrapParagraph(WrapConfig{
							BreakPolicy:        policy,
							TruncateAfterLines: truncateAfter,
							Truncator:          truncator,
						}, tc.maxWidth, tc.text, NewSliceIterator(tc.inputs))
						if truncated != 0 && !truncation {
							t.Errorf("did not expect truncation, got truncated=%d", truncated)
						}
						checkRuneCounts(t, tc.text, lines, truncated)
					})
				}
			}
		})
	}
}

func checkRuneCounts(t *testing.T, source []rune, lines []Line, truncated int) []int {
	t.Helper()
	counts := []int{}
	totalRunes := 0
	alreadyFailed := t.Failed()
	for lineIdx, line := range lines {
		lineTotalRunes := 0
		for runIdx, run := range line {
			if truncated > 0 && totalRunes == len(source)-truncated && run.Runes.Offset == 0 {
				// Skip the truncator run.
				continue
			}
			if run.Runes.Offset != totalRunes {
				if !alreadyFailed {
					for i := 0; i <= lineIdx; i++ {
						line := lines[i]
						for k := 0; (i < lineIdx && k < len(line)) || (i == lineIdx && k < runIdx); k++ {
							run := line[k]
							t.Errorf("lines[%d][%d].Runes.Offset=%d, Count=%d", i, k, run.Runes.Offset, run.Runes.Count)
						}
					}
					alreadyFailed = true
				}
				t.Errorf("lines[%d][%d].Runes.Offset=%d, Count=%d, expected Offset %d", lineIdx, runIdx, run.Runes.Offset, run.Runes.Count, totalRunes)
			}
			totalRunes += run.Runes.Count
			lineTotalRunes += run.Runes.Count
		}
		counts = append(counts, lineTotalRunes)
	}
	if len(source)-truncated != totalRunes {
		t.Errorf("expected %d runes total, got %d", len(source)-truncated, totalRunes)
	}
	return counts
}

// TestWrappingTruncation checks that the line wrapper's truncation features
// behave as expected.
func TestWrappingTruncation(t *testing.T) {
	textInput := []rune("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	face := benchEnFace
	var shaper HarfbuzzShaper
	out := []Output{shaper.Shape(Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})}
	var l LineWrapper
	outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, NewSliceIterator(out))
	untruncatedCount := len(outs)

	for _, truncator := range []Output{
		{}, // No truncator.
		shaper.Shape(Input{ // Multi-rune truncator.
			Text:      []rune("..."),
			RunStart:  0,
			RunEnd:    len([]rune("...")),
			Direction: di.DirectionLTR,
			Face:      face,
			Size:      fixed.I(16),
			Script:    language.Latin,
			Language:  language.NewLanguage("EN"),
		}),
		shaper.Shape(Input{ // Single-rune truncator.
			Text:      []rune("…"),
			RunStart:  0,
			RunEnd:    len([]rune("…")),
			Direction: di.DirectionLTR,
			Face:      face,
			Size:      fixed.I(16),
			Script:    language.Latin,
			Language:  language.NewLanguage("EN"),
		}),
	} {
		for i := untruncatedCount + 1; i > 0; i-- {
			wc := WrapConfig{
				BreakPolicy:        Never,
				TruncateAfterLines: i,
				Truncator:          truncator,
			}
			newLines, truncated := l.WrapParagraph(wc, 250, textInput, NewSliceIterator(out))
			lineCount := len(newLines)
			t.Logf("wrapping with max lines=%d, untruncatedCount=%d", i, untruncatedCount)
			if i < untruncatedCount {
				if lineCount != i {
					t.Errorf("expected %d lines, got %d", i, lineCount)
				}
				if truncated < 1 {
					t.Errorf("expected lines to indicate truncation")
				}
				lastLine := newLines[len(newLines)-1]
				lastRun := lastLine[len(lastLine)-1]
				if !reflect.DeepEqual(lastRun, truncator) {
					t.Errorf("expected truncator as last run")
				}
			} else if i >= untruncatedCount {
				if lineCount != untruncatedCount {
					t.Errorf("expected %d lines, got %d", untruncatedCount, lineCount)
				}
				if truncated > 0 {
					t.Errorf("expected lines to indicate no truncation")
				}
			}
			runeCount := 0
			for _, line := range newLines {
				for _, run := range line {
					runeCount += run.Runes.Count
				}
			}
			// Remove the runes of the truncator, if any.
			if truncated > 0 {
				runeCount -= truncator.Runes.Count
			}
			if runeCount+truncated != len(textInput) {
				t.Errorf("expected %d runes total, got %d output and %d truncated", len(textInput), runeCount, truncated)
			}
		}
	}
}

func TestWrapping_oneLine(t *testing.T) {
	textInput := []rune("Lorem ipsum") // a simple input that fits on one line
	face := benchEnFace
	var shaper HarfbuzzShaper
	out := []Output{shaper.Shape(Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})}
	iter := NewSliceIterator(out)
	var l LineWrapper

	outs, _ := l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, iter)
	if len(outs) != 1 {
		t.Errorf("expected one line, got %d", len(outs))
	}

	// the run in iter should have been consumed
	outs, _ = l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 250, textInput, iter)
	if len(outs) != 0 {
		t.Errorf("expected no line, got %d", len(outs))
	}
}

// TestWrappingTruncation checks that the line wrapper's truncation features
// handle some edge cases.
func TestWrappingTruncationEdgeCases(t *testing.T) {
	type testcase struct {
		// name describing what is being tested.
		name string
		// input string to shape.
		input string
		// width to wrap shaped text to.
		wrapWidth int
		// cutInto is a count of how many runs the shaped input should be cut into
		// before wrapping, in order to ensure consistent wrapping behavior regardless
		// of the number of runs involved.
		cutInto int
		// maxLines controls how many lines of text will be wrapped before attempting
		// truncation.
		maxLines int
		// truncator to shape and use as the final run on truncated lines.
		truncator string
		// forceTruncation ensures that the final line in maxLines will show the truncator
		// symbol.
		forceTruncation bool
		// expectedTruncated is the expected count of truncated runes.
		expectedTruncated int
	}
	for _, tc := range []testcase{
		{
			name:              "only run doesn't fit 1 part",
			input:             "mmmmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 5,
		},
		{
			name:              "only run doesn't fit 5 parts",
			input:             "mmmmm",
			wrapWidth:         40,
			cutInto:           5,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 5,
		},
		{
			name:              "run only fits without truncator 1 part",
			input:             "mmmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "run only fits without truncator 1 part, forced to truncate",
			input:             "mmmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			forceTruncation:   true,
			expectedTruncated: 4,
		},
		{
			name:              "run only fits without truncator 4 part",
			input:             "mmmm",
			wrapWidth:         40,
			cutInto:           4,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run only fits without truncator 1 part",
			input:             "m mm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run only fits without truncator 1 part, forced to truncate",
			input:             "m mm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			forceTruncation:   true,
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run only fits without truncator 4 part",
			input:             "m mm",
			wrapWidth:         40,
			cutInto:           4,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 0,
		},
		{
			name:              "multi-word run doesn't fit 1 part",
			input:             "mmm mm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 2,
		},
		{
			name:              "multi-word run doesn't fit 4 part",
			input:             "mmm mm",
			wrapWidth:         40,
			cutInto:           4,
			maxLines:          1,
			truncator:         "...",
			expectedTruncated: 2,
		},
		{
			name:              "only run doesn't fit and truncator doesn't fit",
			input:             "mm mmm",
			wrapWidth:         40,
			cutInto:           1,
			maxLines:          1,
			truncator:         "mmmmm",
			expectedTruncated: 6,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputRunes := []rune(tc.input)
			truncRunes := []rune(tc.truncator)
			var shaper HarfbuzzShaper
			trunc := shaper.Shape(Input{
				Text:      truncRunes,
				RunStart:  0,
				RunEnd:    len(truncRunes),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(10),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			out := shaper.Shape(Input{
				Text:      inputRunes,
				RunStart:  0,
				RunEnd:    len(inputRunes),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(10),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			var outs []Output
			if tc.cutInto > 1 {
				outs = cutRunInto(out, tc.cutInto)
			} else {
				outs = append(outs, out)
			}
			var l LineWrapper
			lines, truncatedRunes := l.WrapParagraph(WrapConfig{
				BreakPolicy:        Never,
				Truncator:          trunc,
				TruncateAfterLines: tc.maxLines,
				TextContinues:      tc.forceTruncation,
			}, tc.wrapWidth, inputRunes, NewSliceIterator(outs))
			if truncatedRunes != tc.expectedTruncated {
				t.Errorf("got %d truncated runes when truncation expectation was %d", truncatedRunes, tc.expectedTruncated)
			}
			lastLine := lines[len(lines)-1]
			lastRun := lastLine[len(lastLine)-1]
			shouldTruncate := (tc.expectedTruncated > 0 || tc.forceTruncation)
			if lastRunIsTruncator := reflect.DeepEqual(lastRun, trunc); lastRunIsTruncator != shouldTruncate {
				t.Errorf("shouldTruncate = %v, but lastRunIsTruncator = %v", shouldTruncate, lastRunIsTruncator)
			}
		})
	}
}

// TestTruncationWithBreaking tests cases of interest involving both text truncation and
// line breaking policies.
func TestTruncationWithBreaking(t *testing.T) {
	inputText := []rune("Fortunately, there's another way to build streams that is nearly API-compatible with the current approach. You can model them as computational Directed Acyclic Graphs and build a structure of nodes and edges that process values. Nodes essentially run a function on their input edges and propagate the results to their output edges. Edges act as the synchronization primitive in this scheme, ensuring that nodes are able to execute safely in parallel with one another.")
	inputRun := Input{
		Text:      inputText,
		RunStart:  0,
		RunEnd:    len(inputText),
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}
	truncatorRun := Input{
		Text:      []rune("…"),
		RunStart:  0,
		RunEnd:    1,
		Direction: di.DirectionLTR,
		Face:      benchEnFace,
		Size:      fixed.I(16),
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	}
	var shaper HarfbuzzShaper
	output := shaper.Shape(inputRun)
	truncator := shaper.Shape(truncatorRun)
	type testcase struct {
		name      string
		width     int
		truncated int
		policy    LineBreakPolicy
	}

	for _, tc := range []testcase{
		{
			name:      "Never",
			width:     936,
			policy:    Never,
			truncated: 341,
		},
		{
			name:      "WhenNecessary",
			width:     936,
			policy:    WhenNecessary,
			truncated: 341,
		},
		{
			name:      "Always",
			width:     936,
			policy:    Always,
			truncated: 341,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var wrapper LineWrapper
			lines, truncated := wrapper.WrapParagraph(WrapConfig{
				BreakPolicy:        tc.policy,
				TruncateAfterLines: 1,
				Truncator:          truncator,
			}, tc.width, inputText, NewSliceIterator([]Output{output.copy()}))
			if truncated != tc.truncated {
				t.Errorf("expected %d truncated runes, got %d (total %d)", tc.truncated, truncated, len(inputText))
			}
			checkRuneCounts(t, inputText, lines, truncated)
		})
	}
}

func TestGraphemeBreakingRegression(t *testing.T) {
	bidiText2 := []rune("renنيتذدagلaaiua.ئPocttأior رادرs")
	inputRuns := []Input{
		{Text: bidiText2, RunStart: 0, RunEnd: 3, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 3, RunEnd: 8, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 8, RunEnd: 10, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 10, RunEnd: 11, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 11, RunEnd: 16, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 16, RunEnd: 18, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 18, RunEnd: 23, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 23, RunEnd: 24, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 24, RunEnd: 27, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
		{Text: bidiText2, RunStart: 27, RunEnd: 32, Direction: 0x1, Face: benchEnFace, Size: 896, Script: language.Arabic, Language: "en"},
		{Text: bidiText2, RunStart: 32, RunEnd: 33, Direction: 0x0, Face: benchEnFace, Size: 896, Script: language.Latin, Language: "en"},
	}
	var shaper HarfbuzzShaper
	shaped := []Output{}
	for _, run := range inputRuns {
		shaped = append(shaped, shaper.Shape(run))
	}
	var wrapper LineWrapper
	maxWidth := 0
	for _, run := range shaped {
		maxWidth += run.Advance.Ceil()
	}
	// Wrap at increasingly unreasonable widths, checking for sane wrapping decisions.
	for maxWidth := maxWidth; maxWidth > 0; maxWidth-- {
		t.Run(fmt.Sprintf("maxWidth=%d", maxWidth), func(t *testing.T) {
			runs := make([]Output, len(shaped))
			for i, run := range shaped {
				runs[i] = run.copy()
			}
			lines, truncated := wrapper.WrapParagraph(WrapConfig{BreakPolicy: Always}, maxWidth, bidiText2, NewSliceIterator(runs))
			checkRuneCounts(t, bidiText2, lines, truncated)

			for i, baseLine := range lines[:len(lines)-1] {
				nextLine := lines[i+1]
				baseAdv := fixed.Int26_6(0)
				for _, run := range baseLine {
					baseAdv += run.Advance
				}
				nextAdv := fixed.Int26_6(0)
				for _, run := range nextLine {
					nextAdv += run.Advance
				}
				if total := (baseAdv + nextAdv).Ceil(); total <= maxWidth {
					t.Errorf("lines[%d] and lines[%d] could have fit on the same line (total width %d)", i, i+1, total)
				}
			}

			if maxWidth == 1 {
				// Check that every grapheme cluster was wrapped properly to its own line.
				seg := segmenter.Segmenter{}
				seg.Init(bidiText2)
				iter := seg.GraphemeIterator()
				lengths := []int{}
				for iter.Next() {
					cluster := iter.Grapheme()
					lengths = append(lengths, len(cluster.Text))
				}
				if len(lengths) != len(lines) {
					t.Errorf("got %d lines but there are %d grapheme clusters", len(lines), len(lengths))
				}
				for lineNum, line := range lines {
					totalRunes := 0
					for _, run := range line {
						totalRunes += run.Runes.Count
					}
					if totalRunes != lengths[lineNum] {
						t.Errorf("lines[%d] has %d runes, but grapheme cluster at index %d has %d", lineNum, totalRunes, lineNum, lengths[lineNum])
					}
				}
			}
		})
	}
}

func BenchmarkMapping(b *testing.B) {
	type wrapfunc func(di.Direction, Range, []Glyph, []glyphIndex) []glyphIndex
	for _, langInfo := range benchLangs {
		for _, size := range []int{10, 100, 1000} {
			for impl, f := range map[string]wrapfunc{
				//"v1": mapRunesToClusterIndices,
				//"v2": mapRunesToClusterIndices2,
				"v3": mapRunesToClusterIndices3,
			} {
				b.Run(fmt.Sprintf("%drunes-%s-%s", size, langInfo.name, impl), func(b *testing.B) {
					var shaper HarfbuzzShaper
					out := shaper.Shape(Input{
						Text:      langInfo.text[:size],
						RunStart:  0,
						RunEnd:    size,
						Direction: langInfo.dir,
						Face:      langInfo.face,
						Size:      16 * 72,
						Script:    langInfo.script,
						Language:  langInfo.lang,
					})
					var m []glyphIndex
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						m = f(out.Direction, out.Runes, out.Glyphs, m)
					}
					_ = m
				})
			}
		}
	}
}

// benchLangInfo describes the language configuration for a text shaping input.
type benchLangInfo struct {
	name   string
	dir    di.Direction
	script language.Script
	lang   language.Language
	face   *font.Face
	text   []rune
}

// benchSizeConfig describes a size of benchmarking input and a set of configurations
// for cutting the runes into quantities of equal-sized parts.
type benchSizeConfig struct {
	runes int
	parts []int
}

// benchArFace is an arabic font face for use in benchmarks.
var benchArFace = func() *font.Face {
	data, err := os.ReadFile("../font/testdata/Amiri-Regular.ttf")
	if err != nil {
		panic(err)
	}
	arFace, err := font.ParseTTF(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return arFace
}()

// benchEnFace is a latin font face for use in benchmarks.
var benchEnFace = func() *font.Face {
	enFace, err := font.ParseTTF(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	return enFace
}()

var benchLangs = []benchLangInfo{
	{
		name:   "arabic",
		dir:    di.DirectionRTL,
		script: language.Arabic,
		lang:   language.NewLanguage("AR"),
		face:   benchArFace,
		text:   []rune(benchParagraphArabic),
	},
	{
		name:   "latin",
		dir:    di.DirectionLTR,
		script: language.Latin,
		lang:   language.NewLanguage("EN"),
		face:   benchEnFace,
		text:   []rune(benchParagraphLatin),
	},
}

var benchSizes = []benchSizeConfig{
	{runes: 10, parts: []int{1, 10}},
	{runes: 100, parts: []int{1, 10, 100}},
	{runes: 1000, parts: []int{1, 10, 100, 1000}},
}

// cutRunInto divides the run into [parts] of same size (with the last part absorbing any remainder).
func cutRunInto(run Output, parts int) []Output {
	var outs []Output
	mapping := mapRunesToClusterIndices3(run.Direction, run.Runes, run.Glyphs, nil)
	runesPerPart := run.Runes.Count / parts
	partStart := 0
	for i := 0; i < parts-1; i++ {
		outs = append(outs, cutRun(run, mapping, partStart, partStart+runesPerPart-1, false))
		partStart += runesPerPart
	}
	outs = append(outs, cutRun(run, mapping, partStart, run.Runes.Count-1, false))
	return outs
}

// TestCutRunInto ensures that the cutRunInto helper function actually cuts the run into the
// right pieces.
func TestCutRunInto(t *testing.T) {
	for _, langInfo := range benchLangs {
		for _, size := range benchSizes {
			for _, parts := range size.parts {
				var shaper HarfbuzzShaper
				out := shaper.Shape(Input{
					Text:      langInfo.text[:size.runes],
					RunStart:  0,
					RunEnd:    size.runes,
					Direction: langInfo.dir,
					Face:      langInfo.face,
					Size:      16 * 72,
					Script:    langInfo.script,
					Language:  langInfo.lang,
				})
				outs := cutRunInto(out, parts)
				accountedRunes := make([]int, size.runes)
				maxRune := -1
				for _, part := range outs {
					for i := part.Runes.Offset; i < part.Runes.Count+part.Runes.Offset; i++ {
						accountedRunes[i]++
						if i > maxRune {
							maxRune = i
						}
					}
				}
				if maxRune != size.runes-1 {
					t.Errorf("maximum rune in cut result is %d, expected %d", maxRune, size.runes-1)
				}
				for runeIdx, count := range accountedRunes {
					if count != 1 {
						t.Errorf("rune at position %d seen %d times", runeIdx, count)
					}
				}
			}
		}
	}
}

func (w *wrapBuffer) stats() string {
	return fmt.Sprintf("paragraph: %d(%d), line: %d(%d), used: %d, exhausted: %v, alt: %d(%d)", len(w.paragraph), cap(w.paragraph), len(w.line), cap(w.line), w.lineUsed, w.lineExhausted, len(w.alt), cap(w.alt))
}

func TestWrapBuffer(t *testing.T) {
	t.Run("new and reset have same state", func(t *testing.T) {
		b1 := wrapBuffer{}
		b1.reset()
		b2 := wrapBuffer{}
		b2.reset()
		defer func() {
			if t.Failed() {
				t.Logf("b1: %s\nb2: %s\n", b1.stats(), b2.stats())
			}
		}()
		b2.reset()
		if !reflect.DeepEqual(b1, b2) {
			t.Errorf("expected new and new+reset buffer to have same fields")
		}
	})
	t.Run("paragraph functions", func(t *testing.T) {
		b1 := wrapBuffer{}
		b1.reset()
		defer func() {
			if t.Failed() {
				t.Logf("b1: %s\n", b1.stats())
			}
		}()
		line := Line{Output{Advance: 10}}
		for i := 0; i < 5; i++ {
			maxLines := cap(b1.paragraph)
			startAddr := &b1.paragraph[0:1][0]
			b1.reset()
			for k := 0; k < maxLines; k++ {
				b1.paragraphAppend(line)
			}
			para := b1.finalParagraph()
			if actAddr := &para[0:1][0]; startAddr != actAddr {
				t.Errorf("expected paragraph to reuse slice starting at %p, got %p", startAddr, actAddr)
			}
		}
		for i := 0; i < 5; i++ {
			maxLines := cap(b1.paragraph)
			startAddr := &b1.paragraph[0:1][0]
			b1.reset()
			for k := 0; k < maxLines+1; k++ {
				b1.paragraphAppend(line)
			}
			para := b1.finalParagraph()
			if actAddr := &para[0:1][0]; startAddr == actAddr {
				t.Errorf("expected paragraph to enlarge slice starting at %p (changing start addres), but got %p", startAddr, actAddr)
			}
		}
		for i := 0; i < 5; i++ {
			startAddr := &b1.paragraph[0:1][0]
			b1.reset()
			para := b1.singleRunParagraph(line[0])
			if actAddr := &para[0:1][0]; startAddr != actAddr {
				t.Errorf("expected singleRunParagraph to reuse slice starting at %p, got %p", startAddr, actAddr)
			}
		}
	})
	t.Run("line building", func(t *testing.T) {
		b := wrapBuffer{}
		b.reset()
		defer func() {
			if t.Failed() {
				t.Logf("b: %s\n", b.stats())
			}
		}()
		b.startLine()
		run := Output{Advance: 10}
		lineLen := 0
		for i := 0; i < cap(b.line)-1; i++ {
			b.candidateAppend(run)
			lineLen++
			if b.hasBest() {
				t.Errorf("no best committed, but hasBest() true")
			}
		}
		best := b.finalizeBest()
		if best != nil {
			t.Errorf("no best committed, but finalizeBest() returned non-nil %#+v", best)
		}
		preCommitAltLen := len(b.alt)
		b.markCandidateBest(run)
		if postCommitAltLen := len(b.alt); preCommitAltLen != postCommitAltLen {
			t.Errorf("modified candidate when committing best, expected len %d, got %d", preCommitAltLen, postCommitAltLen)
		}
		if !b.hasBest() {
			t.Errorf("best committed, but hasBest() false")
		}
		preLineUsed := b.lineUsed
		best = b.finalizeBest()
		postLineUsed := b.lineUsed
		if len(best) != lineLen+1 {
			t.Errorf("expected best candidate to have len %d, got %d", lineLen+1, len(best))
		}
		if used := postLineUsed - preLineUsed; used != len(best) {
			t.Errorf("expected best candidate to use %d capacity of line, used %d", len(best), used)
		}
		if b.lineExhausted {
			t.Errorf("did not expect line to be exhausted yet")
		}
		b.startLine()
		b.candidateAppend(run)
		b.candidateAppend(run)
		if b.hasBest() {
			t.Errorf("no best committed, but hasBest() true")
		}
		b.markCandidateBest()
		if !b.hasBest() {
			t.Errorf("best committed, but hasBest() false")
		}
		preLineUsed = b.lineUsed
		best = b.finalizeBest()
		postLineUsed = b.lineUsed
		if len(best) != 2 {
			t.Errorf("expected best candidate to have len %d, got %d", 2, len(best))
		}
		if used := postLineUsed - preLineUsed; used != 0 {
			t.Errorf("expected best candidate to use %d capacity of line, used %d", 0, used)
		}
		if !b.lineExhausted {
			t.Errorf("expected line to be exhausted")
		}
		preResetCap := cap(b.line)
		b.reset()
		postResetCap := cap(b.line)
		if postResetCap <= preResetCap {
			t.Errorf("expected exhausted line to expand on reset")
		}
		if b.lineExhausted {
			t.Errorf("expected lineExhausted to clear on reset")
		}
	})
}

func TestLineWrapperBreakPolicies(t *testing.T) {
	type testcase struct {
		name      string
		paragraph []rune
	}
	for _, tc := range []testcase{
		{
			name:      "hello world",
			paragraph: []rune("hello, world"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := HarfbuzzShaper{}
			out := s.Shape(Input{
				Text:      tc.paragraph,
				RunStart:  0,
				RunEnd:    len(tc.paragraph),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(16),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			textWidth := out.Advance.Ceil()
			for maxWidth := textWidth; maxWidth > 0; maxWidth -= max(maxWidth/10, 1) {
				t.Run(fmt.Sprintf("maxWidth%d", maxWidth), func(t *testing.T) {
					for _, policy := range []LineBreakPolicy{WhenNecessary, Never, Always} {
						t.Run(policy.String(), func(t *testing.T) {
							w := LineWrapper{}
							lines, truncated := w.WrapParagraph(WrapConfig{
								BreakPolicy: policy,
							}, maxWidth, tc.paragraph, NewSliceIterator([]Output{out}))
							checkRuneCounts(t, tc.paragraph, lines, truncated)
						})
					}
				})
			}
		})
	}
}

// Checks that grapheme break in ligature are properly
// avoided
func TestGraphemeBreakLigature(t *testing.T) {
	const text1 = "il est affable"
	gls := append(
		glyphs(0, 7),        // simple 1:1 mapping
		ligatureGlyph(8, 2), // ligature
		simpleGlyph(10),     // a
		simpleGlyph(11),     // b
		simpleGlyph(12),     // l
		simpleGlyph(13),     // e
	)
	shapedText := Output{
		Advance: fixed.I(10 * len(gls)),
		Runes:   Range{Count: len([]rune(text1))},
		Glyphs:  gls,
		LineBounds: Bounds{
			Ascent:  fixed.I(10),
			Descent: fixed.I(5),
			// No line gap.
		},
		GlyphBounds: Bounds{
			Ascent: fixed.I(10),
			// No glyphs descend.
		},
	}
	var w LineWrapper
	lines, _ := w.WrapParagraph(WrapConfig{BreakPolicy: Always}, 90, []rune(text1), NewSliceIterator([]Output{shapedText}))
	if L := lines[0][0].Runes.Count; L != 10 {
		t.Errorf("invalid break with ligature %d", L)
	}
}

func (out Output) copy() Output {
	copy := out
	copy.Glyphs = append([]Glyph(nil), out.Glyphs...)
	return copy
}

func TestLineWrapperBreakSpecific(t *testing.T) {
	type expectation struct {
		name           string
		config         WrapConfig
		runeCounts     []int
		truncatedCount int
	}
	type testcase struct {
		expectations []expectation
		// This truncator will be inserted automatically into each expectation's WrapConfig
		// to avoid redundantly specifying it.
		truncator []rune
		maxWidth  int
		paragraph []rune
	}
	for _, tc := range []testcase{
		{
			paragraph: []rune("hello, world"),
			maxWidth:  80,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{7, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{11},
					truncatedCount: 1,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{11, 1},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{11},
					truncatedCount: 1,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  80,
			truncator: []rune("…"),
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{7, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{11, 1},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  60,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{7, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{8, 4},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{8},
					truncatedCount: 4,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			// maxWidth of 56 with this truncator results in ~40px of usable space for the text if the
			// truncator is present. The text "hello, " needs a little more space than this even if the
			// space is shortened to width zero.
			maxWidth:  56,
			truncator: []rune("…"),
			expectations: []expectation{
				{
					// We are forced to wrap within a word here.
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					// We are forced to wrap within a word here.
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					// Since we can't wrap within words in this configuration, we truncate all runes.
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{},
					truncatedCount: 12,
				},
			},
		},
		{
			// The first word "hello, " doesn't quite fit in 40 pixels. These tests exercise this boundary
			// condition with various break policies and truncation policies.
			paragraph: []rune("hello, world"),
			maxWidth:  40,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{5, 2, 5},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{5, 6, 1},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{5},
					truncatedCount: 7,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{},
					truncatedCount: 12,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  33,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{4, 3, 4, 1},
				},
				{
					name:           "WhenNecessary-Truncate",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 1},
					runeCounts:     []int{4},
					truncatedCount: 8,
				},
				{
					name:           "WhenNecessary-Truncate2",
					config:         WrapConfig{BreakPolicy: WhenNecessary, TruncateAfterLines: 2},
					runeCounts:     []int{4, 4},
					truncatedCount: 4,
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{4, 4, 4},
				},
				{
					name:           "Always-Truncate",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 1},
					runeCounts:     []int{4},
					truncatedCount: 8,
				},
				{
					name:           "Always-Truncate2",
					config:         WrapConfig{BreakPolicy: Always, TruncateAfterLines: 2},
					runeCounts:     []int{4, 4},
					truncatedCount: 4,
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
				{
					name:           "Never-Truncate",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 1},
					runeCounts:     []int{0},
					truncatedCount: 12,
				},
				{
					name:           "Never-Truncate2",
					config:         WrapConfig{BreakPolicy: Never, TruncateAfterLines: 2},
					runeCounts:     []int{7},
					truncatedCount: 5,
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  9,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{1, 1, 2, 1, 2, 1, 1, 1, 1, 1},
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{1, 1, 2, 1, 2, 1, 1, 1, 1, 1},
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
			},
		},
		{
			paragraph: []rune("hello, world"),
			maxWidth:  1,
			expectations: []expectation{
				{
					name:       "WhenNecessary",
					config:     WrapConfig{BreakPolicy: WhenNecessary},
					runeCounts: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
				{
					name:       "Always",
					config:     WrapConfig{BreakPolicy: Always},
					runeCounts: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				},
				{
					name:       "Never",
					config:     WrapConfig{BreakPolicy: Never},
					runeCounts: []int{7, 5},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%q-width%d-truncator%q", string(tc.paragraph), tc.maxWidth, string(tc.truncator)), func(t *testing.T) {
			s := HarfbuzzShaper{}
			out := s.Shape(Input{
				Text:      tc.paragraph,
				RunStart:  0,
				RunEnd:    len(tc.paragraph),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(16),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			shapedTrunc := s.Shape(Input{
				Text:      tc.truncator,
				RunStart:  0,
				RunEnd:    len(tc.truncator),
				Direction: di.DirectionLTR,
				Face:      benchEnFace,
				Size:      fixed.I(16),
				Script:    language.Latin,
				Language:  language.NewLanguage("EN"),
			})
			w := LineWrapper{}
			for _, expec := range tc.expectations {
				t.Run(expec.name, func(t *testing.T) {
					expec.config.Truncator = shapedTrunc
					lines, truncated := w.WrapParagraph(expec.config, tc.maxWidth, tc.paragraph, NewSliceIterator([]Output{out.copy()}))
					actualCounts := checkRuneCounts(t, tc.paragraph, lines, truncated)
					if truncated != expec.truncatedCount {
						t.Errorf("expected %d truncated runes, got %d", expec.truncatedCount, truncated)
					}
					for lineNo, expected := range expec.runeCounts {
						if len(actualCounts) <= lineNo {
							t.Errorf("expected %d lines, got %d", len(expec.runeCounts), len(actualCounts))
							break
						}
						actual := actualCounts[lineNo]
						if actual != expected {
							t.Errorf("line %d: expected %d runes, got %d", lineNo, expected, actual)
						}
					}
				})
			}
		})
	}
}

func BenchmarkWrapping(b *testing.B) {
	for _, langInfo := range benchLangs {
		for _, size := range benchSizes {
			for _, parts := range size.parts {
				b.Run(fmt.Sprintf("%drunes-%s-%dparts", size.runes, langInfo.name, parts), func(b *testing.B) {
					var shaper HarfbuzzShaper
					out := shaper.Shape(Input{
						Text:      langInfo.text[:size.runes],
						RunStart:  0,
						RunEnd:    size.runes,
						Direction: langInfo.dir,
						Face:      langInfo.face,
						Size:      16 * 72,
						Script:    langInfo.script,
						Language:  langInfo.lang,
					})
					outs := cutRunInto(out, parts)
					var l LineWrapper
					iter := NewSliceIterator(outs)
					b.ResetTimer()
					lines := make([]Line, 1)
					for i := 0; i < b.N; i++ {
						lines, _ = l.WrapParagraph(WrapConfig{}, 100, langInfo.text[:size.runes], iter)
						iter.(*shapedRunSlice).Reset(outs)
					}
					_ = lines
				})
			}
		}
	}
}

// BenchmarkWrappingHappyPath measures the performance when it's obvious that
// the shaped text will fit within the available space without doing any extra
// work.
func BenchmarkWrappingHappyPath(b *testing.B) {
	textInput := []rune("happy path")
	face := benchEnFace
	var shaper HarfbuzzShaper
	out := []Output{shaper.Shape(Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      16 * 72,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})}
	var l LineWrapper
	iter := NewSliceIterator(out)
	b.ResetTimer()
	var outs []Line
	for i := 0; i < b.N; i++ {
		outs, _ = l.WrapParagraph(WrapConfig{BreakPolicy: Never}, 100, textInput, iter)
		iter.(*shapedRunSlice).Reset(out)
	}
	_ = outs
}

const benchParagraphLatin = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Porttitor eget dolor morbi non arcu risus quis. Nibh sit amet commodo nulla. Posuere ac ut consequat semper viverra nam libero justo. Risus in hendrerit gravida rutrum quisque. Natoque penatibus et magnis dis parturient montes nascetur. In metus vulputate eu scelerisque felis imperdiet proin fermentum. Mattis rhoncus urna neque viverra. Elit pellentesque habitant morbi tristique. Nisl nunc mi ipsum faucibus vitae aliquet nec. Sed augue lacus viverra vitae congue eu consequat. At quis risus sed vulputate odio ut. Sit amet volutpat consequat mauris nunc congue nisi. Dignissim cras tincidunt lobortis feugiat. Faucibus turpis in eu mi bibendum. Odio aenean sed adipiscing diam donec adipiscing tristique. Fermentum leo vel orci porta non pulvinar. Ut venenatis tellus in metus vulputate eu scelerisque felis imperdiet. Et netus et malesuada fames ac turpis. Venenatis urna cursus eget nunc scelerisque viverra mauris in. Risus ultricies tristique nulla aliquet enim tortor. Risus pretium quam vulputate dignissim suspendisse in. Interdum velit euismod in pellentesque massa placerat duis ultricies lacus. Proin gravida hendrerit lectus a. Auctor augue mauris augue neque gravida in fermentum et. Laoreet sit amet cursus sit amet dictum. In fermentum et sollicitudin ac orci phasellus egestas tellus rutrum. Tempus imperdiet nulla malesuada pellentesque elit eget gravida. Consequat id porta nibh venenatis cras sed. Vulputate ut pharetra sit amet aliquam. Congue mauris rhoncus aenean vel elit. Risus quis varius quam quisque id diam vel quam elementum. Pretium lectus quam id leo in vitae. Sed sed risus pretium quam vulputate dignissim suspendisse in est. Velit laoreet id donec ultrices. Nunc sed velit dignissim sodales ut. Nunc scelerisque viverra mauris in aliquam sem fringilla ut. Sed enim ut sem viverra aliquet eget sit. Convallis posuere morbi leo urna molestie at. Aliquam id diam maecenas ultricies mi eget mauris. Ipsum dolor sit amet consectetur adipiscing elit ut aliquam. Accumsan tortor posuere ac ut consequat semper. Viverra vitae congue eu consequat ac felis donec et odio. Scelerisque in dictum non consectetur a. Consequat nisl vel pretium lectus quam id leo in vitae. Morbi tristique senectus et netus et malesuada fames ac turpis. Ac orci phasellus egestas tellus. Tempus egestas sed sed risus. Ullamcorper morbi tincidunt ornare massa eget egestas purus. Nibh venenatis cras sed felis eget velit.`

const benchParagraphArabic = `و سأعرض مثال حي لهذا، من منا لم يتحمل جهد بدني شاق إلا من أجل الحصول على ميزة أو فائدة؟ ولكن من لديه الحق أن ينتقد شخص ما أراد أن يشعر بالسعادة التي لا تشوبها عواقب أليمة أو آخر أراد أن يتجنب الألم الذي ربما تنجم عنه بعض المتعة ؟ علي الجانب الآخر نشجب ونستنكر هؤلاء الرجال المفتونون بنشوة اللحظة الهائمون في رغباتهم فلا يدركون ما يعقبها من الألم والأسي المحتم، واللوم كذلك يشمل هؤلاء الذين أخفقوا في واجباتهم نتيجة لضعف إرادتهم فيتساوي مع هؤلاء الذين يتجنبون وينأون عن تحمل الكدح والألم . من المفترض أن نفرق بين هذه الحالات بكل سهولة ومرونة. في ذاك الوقت عندما تكون قدرتنا علي الاختيار غير مقيدة بشرط وعندما لا نجد ما يمنعنا أن نفعل الأفضل فها نحن نرحب بالسرور والسعادة ونتجنب كل ما يبعث إلينا الألم. في بعض الأحيان ونظراً للالتزامات التي يفرضها علينا الواجب والعمل سنتنازل غالباً ونرفض الشعور بالسرور ونقبل ما يجلبه إلينا الأسى. الإنسان الحكيم عليه أن يمسك زمام الأمور ويختار إما أن يرفض مصادر السعادة من أجل ما هو أكثر أهمية أو يتحمل الألم من أجل ألا يتحمل ما هو أسوأ. و سأعرض مثال حي لهذا، من منا لم يتحمل جهد بدني شاق إلا من أجل الحصول على ميزة أو فائدة؟ ولكن من لديه الحق أن ينتقد شخص ما أراد أن يشعر بالسعادة التي لا تشوبها عواقب أليمة أو آخر أراد أن يتجنب الألم الذي ربما تنجم عنه بعض المتعة ؟ علي الجانب الآخر نشجب ونستنكر هؤلاء الرجال المفتونون بنشوة اللحظة الهائمون في رغباتهم فلا يدركون ما يعقبها من الألم والأسي المحتم، واللوم كذلك يشمل هؤلاء الذين أخفقوا في واجباتهم نتيجة لضعف إرادتهم فيتساوي مع هؤلاء الذين يتجنبون وينأون عن تحمل الكدح والألم . من المفترض أن نفرق بين هذه الحالات بكل سهولة ومرونة. في ذاك الوقت عندما تكون قدرتنا علي الاختيار غير مقيدة بشرط وعندما لا نجد ما يمنعنا أن نفعل الأفضل فها نحن نرحب بالسرور والسعادة ونتجنب كل ما يبعث إلينا الألم. في بعض الأحيان ونظراً للالتزامات التي يفرضها علينا الواجب والعمل سنتنازل غالباً ونرفض الشعور بالسرور ونقبل ما يجلبه إلينا الأسى. الإنسان الحكيم عليه أن يمسك زمام الأمور ويختار إما أن يرفض مصادر السعادة من أجل ما هو أكثر أهمية أو يتحمل الألم من أجل ألا يتحمل ما هو أسوأ.`

var regressionRuns = []Output{
	{
		Advance: 5490,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 0, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 1, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 2, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 3, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 4, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 5, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 6, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 7, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 8, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 9, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 10, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 11, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 12, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 13, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 14, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 464, Descent: -6, Gap: 0,
		},
		Direction: 0x0,
		Runes: Range{
			Offset: 0, Count: 15,
		},
	},
	{
		Advance: 166,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 0, Height: 0, XBearing: 0, YBearing: 0, XAdvance: 166, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 15, RuneCount: 1, GlyphCount: 1, GlyphID: 0x3, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 0, Descent: 0, Gap: 0,
		},
		Direction: 0x1,
		Runes: Range{
			Offset: 15, Count: 1,
		},
	},
	{
		Advance: 2928,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 16, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 17, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 18, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 19, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 20, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 21, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 22, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 23, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 464, Descent: -6, Gap: 0,
		},
		Direction: 0x0,
		Runes: Range{
			Offset: 16, Count: 8,
		},
	},
	{
		Advance: 166,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 0, Height: 0, XBearing: 0, YBearing: 0, XAdvance: 166, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 24, RuneCount: 2, GlyphCount: 2, GlyphID: 0x3, Mask: 0x80000000,
			},
			{
				Width: 39, Height: -111, XBearing: 29, YBearing: 495, XAdvance: 0, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 24, RuneCount: 2, GlyphCount: 2, GlyphID: 0x57, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 495, Descent: 0, Gap: 0,
		},
		Direction: 0x1,
		Runes: Range{
			Offset: 24, Count: 2,
		},
	},
	{
		Advance: 1830,
		Size:    640,
		Glyphs: []Glyph{
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 26, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 27, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 28, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 29, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
			{
				Width: 303, Height: -470, XBearing: 31, YBearing: 464, XAdvance: 366, YAdvance: 0, XOffset: 0, YOffset: 0, ClusterIndex: 30, RuneCount: 1, GlyphCount: 1, GlyphID: 0x65f, Mask: 0x80000000,
			},
		},
		LineBounds: Bounds{
			Ascent: 879, Descent: -472, Gap: 0,
		},
		GlyphBounds: Bounds{
			Ascent: 464, Descent: -6, Gap: 0,
		},
		Direction: 0x0,
		Runes: Range{
			Offset: 26, Count: 5,
		},
	},
}

func TestTrailingSpace(t *testing.T) {
	// assume single run, 1 to 1 rune <-> glyph mapping
	shapedText := func(text []rune, line Line) string {
		tu.Assert(t, len(line) == 1)
		runLine := line[0]
		var s []rune
		for _, g := range runLine.Glyphs {
			s = append(s, text[g.ClusterIndex])
		}
		return string(s)
	}

	text := []rune(" The quick brown fox")

	face := loadOpentypeFont(t, "../font/testdata/UbuntuMono-R.ttf")
	run := (&HarfbuzzShaper{}).Shape(Input{
		Text:   text,
		Face:   face,
		Size:   72,
		RunEnd: len(text),
	})
	tu.Assert(t, run.Glyphs[0].XAdvance == fixed.I(1))

	lines, _ := (&LineWrapper{}).WrapParagraph(WrapConfig{BreakPolicy: Always}, 5, text, NewSliceIterator([]Output{run}))

	expected := []string{
		" The ",
		"quick ",
		"brown ",
		"fox",
	}
	tu.Assert(t, len(lines) == 4)
	for i, line := range lines {
		s := shapedText(text, line)
		tu.Assert(t, s == expected[i])
	}
}
