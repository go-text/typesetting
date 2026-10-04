package bidi

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	ucd "github.com/go-text/typesetting/internal/unicodedata"
	tu "github.com/go-text/typesetting/testutils"
)

func TestAnalysisBaseDirection(t *testing.T) {
	tests := []struct {
		text            string
		direction, want Direction
	}{
		{"", Neutral, LeftToRight},
		{"", RightToLeft, RightToLeft},
		{"abc", Neutral, LeftToRight},
		{"אב", Neutral, RightToLeft},
		{"abc", RightToLeft, RightToLeft},
		{"אב", LeftToRight, LeftToRight},
		{"123", Neutral, LeftToRight},
		{"\u2067אב\u2069abc", Neutral, LeftToRight},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q/%d", tt.text, tt.direction), func(t *testing.T) {
			analysis, err := Analyze([]rune(tt.text), tt.direction)
			if err != nil {
				t.Fatal(err)
			}
			if got := analysis.BaseDirection(); got != tt.want {
				t.Fatalf("base direction = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalysisInvalidInput(t *testing.T) {
	if _, err := Analyze(nil, Direction(255)); err == nil {
		t.Error("Analyze accepted an invalid direction")
	}
}

func TestAnalysisFirstParagraph(t *testing.T) {
	for _, tt := range []struct {
		text       string
		lengths    []int
		directions []Direction
	}{
		{"a\nb", []int{2, 1}, []Direction{LeftToRight, LeftToRight}},
		{"א\u2029a", []int{2, 1}, []Direction{RightToLeft, LeftToRight}},
		{"\n\n", []int{1, 1}, []Direction{LeftToRight, LeftToRight}},
		{"a\r\nb", []int{2, 1, 1}, []Direction{LeftToRight, LeftToRight, LeftToRight}},
		{"a\u2028b", []int{3}, []Direction{LeftToRight}},
		{"a\n", []int{2}, []Direction{LeftToRight}},
		{"abc", []int{3}, []Direction{LeftToRight}},
	} {
		runes := []rune(tt.text)
		at := 0
		for i, length := range tt.lengths {
			analysis, err := Analyze(runes[at:], Neutral)
			if err != nil {
				t.Fatalf("Analyze(%q): %v", tt.text, err)
			}
			if analysis.Len() != length || analysis.BaseDirection() != tt.directions[i] {
				t.Fatalf("Analyze(%q), paragraph %d: length %d, direction %d; want %d, %d", tt.text, i, analysis.Len(), analysis.BaseDirection(), length, tt.directions[i])
			}
			if _, err := analysis.Line(0, analysis.Len()); err != nil {
				t.Fatal(err)
			}
			if analysis.Len() < len(runes)-at {
				if _, err := analysis.Line(0, len(runes)-at); err == nil {
					t.Fatal("Line accepted a range across paragraphs")
				}
			}
			at += analysis.Len()
		}
		if at != len(runes) {
			t.Fatalf("consumed %d runes, want %d", at, len(runes))
		}
	}
}

// Test created from https://github.com/golang/go/issues/69819
func TestNestedIsolates(t *testing.T) {
	str := "The title is \u2067אבג \u2066C++\u2069 דהו\u2069 in Hebrew."
	runs := (&Paragraph{}).Segment([]rune(str), LeftToRight)

	expectedRuns := []Run{
		{0, 14, 0},
		{14, 19, 1},
		{19, 22, 2},
		{22, 27, 1},
		{27, 39, 0},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.Level == want.Level)
	}
}

// Test copied from https://github.com/golang/go/issues/71809
func TestN2(t *testing.T) {
	str := `ع a`
	runs := (&Paragraph{}).Segment([]rune(str), LeftToRight)

	expectedRuns := []Run{
		{0, 1, 1},
		{1, 3, 0},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.IsLeftToRight() == want.IsLeftToRight())
	}
}

// U+05FF is unassigned. DerivedBidiClass @missing gives it R. With class 0
// instead, the text splits into three runs.
func TestUnassignedDefaultsToBlockClass(t *testing.T) {
	runs := (&Paragraph{}).Segment([]rune("א\u05FFב"), RightToLeft)
	tu.Assert(t, runs.NumRuns() == 1)
	tu.Assert(t, !runs.Run(0).IsLeftToRight())
}

func TestSpaces(t *testing.T) {
	str := `ااب   `
	runs := (&Paragraph{}).Segment([]rune(str), LeftToRight)

	expectedRuns := []Run{
		{0, 3, 1},
		{3, 6, 0},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.IsLeftToRight() == want.IsLeftToRight())
	}
}

