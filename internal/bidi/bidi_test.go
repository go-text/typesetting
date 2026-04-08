package bidi

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-text/typesetting/internal/unicodedata"
)

// TODO: investigate https://github.com/golang/go/issues/69819

// Test copied from https://github.com/golang/go/issues/71809
func TestN2(t *testing.T) {
	str := `ع a`
	p := Paragraph{}
	p.SetString(str, DefaultDirection(LeftToRight))
	order, err := p.Order()
	if err != nil {
		log.Fatal(err)
	}

	expectedRuns := []runInformation{
		{"ع", RightToLeft, 0, 0},
		{" a", LeftToRight, 1, 2},
	}

	if nr, want := order.NumRuns(), len(expectedRuns); nr != want {
		t.Errorf("order.NumRuns() = %d; want %d", nr, want)
	}

	for i, want := range expectedRuns {
		r := order.Run(i)
		if got := r.String(); got != want.str {
			t.Errorf("Run(%d) = %q; want %q", i, got, want.str)
		}
		if s, e := r.Pos(); s != want.start || e != want.end {
			t.Errorf("Run(%d).start = %d, .end = %d; want start = %d, end = %d", i, s, e, want.start, want.end)
		}
		if d := r.Direction(); d != want.dir {
			t.Errorf("Run(%d).Direction = %d; want %d", i, d, want.dir)
		}
	}
}

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

func parseLevels(line string) ([]level, error) {
	fields := strings.Fields(line)
	out := make([]level, len(fields))
	for i, f := range fields {
		if f == "x" {
			out[i] = -1
		} else {
			lev, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("invalid level %s: %s", f, err)
			}
			out[i] = level(lev)
		}
	}
	return out, nil
}

