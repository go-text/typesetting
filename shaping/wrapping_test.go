package shaping

import (
	"bytes"
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
	"github.com/go-text/typesetting/di"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

// glyph returns a glyph with the given cluster. Its dimensions
// are a square sitting atop the baseline, with 10 units to a side.
func glyph(cluster int) Glyph {
	return Glyph{
		XAdvance:     fixed.I(10),
		YAdvance:     fixed.I(10),
		Width:        fixed.I(10),
		Height:       fixed.I(10),
		YBearing:     fixed.I(10),
		ClusterIndex: cluster,
	}
}

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
	for i := start; i >= 0 && i <= max(start, end); i += inc {
		g = append(g, glyph(i))
	}
	return g
}

func TestMapRunesToClusterIndices(t *testing.T) {
	type testcase struct {
		name     string
		runes    []rune
		glyphs   []Glyph
		expected []int
	}
	for _, tc := range []testcase{
		{
			name:  "simple",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(0),
				glyph(1),
				glyph(2),
				glyph(3),
				glyph(4),
			},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:  "simple rtl",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(4),
				glyph(3),
				glyph(2),
				glyph(1),
				glyph(0),
			},
			expected: []int{4, 3, 2, 1, 0},
		},
		{
			name:  "fused clusters",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(0),
				glyph(0),
				glyph(2),
				glyph(3),
				glyph(3),
			},
			expected: []int{0, 0, 2, 3, 3},
		},
		{
			name:  "fused clusters rtl",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(3),
				glyph(3),
				glyph(2),
				glyph(0),
				glyph(0),
			},
			expected: []int{3, 3, 2, 0, 0},
		},
		{
			name:  "ligatures",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(0),
				glyph(2),
				glyph(3),
			},
			expected: []int{0, 0, 1, 2, 2},
		},
		{
			name:  "ligatures rtl",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(3),
				glyph(2),
				glyph(0),
			},
			expected: []int{2, 2, 1, 0, 0},
		},
		{
			name:  "expansion",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(0),
				glyph(1),
				glyph(1),
				glyph(1),
				glyph(2),
				glyph(3),
				glyph(4),
			},
			expected: []int{0, 1, 4, 5, 6},
		},
		{
			name:  "expansion rtl",
			runes: make([]rune, 5),
			glyphs: []Glyph{
				glyph(4),
				glyph(3),
				glyph(2),
				glyph(1),
				glyph(1),
				glyph(1),
				glyph(0),
			},
			expected: []int{6, 3, 2, 1, 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mapping := mapRunesToClusterIndices(tc.runes, tc.glyphs)
			if !reflect.DeepEqual(tc.expected, mapping) {
				t.Errorf("expected %v, got %v", tc.expected, mapping)
			}
		})
	}
}