func TestStringBytes(t *testing.T) {
	str := "The title is \u2067אבג \u2066C++\u2069 דהו\u2069 in Hebrew."
	outRunes := (&Paragraph{}).Segment([]rune(str), Neutral)
	outString := (&Paragraph{}).SegmentString(str, Neutral)
	outBytes := (&Paragraph{}).SegmentBytes([]byte(str), Neutral)
	tu.Assert(t, reflect.DeepEqual(outRunes, outString) && reflect.DeepEqual(outString, outBytes))
}

// Tests copied from x/text

func TestSimple(t *testing.T) {
	str := "He l-lö.9"
	runs := (&Paragraph{}).Segment([]rune(str), Neutral)

	expectedRuns := []Run{
		{0, 9, 0},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.IsLeftToRight() == want.IsLeftToRight())
	}
}

func TestMixed(t *testing.T) {
	str := `العاشر ليونيكود (Unicode Conference)، الذي سيعقد في 10-12 آذار 1997 مبدينة`
	runs := (&Paragraph{}).Segment([]rune(str), Neutral)

	expectedRuns := []Run{
		{0, 16 + 1, 1},
		{17, 34 + 1, 0},
		{35, 51 + 1, 1},
		{52, 53 + 1, 0},
		{54, 54 + 1, 1},
		{55, 56 + 1, 0},
		{57, 62 + 1, 1},
		{63, 66 + 1, 0},
		{67, 73 + 1, 1},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.IsLeftToRight() == want.IsLeftToRight())
	}
}

func TestExplicitIsolate(t *testing.T) {
	// https://www.w3.org/International/articles/inline-bidi-markup/uba-basics.en#beyond
	str := "The names of these states in Arabic are \u2067مصر\u2069, \u2067البحرين\u2069 and \u2067الكويت\u2069 respectively."
	runs := (&Paragraph{}).Segment([]rune(str), Neutral)

	expectedRuns := []Run{
		{0, 40 + 1, 0},
		{41, 43 + 1, 1},
		{44, 47 + 1, 0},
		{48, 54 + 1, 1},
		{55, 61 + 1, 0},
		{62, 67 + 1, 1},
		{68, 82 + 1, 0},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.IsLeftToRight() == want.IsLeftToRight())
	}
}

func TestWithoutExplicitIsolate(t *testing.T) {
	str := "The names of these states in Arabic are مصر, البحرين and الكويت respectively."
	runs := (&Paragraph{}).Segment([]rune(str), Neutral)

	expectedRuns := []Run{
		{0, 39 + 1, 0},
		{40, 51 + 1, 1},
		{52, 56 + 1, 0},
		{57, 62 + 1, 1},
		{63, 76 + 1, 0},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.IsLeftToRight() == want.IsLeftToRight())
	}
}

func TestMixedSimple(t *testing.T) {
	str := `Uا`
	runs := (&Paragraph{}).Segment([]rune(str), Neutral)

	expectedRuns := []Run{
		{0, 0 + 1, 0},
		{1, 1 + 1, 1},
	}

	tu.Assert(t, runs.NumRuns() == len(expectedRuns))
	for i, want := range expectedRuns {
		r := runs.Run(i)
		tu.Assert(t, r.Start == want.Start && r.End == want.End)
		tu.Assert(t, r.IsLeftToRight() == want.IsLeftToRight())
	}
}

