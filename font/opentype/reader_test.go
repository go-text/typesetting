// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package opentype

import (
	"bytes"
	"math/rand"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

func TestParseCrashers(t *testing.T) {
	font, err := NewLoader(bytes.NewReader([]byte{}))
	tu.Assert(t, font == nil)
	tu.Assert(t, err != nil)

	for range [50]int{} {
		L := rand.Intn(100)
		input := make([]byte, L)
		rand.Read(input)

		_, err = NewLoader(bytes.NewReader(input))
		tu.Assert(t, err != nil)

		_, err = NewLoaders(bytes.NewReader(input))
		tu.Assert(t, err != nil)
	}
}

func TestCollection(t *testing.T) {
	for _, filename := range tu.Filenames(t, "collections") {
		f, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)

		fonts, err := NewLoaders(bytes.NewReader(f))
		tu.AssertC(t, err == nil, filename)

		for _, font := range fonts {
			tu.Assert(t, len(font.tables) != 0)
		}

		// check that NewLoader indeed fail on collections
		_, err = NewLoader(bytes.NewReader(f))
		tu.Assert(t, err != nil)
	}

	// check that it also works for single font files
	for _, filename := range tu.Filenames(t, "common") {
		f, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)

		fonts, err := NewLoaders(bytes.NewReader(f))
		tu.AssertC(t, err == nil, filename)

		if len(fonts) != 1 {
			tu.Assert(t, len(fonts) == 1)
		}
	}
}

func TestRawTable(t *testing.T) {
	for _, filename := range tu.Filenames(t, "common") {
		f, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)

		font, err := NewLoader(bytes.NewReader(f))
		tu.AssertC(t, err == nil, filename)

		_, err = font.RawTable(MustNewTag("xxxx"))
		tu.Assert(t, err != nil)

		_, err = font.RawTable(MustNewTag("head"))
		tu.AssertC(t, err == nil, filename)

		_, err = font.RawTable(MustNewTag("OS/2"))
		tu.AssertC(t, err == nil, filename)
	}
}

// shortReader returns at most one byte per Read call
type shortReader struct{ *bytes.Reader }

func (s shortReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return s.Reader.Read(p)
}

func TestShortReads(t *testing.T) {
	for _, file := range td.WithOTLayout[:1] {
		content, err := td.Files.ReadFile(file)
		tu.AssertNoErr(t, err)
		short, err := NewLoader(shortReader{bytes.NewReader(content)})
		tu.AssertNoErr(t, err)
		full, err := NewLoader(bytes.NewReader(content))
		tu.AssertNoErr(t, err)
		tu.Assert(t, len(short.tables) == len(full.tables))
	}
}