type testData struct {
	line             int // just for easier debugging
	codePoints       []rune
	expectedLevels   []level
	visualOrdering   []int
	parDir           int
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
	out.parDir, err = strconv.Atoi(fields[1])
	if err != nil {
		return out, fmt.Errorf("invalid paragraph direction %s: %s", fields[1], err)
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

func runOneSimpleBidi(lineData testData) []Level {
	var defaultDirection DefaultDirection
	switch lineData.parDir {
	case 0:
		defaultDirection = LeftToRight
	case 1:
		defaultDirection = RightToLeft
	case 2:
		defaultDirection = Neutral
	}

	levels := (&Paragraph{}).Segment(lineData.codePoints, defaultDirection).levels

	// ltor := make([]int, len(lineData.codePoints))
	// for i := range ltor {
	// 	ltor[i] = i
	// }

	// ReorderLine(0 /*FRIBIDI_FLAG_REORDER_NSM*/, types, len(types), 0, baseDir, levels, nil, ltor)

	// j := 0
	// for _, lr := range ltor {
	// 	if !types[lr].isExplicitOrBn() {
	// 		ltor[j] = lr
	// 		j++
	// 	}
	// }
	// ltor = ltor[0:j] // slice to length

	return levels
}

func TestBidiCharacters(t *testing.T) {
	ti := time.Now()

	datas, err := parseBidiCharacterTests()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("input data parsed in", time.Since(ti))
	ti = time.Now()

	for _, test := range datas {
		levels := runOneSimpleBidi(test)

		/* Compare */
		for i, level := range levels {
			if exp := test.expectedLevels[i]; level != exp && exp != -1 {
				t.Fatalf("failure at line %d: levels[%d]: expected %d, got %d", test.line, i, exp, level)
				break
			}
		}

		// if len(lineData.visualOrdering) != len(ltor) {
		// 	t.Fatalf("failure visual ordering: got %v, expected %v", ltor, lineData.visualOrdering)
		// }
		// for i := range ltor {
		// 	if lineData.visualOrdering[i] != ltor[i] {
		// 		t.Fatalf("failure visual ordering: got %v, expected %v", ltor, lineData.visualOrdering)
		// 	}
		// }
	}

	fmt.Println("test run in", time.Since(ti))
}

func parseCharType(s string) (charType, error) {
	switch s {
	case "L":
		return unicodedata.BD_L, nil
	case "R":
		return unicodedata.BD_R, nil
	case "AL":
		return unicodedata.BD_AL, nil
	case "EN":
		return unicodedata.BD_EN, nil
	case "AN":
		return unicodedata.BD_AN, nil
	case "ES":
		return unicodedata.BD_ES, nil
	case "ET":
		return unicodedata.BD_ET, nil
	case "CS":
		return unicodedata.BD_CS, nil
	case "NSM":
		return unicodedata.BD_NSM, nil
	case "BN":
		return unicodedata.BD_BN, nil
	case "B":
		return unicodedata.BD_B, nil
	case "S":
		return unicodedata.BD_S, nil
	case "WS":
		return unicodedata.BD_WS, nil
	case "ON":
		return unicodedata.BD_ON, nil
	case "LRE":
		return unicodedata.BD_LRE, nil
	case "RLE":
		return unicodedata.BD_RLE, nil
	case "LRO":
		return unicodedata.BD_LRO, nil
	case "RLO":
		return unicodedata.BD_RLO, nil
	case "PDF":
		return unicodedata.BD_PDF, nil
	case "LRI":
		return unicodedata.BD_LRI, nil
	case "RLI":
		return unicodedata.BD_RLI, nil
	case "FSI":
		return unicodedata.BD_FSI, nil
	case "PDI":
		return unicodedata.BD_PDI, nil
	default:
		return 0, fmt.Errorf("invalid char type %s", s)
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

func parseCharsLine(line string) ([]charType, int, error) {
	fields := strings.Split(line, ";")
	if len(fields) != 2 {
		return nil, 0, fmt.Errorf("invalid line: %s", line)
	}
	var err error
	chars := strings.Fields(fields[0])
	out := make([]charType, len(chars))
	for i, cs := range chars {
		out[i], err = parseCharType(cs)
		if err != nil {
			return nil, 0, err
		}
	}
	baseDirFlags, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	return out, baseDirFlags, err
}

type oneBidiData struct {
	types       []charType
	baseDirFlag int
}

type bidiTest struct {
	ltor   []int
	levels []Level
	data   []oneBidiData
}

func parseBidiTests() ([]bidiTest, error) {
	const filename = "test/unicode-conformance/BidiTest.txt"

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
		var lineData oneBidiData
		lineData.types, lineData.baseDirFlag, err = parseCharsLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid line %d: %s", lineNumber+1, err)
		}
		current.data = append(current.data, lineData)
	}
	return out, nil
}

// func runOneComplexBidi(data bidiTest) (levelsList [][]Level) {
// 	for _, line := range data.data {
// 		for baseDirMode := 0; baseDirMode < 3; baseDirMode++ {

// 			if (line.baseDirFlag & (1 << baseDirMode)) == 0 {
// 				continue
// 			}

// 			var baseDir ParType
// 			switch baseDirMode {
// 			case 0:
// 				baseDir = ucd.BD_ON
// 			case 1:
// 				baseDir = ucd.BD_L
// 			case 2:
// 				baseDir = ucd.BD_R
// 			}

// 			// Brackets are not used in the BidiTest.txt file
// 			levels, _ := GetParEmbeddingLevels(line.types, nil, &baseDir)

// 			ltor := make([]int, len(levels))
// 			for i := range ltor {
// 				ltor[i] = i
// 			}

// 			ReorderLine(0 /*FRIBIDI_FLAG_REORDER_NSM*/, line.types, len(line.types),
// 				0, baseDir, levels,
// 				nil, ltor)

// 			j := 0
// 			for _, lr := range ltor {
// 				if !line.types[lr].isExplicitOrBn() {
// 					ltor[j] = lr
// 					j++
// 				}
// 			}
// 			ltor = ltor[0:j] // slice to length

// 			levelsList = append(levelsList, levels)
// 			ltorList = append(ltorList, ltor)
// 		}
// 	}
// 	return
// }

// func TestBidi(t *testing.T) {
// 	ti := time.Now()
// 	datas, err := parseBidiTests()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	fmt.Println("parsed BidiTest.txt in", time.Since(ti))
// 	ti = time.Now()

// 	for index, data := range datas {
// 		/* Test it */
// 		levelsList, ltorList := runOneComplexBidi(data)

// 		/* Compare */
// 		for j := range levelsList {
// 			levels, ltor := levelsList[j], ltorList[j]

// 			for i, level := range levels {
// 				if exp := data.levels[i]; level != exp && exp != -1 {
// 					t.Fatalf("failure on test %d: levels[%d]: expected %d, got %d", index+1, i, exp, level)
// 					break
// 				}
// 			}

// 			if len(data.ltor) != len(ltor) {
// 				t.Fatalf("failure on test %d: visual ordering: got %v, expected %v", index+1, ltor, data.ltor)
// 			}
// 			for i := range ltor {
// 				if data.ltor[i] != ltor[i] {
// 					t.Fatalf("failure on test %d: visual ordering: got %v, expected %v", index+1, ltor, data.ltor)
// 				}
// 			}
// 		}
// 	}

// 	fmt.Println("test run in", time.Since(ti))
// }

func BenchmarkSimple(b *testing.B) {
	datas, err := parseBidiCharacterTests()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, lineData := range datas {
			runOneSimpleBidi(lineData)
		}
	}
}

// func BenchmarkComplex(b *testing.B) {
// 	datas, err := parseBidiTests()
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		for _, lineData := range datas {
// 			runOneComplexBidi(lineData)
// 		}
// 	}
// }