func TestDefaultDirection(t *testing.T) {
	str := "+"
	runs := (&Paragraph{}).Segment([]rune(str), RightToLeft)
	tu.Assert(t, runs.Run(0).IsLeftToRight() == false)

	runs = (&Paragraph{}).Segment([]rune(str), LeftToRight)
	tu.Assert(t, runs.Run(0).IsLeftToRight() == true)
}

func TestEmpty(t *testing.T) {
	runs := (&Paragraph{}).Segment(nil, Neutral)
	tu.Assert(t, runs.NumRuns() == 0)
}

func TestNewline(t *testing.T) {
	c, _ := ucd.LookupBidiClass('\n')
	tu.Assert(t, c == ucd.BD_B) // paragraph separator
	c, _ = ucd.LookupBidiClass('\u2028')
	tu.Assert(t, c == ucd.BD_WS) // line separator

	runs := (&Paragraph{}).Segment([]rune("Hello\nworld"), Neutral)
	// 6 is the length up to and including the \n
	tu.Assert(t, runs.Run(0).End == 6)

	runs = (&Paragraph{}).Segment([]rune("Hello\u2028world"), Neutral)
	tu.Assert(t, runs.Run(0).End == 11)
}

func TestDoubleSetString(t *testing.T) {
	str := "العاشر ليونيكود (Unicode Conference)،"
	var p Paragraph
	_ = p.Segment([]rune(str), Neutral)
	_ = p.Segment([]rune(str), Neutral)
}

// ------------------------- Unicode conformance tests -------------------------

//go:embed test/BidiCharacterTest.txt
var bidiCharacterTestSrc []byte

//go:embed test/BidiTest.txt
var bidiTestSrc []byte

func parseOrdering(line string) ([]int, error) {
	fields := strings.Fields(line)
	out := make([]int, len(fields))
	for i, posLit := range fields {
		pos, err := strconv.Atoi(posLit)
		if err != nil {
			return nil, fmt.Errorf("invalid position %s: %s", posLit, err)
		}
		out[i] = pos
	}
	return out, nil
}

func parseLevels(line string) ([]Level, error) {
	fields := strings.Fields(line)
	out := make([]Level, len(fields))
	for i, f := range fields {
		if f == "x" {
			out[i] = -1
		} else {
			lev, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("invalid level %s: %s", f, err)
			}
			out[i] = Level(lev)
		}
	}
	return out, nil
}

type testData struct {
	line       int // just for easier debugging
	codePoints []rune
	parDir     Direction

	expectedLevels []Level

	visualOrdering   []int
	resolvedParLevel int
}

func parseTestLine(line []byte, lineNumber int) (out testData, err error) {
	out.line = lineNumber

	fields := strings.Split(string(line), ";")
	if len(fields) < 5 {
		return out, fmt.Errorf("invalid line %s", line)
	}

	//  Field 0. Code points
	for _, runeLit := range strings.Fields(fields[0]) {
		var c rune
		if _, err = fmt.Sscanf(runeLit, "%04x", &c); err != nil {
			return out, fmt.Errorf("invalid rune %s: %s", runeLit, err)
		}
		out.codePoints = append(out.codePoints, c)
	}

	// Field 1. Paragraph direction
	parDir, err := strconv.Atoi(fields[1])
	if err != nil {
		return out, fmt.Errorf("invalid paragraph direction %s: %s", fields[1], err)
	}

	switch parDir {
	case 0:
		out.parDir = LeftToRight
	case 1:
		out.parDir = RightToLeft
	case 2:
		out.parDir = Neutral
	default:
		return out, fmt.Errorf("unsupported paragraph direction %d", parDir)
	}

	// Field 2. resolved paragraph_dir
	out.resolvedParLevel, err = strconv.Atoi(fields[2])
	if err != nil {
		return out, fmt.Errorf("invalid resolved paragraph embedding level %s: %s", fields[2], err)
	}

	// Field 3. resolved levels (or -1)
	out.expectedLevels, err = parseLevels(fields[3])
	if err != nil {
		return out, err
	}

	if len(out.expectedLevels) != len(out.codePoints) {
		return out, errors.New("different lengths for levels and codepoints")
	}

	//  Field 4 - resulting visual ordering
	out.visualOrdering, err = parseOrdering(fields[4])

	return out, err
}

