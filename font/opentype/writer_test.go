package opentype

import (
	"bytes"
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

		contentT := WriteOpentype(tables, TrueType)
		font2T, err := NewLoader(bytes.NewReader(contentT))
		tu.AssertNoErr(t, err)
		tu.Assert(t, font2T.Type == TrueType)

		for _, table := range tables {
			t2, err := font2T.RawTable(table.Tag)
			tu.AssertNoErr(t, err)

			tu.Assert(t, bytes.Equal(table.Content, t2))
		}

		contentO := WriteOpentype(tables, OpenType)
		font2O, err := NewLoader(bytes.NewReader(contentO))
		tu.AssertNoErr(t, err)
		tu.Assert(t, font2O.Type == OpenType)

		for _, table := range tables {
			t2, err := font2O.RawTable(table.Tag)
			tu.AssertNoErr(t, err)

			tu.Assert(t, bytes.Equal(table.Content, t2))
		}
	}
}
