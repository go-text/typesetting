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
		rand.New(rand.NewSource(int64(L))).Read(input)

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

// fontFiles returns the single fonts and collections used by the
// NewLoadersFromBytes tests. "common" includes a compressed WOFF file.
func fontFiles(t *testing.T) []string {
	return append(tu.Filenames(t, "common"), tu.Filenames(t, "collections")...)
}

func TestNewLoadersFromBytes(t *testing.T) {
	for _, filename := range fontFiles(t) {
		content, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)

		want, err := NewLoaders(bytes.NewReader(content))
		tu.AssertC(t, err == nil, filename)
		got, err := NewLoadersFromBytes(content)
		tu.AssertC(t, err == nil, filename)
		tu.AssertC(t, len(got) == len(want), filename)

		// every table must read back exactly as the copying loader reads it
		for i := range want {
			tu.AssertC(t, got[i].Type == want[i].Type, filename)
			tags := want[i].Tables()
			tu.AssertC(t, len(got[i].Tables()) == len(tags), filename)
			for _, tag := range tags {
				w, err := want[i].RawTable(tag)
				tu.AssertC(t, err == nil, filename)
				g, err := got[i].RawTable(tag)
				tu.AssertC(t, err == nil, filename)
				tu.AssertC(t, bytes.Equal(g, w), filename+" "+tag.String())
			}
		}
	}
}

func TestNewLoadersFromBytesAliasesInput(t *testing.T) {
	for _, filename := range fontFiles(t) {
		content, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)
		lds, err := NewLoadersFromBytes(content)
		tu.AssertNoErr(t, err)

		for _, ld := range lds {
			for tag, s := range ld.tables {
				if s.length == 0 || s.length < s.zLength {
					continue // empty or compressed: there is nothing to alias
				}
				raw, err := ld.RawTable(tag)
				tu.AssertNoErr(t, err)
				// the table is a view of content, not a copy
				tu.AssertC(t, &raw[0] == &content[s.offset], filename+" "+tag.String())
				// and its capacity stops at the table end, so an append by the
				// caller can not overwrite the next table
				tu.AssertC(t, cap(raw) == len(raw), filename+" "+tag.String())
			}
		}
	}
}

// A compressed WOFF table is decoded into a buffer. When dst is a view of the
// input (a table returned before), the decoder must not write into it, even
// when dst is large enough to hold the decoded table.
func TestNewLoadersFromBytesCompressedIgnoresDst(t *testing.T) {
	var compressed int
	for _, filename := range tu.Filenames(t, "common") {
		content, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)
		input := append([]byte(nil), content...)
		lds, err := NewLoadersFromBytes(input)
		tu.AssertNoErr(t, err)

		for _, ld := range lds {
			for tag, s := range ld.tables {
				if s.length == 0 || s.length >= s.zLength || int(s.zLength) > len(input) {
					continue
				}
				compressed++
				view := input[:s.zLength:s.zLength] // a view large enough to reuse
				got, err := ld.RawTableTo(tag, view)
				tu.AssertNoErr(t, err)
				tu.AssertC(t, &got[0] != &input[0], filename+" "+tag.String())
				tu.AssertC(t, bytes.Equal(input, content), filename+" "+tag.String())
			}
		}
	}
	tu.Assert(t, compressed > 0) // the WOFF file in "common" has compressed tables
}

// A loader built from bytes must accept and reject the same tables as the
// copying loader over the same (possibly truncated) bytes.
func TestNewLoadersFromBytesTruncated(t *testing.T) {
	var rejected int
	for _, filename := range fontFiles(t) {
		content, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)
		// keep the table directory, cut the table data
		for _, size := range []int{len(content) / 2, len(content) - 1} {
			input := content[:size]
			got, err := NewLoadersFromBytes(input)
			want, err2 := NewLoaders(bytes.NewReader(input))
			tu.AssertC(t, (err == nil) == (err2 == nil), filename)
			if err != nil {
				continue
			}
			for i, ld := range got {
				for _, tag := range ld.Tables() {
					g, err := ld.RawTable(tag)
					w, err2 := want[i].RawTable(tag)
					tu.AssertC(t, (err == nil) == (err2 == nil), filename+" "+tag.String())
					if err != nil {
						rejected++
						continue
					}
					tu.AssertC(t, bytes.Equal(g, w), filename+" "+tag.String())
				}
			}
		}
	}
	tu.Assert(t, rejected > 0) // tables past the cut must fail
}

// A zero-length table at the very end of the file is rejected, like the
// copying loader does with a bytes.Reader (ReadAt at EOF returns io.EOF).
func TestNewLoadersFromBytesEmptyTableAtEOF(t *testing.T) {
	data := []byte{1, 2, 3, 4}
	copying := Loader{file: bytes.NewReader(data)}
	fromBytes := Loader{file: bytes.NewReader(data), data: data}
	for _, s := range []tableSection{{offset: 4}, {offset: 5}, {offset: 3, length: 2}} {
		_, err := copying.findTableBuffer(s, nil)
		tu.Assert(t, err != nil)
		_, err = fromBytes.findTableBuffer(s, nil)
		tu.Assert(t, err != nil)
	}
	for _, s := range []tableSection{{offset: 3}, {offset: 0, length: 4}} {
		_, err := copying.findTableBuffer(s, nil)
		tu.AssertNoErr(t, err)
		_, err = fromBytes.findTableBuffer(s, nil)
		tu.AssertNoErr(t, err)
	}
}

// Random input is covered by TestParseCrashers: NewLoadersFromBytes parses
// the header with NewLoaders.
func TestNewLoadersFromBytesEmpty(t *testing.T) {
	_, err := NewLoadersFromBytes(nil)
	tu.Assert(t, err != nil)
}