func parseBidiCharacterTests() ([]testData, error) {
	var out []testData
	for lineNumber, line := range bytes.Split(bidiCharacterTestSrc, []byte{'\n'}) {
		if line := bytes.TrimSpace(line); len(line) == 0 || line[0] == '#' {
			continue
		}

		lineData, err := parseTestLine(line, lineNumber+1)
		if err != nil {
			return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
		}
		out = append(out, lineData)
	}
	return out, nil
}

func TestBidiCharacters(t *testing.T) {
	datas, err := parseBidiCharacterTests()
	tu.AssertNoErr(t, err)

	for _, test := range datas {
		levels := (&Paragraph{}).Segment(test.codePoints, test.parDir).levels

		/* Compare */
		for i, level := range levels {
			if exp := test.expectedLevels[i]; level != exp && exp != -1 {
				t.Fatalf("failure at line %d: levels[%d]: expected %d, got %d", test.line, i, exp, level)
				break
			}
		}
	}
}

func TestAnalysisBidiCharacters(t *testing.T) {
	t.Parallel()
	datas, err := parseBidiCharacterTests()
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range datas {
		analysis, err := Analyze(data.codePoints, data.parDir)
		if err != nil {
			t.Fatalf("line %d: %v", data.line, err)
		}
		base := 0
		if analysis.BaseDirection() == RightToLeft {
			base = 1
		}
		if base != data.resolvedParLevel {
			t.Fatalf("line %d: paragraph level = %d, want %d", data.line, base, data.resolvedParLevel)
		}
		runs, err := analysis.Line(0, len(data.codePoints))
		if err != nil {
			t.Fatalf("line %d: %v", data.line, err)
		}
		levels := make([]Level, len(data.codePoints))
		for _, run := range runs {
			for i := run.Start; i < run.End; i++ {
				levels[i] = run.Level
			}
		}
		for i, expected := range data.expectedLevels {
			if expected != -1 && levels[i] != expected {
				t.Fatalf("line %d: level[%d] = %d, want %d", data.line, i, levels[i], expected)
			}
		}
		indices := lineIndices(t, runs, 0, len(data.codePoints), data.expectedLevels)
		if !reflect.DeepEqual(indices, data.visualOrdering) {
			t.Fatalf("line %d: visual indices = %v, want %v", data.line, indices, data.visualOrdering)
		}
	}
}

func parseLevelsLine(line string) ([]Level, error) {
	line = strings.TrimPrefix(line, "@Levels:")
	return parseLevels(line)
}

func parseReorderLine(line string) ([]int, error) {
	line = strings.TrimPrefix(line, "@Reorder:")
	return parseOrdering(line)
}

func parseCharsLine(line string) (oneBidiData, error) {
	fields := strings.Split(line, ";")
	if len(fields) != 2 {
		return oneBidiData{}, fmt.Errorf("invalid line: %s", line)
	}
	var err error
	chars := strings.Fields(fields[0])
	out := make([]rune, len(chars))
	for i, cs := range chars {
		r, ok := runesForClasses[cs]
		if !ok {
			return oneBidiData{}, fmt.Errorf("unsupported class %s", cs)
		}
		out[i] = r
	}
	baseDirFlags, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	return oneBidiData{out, baseDirFlags}, err
}

type oneBidiData struct {
	runes       []rune
	baseDirFlag int
}

type bidiTest struct {
	ltor   []int
	levels []Level
	data   []oneBidiData
}

