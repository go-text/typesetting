// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package segmenter

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
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

func parseUCDTestLine(t *testing.T, line string) (string, []string) {
	var segments []string
	var input string
	var currentSegment string

	line = strings.Split(line, "#")[0] // remove comments
	for _, field := range strings.Fields(line) {
		switch field {
		case string(rune(0x00f7)): // DIVISION SIGN: boundary here
			// do not add empty segments when DIVISION SIGN is at the start
			if currentSegment != "" {
				segments = append(segments, currentSegment)
				currentSegment = ""
			}
		case string(rune(0x00d7)): // MULTIPLICATION SIGN: no boundary here
		default: // read the rune hex code
			character, err := strconv.ParseUint(field, 16, 32)
			if err != nil {
				t.Fatalf("invalid line %s: %s", line, err)
			}
			if character > 0x10ffff {
				t.Fatalf("unexpected character")
			}
			currentSegment += string(rune(character))
			input += string(rune(character))
		}
	}
	return input, segments
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

func segmentCount(s *Segmenter, input []rune) int {
	s.Init(input)
	iter := s.LineIterator()
	var out int
	for iter.Next() {
		out++
	}
	return out
}

func getInputs() [][]rune {
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
	inputs := getInputs()
	b.ResetTimer()

	seg := &Segmenter{}
	for i := 0; i < b.N; i++ {
		for _, line := range inputs {
			segmentCount(seg, line)
		}
	}
}
