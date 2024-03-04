package tables

import (
	"bytes"
	"strings"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	ot "github.com/go-text/typesetting/font/opentype"
	tu "github.com/go-text/typesetting/testutils"
)

func TestEncodings(t *testing.T) {
	utf16s := []struct {
		encoded []byte
		decoded string
	}{
		{[]byte{0, 66, 0, 111, 0, 108, 0, 100}, "Bold"},
		{[]byte{0, 82, 0, 111, 0, 98, 0, 111, 0, 116, 0, 111}, "Roboto"},
		{[]byte{0, 82, 0, 101, 0, 103, 0, 117, 0, 108, 0, 97, 0, 114}, "Regular"},
		{[]byte{0, 79, 0, 98, 0, 108, 0, 105, 0, 113, 0, 117, 0, 101}, "Oblique"},
		{[]byte{0, 84, 0, 101, 0, 115, 0, 116, 0, 32, 0, 84, 0, 84, 0, 70}, "Test TTF"},
		{[]byte{0, 79, 0, 112, 0, 101, 0, 110, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115}, "Open Sans"},
		{[]byte{0, 66, 0, 111, 0, 108, 0, 100, 0, 32, 0, 73, 0, 116, 0, 97, 0, 108, 0, 105, 0, 99}, "Bold Italic"},
		{[]byte{0, 66, 0, 111, 0, 108, 0, 100, 0, 32, 0, 79, 0, 98, 0, 108, 0, 105, 0, 113, 0, 117, 0, 101}, "Bold Oblique"},
		{[]byte{0, 82, 0, 97, 0, 108, 0, 101, 0, 119, 0, 97, 0, 121, 0, 45, 0, 118, 0, 52, 0, 48, 0, 50, 0, 48}, "Raleway-v4020"},
		{[]byte{0, 79, 0, 108, 0, 100, 0, 97, 0, 110, 0, 105, 0, 97, 0, 32, 0, 65, 0, 68, 0, 70, 0, 32, 0, 83, 0, 116, 0, 100}, "Oldania ADF Std"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 74, 0, 80}, "Noto Sans CJK JP"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 75, 0, 82}, "Noto Sans CJK KR"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 83, 0, 67}, "Noto Sans CJK SC"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 84, 0, 67}, "Noto Sans CJK TC"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 72, 0, 75}, "Noto Sans CJK HK"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 101, 0, 114, 0, 105, 0, 102, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 74, 0, 80}, "Noto Serif CJK JP"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 101, 0, 114, 0, 105, 0, 102, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 75, 0, 82}, "Noto Serif CJK KR"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 101, 0, 114, 0, 105, 0, 102, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 83, 0, 67}, "Noto Serif CJK SC"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 101, 0, 114, 0, 105, 0, 102, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 84, 0, 67}, "Noto Serif CJK TC"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 77, 0, 111, 0, 110, 0, 111, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 74, 0, 80}, "Noto Sans Mono CJK JP"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 77, 0, 111, 0, 110, 0, 111, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 75, 0, 82}, "Noto Sans Mono CJK KR"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 77, 0, 111, 0, 110, 0, 111, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 83, 0, 67}, "Noto Sans Mono CJK SC"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 77, 0, 111, 0, 110, 0, 111, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 84, 0, 67}, "Noto Sans Mono CJK TC"},
		{[]byte{0, 78, 0, 111, 0, 116, 0, 111, 0, 32, 0, 83, 0, 97, 0, 110, 0, 115, 0, 32, 0, 77, 0, 111, 0, 110, 0, 111, 0, 32, 0, 67, 0, 74, 0, 75, 0, 32, 0, 72, 0, 75}, "Noto Sans Mono CJK HK"},
	}
	for _, utf16 := range utf16s {
		tu.Assert(t, decodeUtf16(utf16.encoded) == utf16.decoded)
	}

	macs := []struct {
		encoded []byte
		decoded string
	}{
		{[]byte{67, 111, 117, 114, 105, 101, 114}, "Courier"},
		{[]byte{71, 101, 110, 101, 118, 97}, "Geneva"},
	}
	for _, mac := range macs {
		tu.Assert(t, DecodeMacintosh(mac.encoded) == mac.decoded)
	}

	tu.Assert(t, DecodeMacintoshByte(71) == 'G')
}

func TestFamilyNames(t *testing.T) {
	// macintosh encoding

	f, err := td.Files.ReadFile("collections/Courier.dfont")
	tu.AssertNoErr(t, err)

	fonts, err := ot.NewLoaders(bytes.NewReader(f))
	tu.AssertC(t, err == nil, "Courier")

	for _, font := range fonts {
		names, _, err := ParseName(readTable(t, font, "name"))
		tu.AssertNoErr(t, err)

		// NameFontFamily
		tu.Assert(t, names.Name(1) == "Courier")
	}

	// UTF16 encoding
	f, err = td.Files.ReadFile("collections/NotoSansCJK-Bold.ttc")
	tu.AssertNoErr(t, err)

	fonts, err = ot.NewLoaders(bytes.NewReader(f))
	tu.AssertC(t, err == nil, "NotoSansCJK")

	for _, font := range fonts {
		names, _, err := ParseName(readTable(t, font, "name"))
		tu.AssertNoErr(t, err)

		// NameFontFamily
		tu.Assert(t, strings.HasPrefix(names.Name(1), "Noto Sans"))
	}

	font := readFontFile(t, "common/Roboto-BoldItalic.ttf")
	names, _, err := ParseName(readTable(t, font, "name"))
	tu.AssertNoErr(t, err)
	// NameFontFamily
	tu.Assert(t, names.Name(1) == "Roboto")
}

func TestNames(t *testing.T) {
	for _, filename := range tu.Filenames(t, "common") {
		fp := readFontFile(t, filename)
		names, _, err := ParseName(readTable(t, fp, "name"))
		tu.AssertNoErr(t, err)

		for _, rec := range names.nameRecords {
			tu.Assert(t, names.Name(rec.nameID) != "")
		}

		tu.Assert(t, names.selectRecord(0xFFFF) == nil)
		tu.Assert(t, names.Name(0xFFFF) == "")
	}
}