func parseBidiTests() ([]bidiTest, error) {
	var (
		out     []bidiTest
		current bidiTest
	)
	for lineNumber, lineB := range bytes.Split(bidiTestSrc, []byte{'\n'}) {
		line := string(bytes.TrimSpace(lineB))
		if len(line) == 0 || line[0] == '#' {
			// flush the current datas
			if len(current.data) != 0 {
				out = append(out, current)
				current.data = nil
			}
			continue
		}

		var err error
		if strings.HasPrefix(line, "@Reorder:") {
			current.ltor, err = parseReorderLine(line)
			if err != nil {
				return nil, fmt.Errorf("invalid  line %d: %s", lineNumber+1, err)
			}
			continue
		} else if strings.HasPrefix(line, "@Levels:") {
			current.levels, err = parseLevelsLine(line)
			if err != nil {
				return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
			}
			continue
		}

		/* Test line */
		lineData, err := parseCharsLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
		}
		current.data = append(current.data, lineData)
	}
	return out, nil
}

func runOneComplexBidi(paragraph *Paragraph, data bidiTest) (levelsList [][]Level) {
	for _, line := range data.data {
		for baseDirMode := 0; baseDirMode < 3; baseDirMode++ {

			if (line.baseDirFlag & (1 << baseDirMode)) == 0 {
				continue
			}

			var defaultDirection Direction
			switch baseDirMode {
			case 0:
				defaultDirection = Neutral
			case 1:
				defaultDirection = LeftToRight
			case 2:
				defaultDirection = RightToLeft
			}

			levels := paragraph.Segment(line.runes, defaultDirection).levels
			levelsList = append(levelsList, levels)
		}
	}
	return
}

// Unicode BidiTest.txt uses class instead of runes as input :
// use this map to create a compatible text
var runesForClasses = map[string]rune{
	"L":   '\u0061',
	"R":   '\u05d0',
	"EN":  '\u0030',
	"ES":  '\u002B',
	"ET":  '\u0023',
	"AN":  '\u0661',
	"CS":  '\u002E',
	"B":   '\u000A',
	"S":   '\u000B',
	"WS":  '\u0020',
	"ON":  '\u0021',
	"BN":  '\u0000',
	"NSM": '\u0300',
	"AL":  '\u0608',
	"LRO": '\u202D',
	"RLO": '\u202e',
	"LRE": '\u202A',
	"RLE": '\u202B',
	"PDF": '\u202C',
	"LRI": '\u2066',
	"RLI": '\u2067',
	"FSI": '\u2068',
	"PDI": '\u2069',
}

func TestBidi(t *testing.T) {
	datas, err := parseBidiTests()
	tu.AssertNoErr(t, err)

	for index, data := range datas {
		/* Test it */
		levelsList := runOneComplexBidi(&Paragraph{}, data)

		/* Compare */
		for j := range levelsList {
			levels := levelsList[j]

			for i, level := range levels {
				if exp := data.levels[i]; level != exp && exp != -1 {
					t.Fatalf("failure on test %d: levels[%d]: expected %d, got %d", index+1, i, exp, level)
					break
				}
			}
		}
	}
}

func BenchmarkSingleDirection(b *testing.B) {
	var paragraph Paragraph

	fullLTR := []rune(strings.Repeat("A sample tesxt with some digits 7 : 8 9.", 100))
	// fullRTL := []rune(strings.Repeat("דהודהודהודהודהודהודהודהודהו דהודהו דהודהו דהו דהו", 100))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = paragraph.Segment(fullLTR, Neutral)
		_ = paragraph.Segment(fullLTR, LeftToRight)
		// _ = paragraph.Segment(fullRTL, Neutral)
		// _ = paragraph.Segment(fullRTL, RightToLeft)
	}
}

func BenchmarkSimple(b *testing.B) {
	datas, err := parseBidiCharacterTests()
	tu.AssertNoErr(b, err)

	var paragraph Paragraph
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, test := range datas {
			_ = paragraph.Segment(test.codePoints, test.parDir)
		}
	}
}

func BenchmarkComplex(b *testing.B) {
	datas, err := parseBidiTests()
	tu.AssertNoErr(b, err)

	var paragraph Paragraph

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, lineData := range datas {
			runOneComplexBidi(&paragraph, lineData)
		}
	}
}

