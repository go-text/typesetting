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

		gotAspect, gotFamily := Metadata(ld)
		tu.AssertC(t, gotAspect == test.aspect, fmt.Sprint(gotAspect))
		tu.AssertC(t, gotFamily == test.family, gotFamily)
	}
}
