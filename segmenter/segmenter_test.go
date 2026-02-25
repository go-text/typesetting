// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package segmenter

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

type initMode int

const (
	initModeRunes initMode = iota
	initModeString
	initModeBytes
	initModeMax
)

func hex(rs []rune) string {
	out := ""
	for _, r := range rs {
		out += fmt.Sprintf(" 0x%X", r)
	}
	return out[1:]
}

func collectLineBreaks(s *Segmenter) []int {
	iter := s.LineIterator()
	var out []int
	for iter.Next() {
		line := iter.Line()
		out = append(out, line.Offset+len(line.Text))
	}
	return out
}

func collectGraphemes(s *Segmenter) []int {
	iter := s.GraphemeIterator()
	var out []int
	for iter.Next() {
		line := iter.Grapheme()
		out = append(out, line.Offset+len(line.Text))
	}
	return out
}

func collectWords(s *Segmenter) []string {
	iter := s.WordIterator()
	var out []string
	for iter.Next() {
		out = append(out, string(iter.Word().Text))
	}
	return out
}

func collectWordBoundaries(s *Segmenter) []bool {
	out := make([]bool, len(s.attributes))
	for i, a := range s.attributes {
		out[i] = a&wordBoundary != 0
	}
	return out
}

func TestLineBreakUnicodeReference(t *testing.T) {
	file := "test/LineBreakTest.txt"
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	var seg1 Segmenter
	for mode := initModeRunes; mode < initModeMax; mode++ {
		for i, line := range lines {
			if len(line) == 0 || strings.HasPrefix(line, "#") {
				continue
			}
			s, expectedSegments := parseUCDTestLine(t, line)
			text := []rune(s)
			switch mode {
			case initModeRunes:
				seg1.Init(text)
			case initModeString:
				seg1.InitWithString(s)
			case initModeBytes:
				seg1.InitWithBytes([]byte(s))
			}
			actualSegments := collectLineBreaks(&seg1)
			if !reflect.DeepEqual(expectedSegments, actualSegments) {
				t.Fatalf("line %d [%s]: mode %d: expected breaks %v, got %v", i+1, hex(text), mode, expectedSegments, actualSegments)
			}
		}
	}
}

func parseUCDTestLineBoundary(t *testing.T, line string) (runes []rune, boundaries []bool) {
	line = strings.Split(line, "#")[0] // remove comments
	for _, field := range strings.Fields(line) {
		switch field {
		case string(rune(0x00f7)): // DIVISION SIGN: boundary here
			boundaries = append(boundaries, true)
		case string(rune(0x00d7)): // MULTIPLICATION SIGN: no boundary here
			boundaries = append(boundaries, false)
		default: // read the rune hex code
			character, err := strconv.ParseUint(field, 16, 32)
			tu.AssertNoErr(t, err)
			tu.Assert(t, character <= 0x10ffff)

			runes = append(runes, rune(character))
		}
	}

	tu.Assert(t, len(runes)+1 == len(boundaries))

	return
}

func parseUCDTestLine(t *testing.T, line string) (string, []int) {
	var breaks []int

	runes, boundaries := parseUCDTestLineBoundary(t, line)
	for i, b := range boundaries {
		if i == 0 { // do not add empty segment at the start
			continue
		}
		if b {
			// boundary here
			breaks = append(breaks, i)
		}
	}

	return string(runes), breaks
}

func TestGraphemeBreakUnicodeReference(t *testing.T) {
	file := "test/GraphemeBreakTest.txt"
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	var seg1 Segmenter
	for mode := initModeRunes; mode < initModeMax; mode++ {
		for i, line := range lines {
			if len(line) == 0 || strings.HasPrefix(line, "#") {
				continue
			}
			s, expectedSegments := parseUCDTestLine(t, line)
			text := []rune(s)
			switch mode {
			case initModeRunes:
				seg1.Init(text)
			case initModeString:
				seg1.InitWithString(s)
			case initModeBytes:
				seg1.InitWithBytes([]byte(s))
			}
			actualSegments := collectGraphemes(&seg1)
			if !reflect.DeepEqual(expectedSegments, actualSegments) {
				t.Fatalf("line %d [%s]: mode %d: expected %v, got %v", i+1, hex(text), mode, expectedSegments, actualSegments)
			}
		}
	}
}

func TestWordBreakUnicodeReference(t *testing.T) {
	file := "test/WordBreakTest.txt"
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	var seg1 Segmenter
	for mode := initMode(0); mode < initModeMax; mode++ {
		for i, line := range lines {
			if len(line) == 0 || strings.HasPrefix(line, "#") {
				continue
			}
			s, expectedBoundaries := parseUCDTestLineBoundary(t, line)
			text := []rune(s)
			switch mode {
			case initModeRunes:
				seg1.Init(text)
			case initModeString:
				seg1.InitWithString(string(s))
			case initModeBytes:
				seg1.InitWithBytes([]byte(string(s)))
			}
			actualBoundaries := collectWordBoundaries(&seg1)
			if !reflect.DeepEqual(expectedBoundaries, actualBoundaries) {
				t.Errorf("line %d [%s]: mode %d: expected %#v, got %#v", i+1, hex(text), mode, expectedBoundaries, actualBoundaries)
			}
		}
	}
}

func TestWordSegmenter(t *testing.T) {
	var seg Segmenter
	for mode := initMode(0); mode < initModeMax; mode++ {
		for _, test := range []struct {
			input string
			words []string
		}{
			{"My name is Cris", []string{"My", "name", "is", "Cris"}},
			{"Je m'appelle Benoit.", []string{"Je", "m'appelle", "Benoit"}},
			{"Hi : nice ?! suit !", []string{"Hi", "nice", "suit"}},
		} {
			switch mode {
			case initModeRunes:
				seg.Init([]rune(test.input))
			case initModeString:
				seg.InitWithString(test.input)
			case initModeBytes:
				seg.InitWithBytes([]byte(test.input))
			}
			got := collectWords(&seg)
			if !reflect.DeepEqual(test.words, got) {
				t.Errorf("for %s, mode %d, expected %v, got %v", test.input, mode, test.words, got)
			}
		}
	}
}

func lineSegmentCount(s *Segmenter, input []rune) int {
	s.Init(input)
	iter := s.LineIterator()
	var out int
	for iter.Next() {
		out++
	}
	return out
}

func getLineBreakInputs() [][]rune {
	file := "test/LineBreakTest.txt"
	by, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(by), "\n")
	var inputs [][]rune
	for _, li := range lines {
		inputs = append(inputs, []rune(li))
	}
	return inputs
}

func BenchmarkSegmentUnicodeReference(b *testing.B) {
	inputs := getLineBreakInputs()
	b.ResetTimer()

	seg := &Segmenter{}
	for i := 0; i < b.N; i++ {
		for _, line := range inputs {
			lineSegmentCount(seg, line)
		}
	}
}
