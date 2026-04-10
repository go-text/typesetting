package bidi

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

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
	str := "Hellö"
	runs := (&Paragraph{}).Segment([]rune(str), Neutral)

	expectedRuns := []Run{
		{0, 5, 0},
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
	str := "Hello\nworld"
	runs := (&Paragraph{}).Segment([]rune(str), Neutral)

	// 5 is the length up to and excluding the \n
	tu.Assert(t, runs.Run(0).End == 5)
}

func TestDoubleSetString(t *testing.T) {
	str := "العاشر ليونيكود (Unicode Conference)،"
	var p Paragraph
	_ = p.Segment([]rune(str), Neutral)
	_ = p.Segment([]rune(str), Neutral)
}

// ------------------------- Unicode conformance tests -------------------------

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
	const filename = "test/BidiCharacterTest.txt"

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var out []testData
	for lineNumber, line := range bytes.Split(b, []byte{'\n'}) {
		if len(line) == 0 || line[0] == '#' || line[0] == '\n' {
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
	const filename = "test/BidiTest.txt"

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var (
		out     []bidiTest
		current bidiTest
	)
	for lineNumber, lineB := range bytes.Split(b, []byte{'\n'}) {
		line := string(lineB)
		if len(line) == 0 || line[0] == '#' {
			// flush the current datas
			if len(current.data) != 0 {
				out = append(out, current)
				current.data = nil
			}
			continue
		}

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