func BenchmarkPathological(b *testing.B) {
	for _, tc := range []struct {
		name, unit string
	}{
		{"brackets", "()"},
		{"digits", "1"},
		{"unmatchedIsolates", "\u2066"},
		{"unmatchedFSI", "\u2068"},
	} {
		for _, n := range []int{20000, 40000} {
			text := []rune(strings.Repeat(tc.unit, n))
			b.Run(fmt.Sprintf("%s/%d", tc.name, n), func(b *testing.B) {
				var p Paragraph
				for i := 0; i < b.N; i++ {
					p.Segment(text, RightToLeft)
				}
			})
		}
	}
}

func TestRunsPreserveEmbeddingLevels(t *testing.T) {
	var p Paragraph
	runs := p.SegmentString("a\u202abא\u202cב", LeftToRight)
	want := []Run{{0, 2, 0}, {2, 3, 2}, {3, 5, 3}, {5, 6, 1}}
	if runs.NumRuns() != len(want) {
		t.Fatalf("got %d runs, want %d", runs.NumRuns(), len(want))
	}
	for i, expected := range want {
		if got := runs.Run(i); got != expected {
			t.Errorf("run %d: got %+v, want %+v", i, got, expected)
		}
	}
}

func TestAnalysisLogicalLevels(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text      string
		direction Direction
		want      []Level
	}{
		{"abאב12cd", Neutral, []Level{0, 0, 1, 1, 2, 2, 0, 0}},
		{"a\u2067אב 12\u2069z", Neutral, []Level{0, 0, 1, 1, 1, 2, 2, 0, 0}},
		{"abc", RightToLeft, []Level{2, 2, 2}},
		{"אב", Neutral, []Level{1, 1}},
		{"abc\n", Neutral, []Level{0, 0, 0, 0}},
		{"a\u202bאב  ", Neutral, []Level{0, 0, 1, 1, 1, 1}},
		{"", Neutral, nil},
	} {
		t.Run(tt.text, func(t *testing.T) {
			analysis, err := Analyze([]rune(tt.text), tt.direction)
			if err != nil {
				t.Fatal(err)
			}
			var got []Level
			end := 0
			for _, run := range analysis.LogicalRuns() {
				if run.Start != end || run.End <= run.Start {
					t.Fatalf("invalid logical range: %+v after %d", run, end)
				}
				for i := run.Start; i < run.End; i++ {
					got = append(got, run.Level)
				}
				end = run.End
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("logical levels = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalysisLine(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text       string
		start, end int
		want       []int
	}{
		{"abאב12cd", 0, 8, []int{0, 1, 4, 5, 3, 2, 6, 7}},
		{"a\u2067אב 12\u2069z", 0, 9, []int{0, 1, 5, 6, 4, 3, 2, 7, 8}},
		{"a\u202bאב  \u202cz", 0, 6, []int{0, 1, 3, 2, 4, 5}},
		{"a\u202bאב  \u202cz", 2, 6, []int{3, 2, 4, 5}},
		{"אב  ", 0, 4, []int{3, 2, 1, 0}},
		{"", 0, 0, nil},
		{"abאב12cd", 4, 4, nil},
	} {
		t.Run(fmt.Sprintf("%q/%d:%d", tt.text, tt.start, tt.end), func(t *testing.T) {
			analysis, err := Analyze([]rune(tt.text), Neutral)
			if err != nil {
				t.Fatal(err)
			}
			runs, err := analysis.Line(tt.start, tt.end)
			if err != nil {
				t.Fatal(err)
			}
			got := lineIndices(t, runs, tt.start, tt.end, nil)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("visual indices = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalysisLineInvalidRange(t *testing.T) {
	analysis, err := Analyze([]rune("abc"), Neutral)
	if err != nil {
		t.Fatal(err)
	}
	for _, span := range [][2]int{{-1, 1}, {0, 4}, {2, 1}, {4, 4}} {
		if _, err := analysis.Line(span[0], span[1]); err == nil {
			t.Errorf("Line(%d, %d) accepted invalid range", span[0], span[1])
		}
	}
}

func lineIndices(t *testing.T, runs []Run, start, end int, ignored []Level) []int {
	t.Helper()
	seen := make([]bool, end-start)
	var indices []int
	for _, run := range runs {
		if run.Start < start || run.End > end || run.Start >= run.End {
			t.Fatalf("run %+v outside [%d,%d)", run, start, end)
		}
		for n := 0; n < run.End-run.Start; n++ {
			i := run.Start + n
			if !run.IsLeftToRight() {
				i = run.End - 1 - n
			}
			if seen[i-start] {
				t.Fatalf("duplicate rune %d", i)
			}
			seen[i-start] = true
			if ignored == nil || ignored[i] != -1 {
				indices = append(indices, i)
			}
		}
	}
	for i, present := range seen {
		if !present {
			t.Fatalf("missing rune %d", start+i)
		}
	}
	return indices
}

func TestAnalysisDetached(t *testing.T) {
	t.Parallel()
	text := []rune("abאב12cd")
	analysis, err := Analyze(text, Neutral)
	if err != nil {
		t.Fatal(err)
	}
	want := analysis.LogicalRuns()
	line, err := analysis.Line(0, len(text))
	if err != nil {
		t.Fatal(err)
	}
	wantLine := append([]Run(nil), line...)
	line[0].Start = 99
	for i := range text {
		text[i] = 'z'
	}
	runs := analysis.LogicalRuns()
	runs[0].Level = 9
	if got := analysis.LogicalRuns(); !reflect.DeepEqual(got, want) {
		t.Fatalf("input/output mutation changed analysis: %v, want %v", got, want)
	}
	var wg sync.WaitGroup
	for n := 0; n < 8; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if _, err := Analyze([]rune("אב"), Neutral); err != nil {
					t.Error(err)
				}
				if got := analysis.LogicalRuns(); !reflect.DeepEqual(got, want) {
					t.Errorf("concurrent analysis changed result: %v", got)
				}
				if _, err := analysis.Line(2, 6); err != nil {
					t.Error(err)
				}
				got, err := analysis.Line(0, 8)
				if err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(got, wantLine) {
					t.Errorf("line mutation or concurrent reads changed result: %v, want %v", got, wantLine)
				}
			}
		}()
	}
	wg.Wait()
}

func TestAnalysisLineLeavesParagraphLevels(t *testing.T) {
	analysis, err := Analyze([]rune("a\u202bאב  "), Neutral)
	if err != nil {
		t.Fatal(err)
	}
	logical := analysis.LogicalRuns()
	line, err := analysis.Line(0, 6)
	if err != nil {
		t.Fatal(err)
	}
	for _, run := range line {
		if run.Start <= 4 && run.End > 4 && run.Level != 0 {
			t.Fatalf("line trailing space level = %d, want 0", run.Level)
		}
	}
	if got := analysis.LogicalRuns(); !reflect.DeepEqual(got, logical) {
		t.Fatalf("Line changed paragraph levels: %v, want %v", got, logical)
	}
}

func ExampleAnalyze() {
	analysis, err := Analyze([]rune("abאב12cd"), Neutral)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Reuse the paragraph analysis after the caller selects a line boundary.
	runs, err := analysis.Line(0, 8)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, run := range runs {
		fmt.Printf("[%d,%d) level %d\n", run.Start, run.End, run.Level)
	}
	// Output:
	// [0,2) level 0
	// [4,6) level 2
	// [2,4) level 1
	// [6,8) level 0
}

func BenchmarkAnalysisParagraphs(b *testing.B) {
	for _, n := range []int{128, 256} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			text := []rune(strings.Repeat("a\n", n))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for at := 0; at < len(text); {
					analysis, err := Analyze(text[at:], Neutral)
					if err != nil {
						b.Fatal(err)
					}
					at += analysis.Len()
				}
			}
		})
	}
}