func TestInclusiveRange(t *testing.T) {
	type testcase struct {
		name string
		// inputs
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
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          0,
			ge:          2,
		},
		{
			name:        "simple in middle",
			numGlyphs:   5,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          1,
			ge:          3,
		},
		{
			name:        "simple at end",
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 1, 2, 3, 4},
			gs:          2,
			ge:          4,
		},
		{
			name:        "simple at start rtl",
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          2,
			ge:          4,
		},
		{
			name:        "simple in middle rtl",
			numGlyphs:   5,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          1,
			ge:          3,
		},
		{
			name:        "simple at end rtl",
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{4, 3, 2, 1, 0},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters at start",
			numGlyphs:   5,
			start:       0,
			breakAfter:  1,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          0,
			ge:          1,
		},
		{
			name:        "fused clusters start and middle",
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters middle and end",
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          2,
			ge:          4,
		},
		{
			name:        "fused clusters at end",
			numGlyphs:   5,
			start:       3,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 2, 3, 3},
			gs:          3,
			ge:          4,
		},
		{
			name:        "fused clusters at start rtl",
			numGlyphs:   5,
			start:       0,
			breakAfter:  1,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          3,
			ge:          4,
		},
		{
			name:        "fused clusters start and middle rtl",
			numGlyphs:   5,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          2,
			ge:          4,
		},
		{
			name:        "fused clusters middle and end rtl",
			numGlyphs:   5,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          0,
			ge:          2,
		},
		{
			name:        "fused clusters at end rtl",
			numGlyphs:   5,
			start:       3,
			breakAfter:  4,
			runeToGlyph: []int{3, 3, 2, 0, 0},
			gs:          0,
			ge:          1,
		},
		{
			name:        "ligatures at start",
			numGlyphs:   3,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          0,
			ge:          1,
		},
		{
			name:        "ligatures in middle",
			numGlyphs:   3,
			start:       2,
			breakAfter:  2,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          1,
			ge:          1,
		},
		{
			name:        "ligatures at end",
			numGlyphs:   3,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 0, 1, 2, 2},
			gs:          1,
			ge:          2,
		},
		{
			name:        "ligatures at start rtl",
			numGlyphs:   3,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          1,
			ge:          2,
		},
		{
			name:        "ligatures in middle rtl",
			numGlyphs:   3,
			start:       2,
			breakAfter:  2,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          1,
			ge:          1,
		},
		{
			name:        "ligatures at end rtl",
			numGlyphs:   3,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{2, 2, 1, 0, 0},
			gs:          0,
			ge:          1,
		},
		{
			name:        "expansion at start",
			numGlyphs:   7,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          0,
			ge:          4,
		},
		{
			name:        "expansion in middle",
			numGlyphs:   7,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          1,
			ge:          5,
		},
		{
			name:        "expansion at end",
			numGlyphs:   7,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{0, 1, 4, 5, 6},
			gs:          4,
			ge:          6,
		},
		{
			name:        "expansion at start rtl",
			numGlyphs:   7,
			start:       0,
			breakAfter:  2,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          2,
			ge:          6,
		},
		{
			name:        "expansion in middle rtl",
			numGlyphs:   7,
			start:       1,
			breakAfter:  3,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          1,
			ge:          5,
		},
		{
			name:        "expansion at end rtl",
			numGlyphs:   7,
			start:       2,
			breakAfter:  4,
			runeToGlyph: []int{6, 3, 2, 1, 0},
			gs:          0,
			ge:          2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gs, ge := inclusiveGlyphRange(tc.start, tc.breakAfter, tc.runeToGlyph, tc.numGlyphs)
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
	}
	text1Trailing       = text1 + " "
	shapedText1Trailing = func() Output {
		out := shapedText1
		out.Glyphs = append(out.Glyphs, glyph(len(out.Glyphs)))
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
			0: glyph(0), // 안        - 4 bytes
			1: glyph(1), // П         - 3 bytes
			2: glyph(2), // 你        - 4 bytes
			3: glyph(3), // <space>   - 1 byte
			4: glyph(4), // lig       - 3 runes, 3 bytes
			// DROP                   - 4 runes, 4 bytes
			5:  glyph(11), // <space> - 1 byte
			6:  glyph(12), // 안      - 4 bytes
			7:  glyph(13), // П       - 3 bytes
			8:  glyph(14), // 你      - 4 bytes
			9:  glyph(15), // <space> - 1 byte
			10: glyph(16), // lig     - 3 runes, 3 bytes
			// DROP                   - 4 runes, 4 bytes
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
			0: glyph(16), // LIGATURE of three runes:
			//               ا - 3 bytes
			//               ل - 3 bytes
			//               ه - 3 bytes
			1: glyph(15), // أ - 3 bytes
			2: glyph(14), // <space> - 1 byte
			3: glyph(13), // ם - 3 bytes
			4: glyph(12), // ו - 3 bytes
			5: glyph(11), // ל - 3 bytes
			6: glyph(10), // ש - 3 bytes
			7: glyph(9),  // <space> - 1 byte
			8: glyph(6),  // LIGATURE of three runes:
			//               ا - 3 bytes
			//               ل - 3 bytes
			//               ه - 3 bytes
			9:  glyph(5), // أ - 3 bytes
			10: glyph(4), // <space> - 1 byte
			11: glyph(3), // ם - 3 bytes
			12: glyph(2), // ו - 3 bytes
			13: glyph(1), // ל - 3 bytes
			14: glyph(0), // ש - 3 bytes
		},
	}
)

