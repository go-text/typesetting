// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package segmenter

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

func hex(rs []rune) string {
	out := ""
	for _, r := range rs {
		out += fmt.Sprintf(" 0x%X", r)
	}
	return out[1:]
}

func collectLines(s *Segmenter, input []rune) []string {
	s.Init(input)
	iter := s.LineIterator()
	var out []string
	for iter.Next() {
		out = append(out, string(iter.Line().Text))
	}
	return out
}

func collectGraphemes(s *Segmenter, input []rune) []string {
	s.Init(input)
	iter := s.GraphemeIterator()
	var out []string
	for iter.Next() {
		out = append(out, string(iter.Grapheme().Text))
	}
	return out
}

func collectWords(s *Segmenter, input []rune) []string {
	s.Init(input)
	iter := s.WordIterator()
	var out []string
	for iter.Next() {
		out = append(out, string(iter.Word().Text))
	}
	return out
}

func collectWordBoundaries(s *Segmenter, input []rune) []bool {
	s.Init(input)
	out := make([]bool, len(s.attributes))
	for i, a := range s.attributes {
		out[i] = a&wordBoundary != 0
	}
	return out
}

func TestLineBreakUnicodeReference(t *testing.T) {
	file := "test/LineBreakTest.txt"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	var seg1 Segmenter
	for i, line := range lines {
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		s, expectedSegments := parseUCDTestLine(t, line)
		text := []rune(s)
		actualSegments := collectLines(&seg1, text)
		if !reflect.DeepEqual(expectedSegments, actualSegments) {
			t.Errorf("line %d [%s]: expected %s, got %s", i+1, hex(text), expectedSegments, actualSegments)
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

func parseUCDTestLine(t *testing.T, line string) (string, []string) {
	var segments []string
	var segmentStart int

	runes, boundaries := parseUCDTestLineBoundary(t, line)
	for i, b := range boundaries {
		if i == 0 { // do not add empty segment at the start
			continue
		}
		if b {
			// boundary here
			segments = append(segments, string(runes[segmentStart:i]))
			segmentStart = i
		}
	}

	return string(runes), segments
}

func TestGraphemeBreakUnicodeReference(t *testing.T) {
	file := "test/GraphemeBreakTest.txt"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	var seg1 Segmenter
	for i, line := range lines {
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		s, expectedSegments := parseUCDTestLine(t, line)
		text := []rune(s)
		actualSegments := collectGraphemes(&seg1, text)
		if !reflect.DeepEqual(expectedSegments, actualSegments) {
			t.Errorf("line %d [%s]: expected %#v, got %#v", i+1, hex(text), expectedSegments, actualSegments)
		}
	}
}

func TestWordBreakUnicodeReference(t *testing.T) {
	file := "test/WordBreakTest.txt"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")

	var seg1 Segmenter
	for i, line := range lines {
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		text, expectedBoundaries := parseUCDTestLineBoundary(t, line)
		actualBoundaries := collectWordBoundaries(&seg1, text)
		if !reflect.DeepEqual(expectedBoundaries, actualBoundaries) {
			t.Errorf("line %d [%s]: expected %#v, got %#v", i+1, hex(text), expectedBoundaries, actualBoundaries)
		}
	}
}

func TestWordSegmenter(t *testing.T) {
	var seg Segmenter
	for _, test := range []struct {
		input string
		words []string
	}{
		{"My name is Cris", []string{"My", "name", "is", "Cris"}},
		{"Je m'appelle Benoit.", []string{"Je", "m'appelle", "Benoit"}},
		{"Hi : nice ?! suit !", []string{"Hi", "nice", "suit"}},
	} {
		got := collectWords(&seg, []rune(test.input))
		if !reflect.DeepEqual(test.words, got) {
			t.Errorf("for %s, expected %v, got %v", test.input, test.words, got)
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
	by, err := ioutil.ReadFile(file)
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
