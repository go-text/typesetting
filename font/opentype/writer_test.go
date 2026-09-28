package opentype

import (
	"bytes"
	"encoding/binary"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

func TestWrite(t *testing.T) {
	for _, filename := range tu.Filenames(t, "common") {
		f, err := td.Files.ReadFile(filename)
		tu.AssertNoErr(t, err)

		font, err := NewLoader(bytes.NewReader(f))
		tu.AssertNoErr(t, err)

		tags := font.Tables()
		tables := make([]Table, len(tags))
		for i, tag := range tags {
			tables[i].Tag = tag
			tables[i].Content, err = font.RawTable(tag)
			tu.AssertNoErr(t, err)
		}

		content := WriteTTF(tables)
		font2, err := NewLoader(bytes.NewReader(content))
		tu.AssertNoErr(t, err)

		for _, table := range tables {
			t2, err := font2.RawTable(table.Tag)
			tu.AssertNoErr(t, err)

			tu.Assert(t, bytes.Equal(table.Content, t2))
		}
	}
}

func TestWriteTTFPartialWordChecksum(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  []byte
		want uint32
	}{
		{"one", []byte{1}, 0x01000000},
		{"two", []byte{1, 2}, 0x01020000},
		{"three", []byte{1, 2, 3}, 0x01020300},
		{"four", []byte{1, 2, 3, 4}, 0x01020304},
		{"five", []byte{1, 2, 3, 4, 5}, 0x06020304},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A table can share backing storage with the next table.
			backing := append(append([]byte(nil), tc.src...), 0xaa, 0xbb, 0xcc)
			before := append([]byte(nil), backing...)
			out := WriteTTF([]Table{{Tag: MustNewTag("test"), Content: backing[:len(tc.src)]}})
			if got := binary.BigEndian.Uint32(out[16:]); got != tc.want {
				t.Errorf("checksum=%08x, want %08x", got, tc.want)
			}
			if !bytes.Equal(backing, before) {
				t.Fatalf("checksum modified input backing storage: %x", backing)
			}
		})
	}
}
