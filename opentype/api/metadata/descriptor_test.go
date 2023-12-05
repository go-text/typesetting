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
		{
			"common/DejaVuSans.ttf",
			Aspect{StyleNormal, WeightNormal, StretchNormal},
			"DejaVu Sans",
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

		fd := newFontMetrics(ld)
		tu.AssertC(t, td.Monospace[file] == fd.isMonospace(), file)
	}

	tu.Assert(t, !(&fontMetrics{}).isMonospace()) // check it does not crash
}

func TestAspect_inferFromStyle(t *testing.T) {
	styn, wn, sten := StyleNormal, WeightNormal, StretchNormal
	tests := []struct {
		args   string
		fields Aspect
		want   Aspect
	}{
		{
			"", Aspect{styn, wn, sten}, Aspect{styn, wn, sten}, // no op
		},
		{
			"Black", Aspect{0, 0, 0}, Aspect{0, WeightBlack, 0},
		},
		{
			"conDensed", Aspect{0, 0, 0}, Aspect{0, 0, StretchCondensed},
		},
		{
			"ITALIC", Aspect{0, 0, 0}, Aspect{StyleItalic, 0, 0},
		},
		{
			"black", Aspect{0, WeightNormal, 0}, Aspect{0, WeightNormal, 0}, // respect initial value
		},
		{
			"black oblique", Aspect{0, 0, 0}, Aspect{StyleItalic, WeightBlack, 0},
		},
	}
	for _, tt := range tests {
		as := tt.fields
		as.inferFromStyle(tt.args)
		tu.AssertC(t, as == tt.want, tt.args)
	}
}

func TestAspectFromOS2(t *testing.T) {
	// This font has two different weight values :
	// 400, in the OS/2 table and 380, in the style description
	f, err := td.Files.ReadFile("common/DejaVuSans.ttf")
	tu.AssertNoErr(t, err)

	ld, err := loader.NewLoader(bytes.NewReader(f))
	tu.AssertNoErr(t, err)

	fd, _ := newFontDescriptor(ld, nil)

	raw := fd.rawAspect()
	tu.Assert(t, raw.Weight == WeightNormal)

	var inferred Aspect
	inferred.inferFromStyle(fd.additionalStyle())
	tu.Assert(t, inferred.Weight == 380)
}
