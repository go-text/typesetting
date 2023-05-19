package harfbuzz

import "testing"

func TestMyanmarProperties(t *testing.T) {
	expecteds := map[rune][2]uint8{
		4100: {16, 15},
		4154: {18, 6},
		4153: {4, 15},
		4123: {16, 15},
		4157: {23, 8},
		4141: {26, 6},
	}
	for u, exp := range expecteds {
		type_ := indicGetCategories(u)
		gotCat := uint8(type_ & 0xFF)
		gotPos := uint8(type_ >> 8)
		if exp[0] != gotCat || exp[1] != gotPos {
			t.Fatalf("for rune %d, expected %v, got [%d, %d]", u, exp, gotCat, gotPos)
		}
	}
}