// splitShapedAt splits a single shaped output into multiple. It splits
// on each provided glyph index in indices, with the index being the end of
// a slice range (so it's exclusive). You can think of the index as the
// first glyph of the next output.
func splitShapedAt(shaped Output, direction di.Direction, indices ...int) []Output {
	numOut := len(indices) + 1
	outputs := make([]Output, 0, numOut)
	start := 0
	for _, i := range indices {
		newOut := shaped
		newOut.Glyphs = newOut.Glyphs[start:i]
		newOut.RecalculateAll()
		outputs = append(outputs, newOut)
		start = i
	}
	newOut := shaped
	newOut.Glyphs = newOut.Glyphs[start:]
	newOut.RecalculateAll()
	outputs = append(outputs, newOut)
	return outputs
}

func TestLineWrap(t *testing.T) {
	type testcase struct {
		name      string
		shaped    []Output
		paragraph []rune
		maxWidth  int
		expected  []Line
	}
	for _, tc := range []testcase{
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
			// This test case verifies that line breaking does occur, and that
			// all lines have proper offsets.
			name:      "line break on last word",
			shaped:    []Output{shapedText1},
			paragraph: []rune(text1),
			maxWidth:  120,
			expected: []Line{
				[]Output{
					withRange(
						splitShapedAt(shapedText1, di.DirectionLTR, 12)[0],
						Range{
							Count: len([]rune(text1)) - 3,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, di.DirectionLTR, 12)[1],
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
						splitShapedAt(shapedText1, di.DirectionLTR, 5)[0],
						Range{
							Count: 5,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, di.DirectionLTR, 5, 12)[1],
						Range{
							Offset: 5,
							Count:  7,
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText1, di.DirectionLTR, 12)[1],
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
						splitShapedAt(shapedText2, di.DirectionLTR, 4)[0],
						Range{
							Count: len([]rune("안П你 ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText2, di.DirectionLTR, 4, 8)[1],
						Range{
							Count:  len([]rune("ligDROP 안П")),
							Offset: len([]rune("안П你 ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText2, di.DirectionLTR, 8, 11)[1],
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
						splitShapedAt(shapedText3, di.DirectionRTL, 7)[1],
						Range{
							Count: len([]rune("שלום أهلا ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, di.DirectionRTL, 7)[0],
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
						splitShapedAt(shapedText3, di.DirectionRTL, 10)[1],
						Range{
							Count: len([]rune("שלום ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, di.DirectionRTL, 7, 10)[1],
						Range{
							Count:  len([]rune("أهلا ")),
							Offset: len([]rune("שלום ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, di.DirectionRTL, 2, 7)[1],
						Range{
							Count:  len([]rune("שלום ")),
							Offset: len([]rune("שלום أهلا ")),
						},
					),
				},
				[]Output{
					withRange(
						splitShapedAt(shapedText3, di.DirectionRTL, 2)[0],
						Range{
							Count:  len([]rune("أهلا")),
							Offset: len([]rune("שלום أهلا שלום ")),
						},
					),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sp := NewLineWrapper(tc.paragraph, tc.shaped...)

			outs := sp.WrapParagraph(tc.maxWidth)
			if len(tc.expected) != len(outs) {
				t.Errorf("expected %d lines, got %d", len(tc.expected), len(outs))
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

func TestGetBreakOptions(t *testing.T) {
	if err := quick.Check(func(runes []rune) bool {
		breaker := newBreaker(runes)
		var options []breakOption
		for b, ok := breaker.next(); ok; b, ok = breaker.next() {
			options = append(options, b)
		}

		// Ensure breaks are in valid range.
		for _, o := range options {
			if o.breakAtRune < 0 || o.breakAtRune > len(runes)-1 {
				return false
			}
		}
		// Ensure breaks are sorted.
		if !sort.SliceIsSorted(options, func(i, j int) bool {
			return options[i].breakAtRune < options[j].breakAtRune
		}) {
			return false
		}

		// Ensure breaks are unique.
		m := make([]bool, len(runes))
		for _, o := range options {
			if m[o.breakAtRune] {
				return false
			} else {
				m[o.breakAtRune] = true
			}
		}

		return true
	}, nil); err != nil {
		t.Errorf("generated invalid break options: %v", err)
	}
}

const benchParagraph = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Porttitor eget dolor morbi non arcu risus quis. Nibh sit amet commodo nulla. Posuere ac ut consequat semper viverra nam libero justo. Risus in hendrerit gravida rutrum quisque. Natoque penatibus et magnis dis parturient montes nascetur. In metus vulputate eu scelerisque felis imperdiet proin fermentum. Mattis rhoncus urna neque viverra. Elit pellentesque habitant morbi tristique. Nisl nunc mi ipsum faucibus vitae aliquet nec. Sed augue lacus viverra vitae congue eu consequat. At quis risus sed vulputate odio ut. Sit amet volutpat consequat mauris nunc congue nisi. Dignissim cras tincidunt lobortis feugiat. Faucibus turpis in eu mi bibendum. Odio aenean sed adipiscing diam donec adipiscing tristique. Fermentum leo vel orci porta non pulvinar. Ut venenatis tellus in metus vulputate eu scelerisque felis imperdiet. Et netus et malesuada fames ac turpis. Venenatis urna cursus eget nunc scelerisque viverra mauris in. Risus ultricies tristique nulla aliquet enim tortor. Risus pretium quam vulputate dignissim suspendisse in. Interdum velit euismod in pellentesque massa placerat duis ultricies lacus. Proin gravida hendrerit lectus a. Auctor augue mauris augue neque gravida in fermentum et. Laoreet sit amet cursus sit amet dictum. In fermentum et sollicitudin ac orci phasellus egestas tellus rutrum. Tempus imperdiet nulla malesuada pellentesque elit eget gravida. Consequat id porta nibh venenatis cras sed. Vulputate ut pharetra sit amet aliquam. Congue mauris rhoncus aenean vel elit. Risus quis varius quam quisque id diam vel quam elementum. Pretium lectus quam id leo in vitae. Sed sed risus pretium quam vulputate dignissim suspendisse in est. Velit laoreet id donec ultrices. Nunc sed velit dignissim sodales ut. Nunc scelerisque viverra mauris in aliquam sem fringilla ut. Sed enim ut sem viverra aliquet eget sit. Convallis posuere morbi leo urna molestie at. Aliquam id diam maecenas ultricies mi eget mauris. Ipsum dolor sit amet consectetur adipiscing elit ut aliquam. Accumsan tortor posuere ac ut consequat semper. Viverra vitae congue eu consequat ac felis donec et odio. Scelerisque in dictum non consectetur a. Consequat nisl vel pretium lectus quam id leo in vitae. Morbi tristique senectus et netus et malesuada fames ac turpis. Ac orci phasellus egestas tellus. Tempus egestas sed sed risus. Ullamcorper morbi tincidunt ornare massa eget egestas purus. Nibh venenatis cras sed felis eget velit.`

func BenchmarkWrapping(b *testing.B) {
	textInput := []rune(benchParagraph)
	face, err := truetype.Parse(bytes.NewReader(goregular.TTF))
	out, err := Shape(Input{
		Text:      textInput,
		RunStart:  0,
		RunEnd:    len(textInput),
		Direction: di.DirectionLTR,
		Face:      face,
		Size:      16,
		Script:    language.Latin,
		Language:  language.NewLanguage("EN"),
	})
	if err != nil {
		b.Skipf("failed shaping: %v", err)
	}
	b.ResetTimer()
	var outs []Line
	for i := 0; i < b.N; i++ {
		lw := NewLineWrapper(textInput, out)
		outs = lw.WrapParagraph(250)
	}
	_ = outs
}
