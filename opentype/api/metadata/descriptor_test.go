package metadata

import (
	"bytes"
	"fmt"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/opentype/loader"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

func TestMetadata(t *testing.T) {
	tests := []struct {
		fontPath string
		aspect   Aspect
		family   string
	}{
		{
			"common/Roboto-BoldItalic.ttf",
			Aspect{StyleItalic, WeightBold, StretchNormal},
			"Roboto",
		},
		{
			"common/NotoSansArabic.ttf",
			Aspect{StyleNormal, WeightNormal, StretchNormal},
			"Noto Sans Arabic",
		},
	}

	for _, test := range tests {
		f, err := td.Files.ReadFile(test.fontPath)
		tu.AssertNoErr(t, err)

		ld, err := loader.NewLoader(bytes.NewReader(f))
		tu.AssertNoErr(t, err)

		got := Metadata(ld)
		tu.AssertC(t, got.Aspect == test.aspect, fmt.Sprint(got.Aspect))
		tu.AssertC(t, got.Family == test.family, got.Family)
	}
}

func Test_isMonospace(t *testing.T) {
	for _, file := range tu.Filenames(t, "common") {
		f, err := td.Files.ReadFile(file)
		tu.AssertNoErr(t, err)

		ld, err := loader.NewLoader(bytes.NewReader(f))
		tu.AssertNoErr(t, err)

		fd := newFontDescriptor(ld)
		tu.AssertC(t, td.Monospace[file] == fd.isMonospace(), file)
	}

	tu.Assert(t, !(&fontDescriptor{}).isMonospace()) // check it does not crash
}
