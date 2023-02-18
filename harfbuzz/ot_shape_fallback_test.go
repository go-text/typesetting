package harfbuzz

import "testing"

func TestRecategorize(t *testing.T) {
	runes := []rune{1615, 1617, 1614, 1616}
	ccc := []uint8{32, 27, 31, 33}
	exps := []uint8{230, 230, 230, 220}
	for i, r := range runes {
		exp := exps[i]
		got := recategorizeCombiningClass(r, ccc[i])
		if exp != got {
			t.Fatalf("for rune %d and class %d, expected %d, got %d", r, ccc[i], exp, got)
		}
	}
}
