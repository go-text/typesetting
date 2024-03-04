// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

func TestParseSVG(t *testing.T) {
	fp := readFontFile(t, "toys/chromacheck-svg.ttf")
	_, _, err := ParseSVG(readTable(t, fp, "SVG "))
	tu.AssertNoErr(t, err)
}

func TestParseCFF(t *testing.T) {
	fp := readFontFile(t, "toys/CFFTest.otf")
	readTable(t, fp, "CFF ")
}
