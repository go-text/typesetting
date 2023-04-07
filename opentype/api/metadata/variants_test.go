package metadata

import (
	"bytes"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/opentype/loader"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

func Test_isMonospace(t *testing.T) {
	for _, file := range tu.Filenames(t, "common") {
		f, err := td.Files.ReadFile(file)
		tu.AssertNoErr(t, err)

		ld, err := loader.NewLoader(bytes.NewReader(f))
		tu.AssertNoErr(t, err)

		fd := newFontDescriptor(ld)
		tu.Assert(t, td.Monospace[file] == fd.isMonospace())
	}

	tu.Assert(t, !(&fontDescriptor{}).isMonospace()) // check it does not crash
}
