package harfbuzz

import (
	"fmt"
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

func TestIndicGetCategories(t *testing.T) {
	expecteds := map[rune]struct{ syllable, position uint8 }{
		0x0b55: {indSM_ex_N, posEnd},
		0x103A: {myaSM_ex_As, posEnd},
		0x103B: {myaSM_ex_MY, posEnd},
		0x17D0: {khmSM_ex_Xgroup, posEnd},
		0x17E0: {myaSM_ex_GB, posBaseC},
		// myanmar
		4100: {indSM_ex_Ra, posBaseC},
		4123: {indSM_ex_Ra, posBaseC},
		4141: {myaSM_ex_VAbv, posAboveC},
		4153: {indSM_ex_H, posEnd},
		4157: {myaSM_ex_MW, posEnd},
	}
	for u, exp := range expecteds {
		got := indicGetCategories(u)
		syl, pos := uint8(got&0xFF), uint8(got>>8)
		tu.AssertC(t, syl == exp.syllable, fmt.Sprint("rune ", u, syl))
		tu.AssertC(t, pos == exp.position, fmt.Sprint("rune ", u, pos))
	}
}
