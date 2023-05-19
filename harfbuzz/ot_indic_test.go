package harfbuzz

import "testing"

func TestGetIndicCategories(t *testing.T) {
	expecteds := map[rune]uint16{
		2901: 1543,
		4154: 1543,
	}
	for u, type_ := range expecteds {
		if got := indicGetCategories(u); got != type_ {
			t.Fatalf("expected indic categorie of %d for rune %d, got %d", type_, u, got)
		}
	}
}

func TestComputeIndicProperties(t *testing.T) {
	type_ := indicGetCategories(2901)
	cat, pos := uint8(type_&0xFF), uint8(type_>>8)
	if cat != 3 || pos != 6 {
		t.Fatalf("expected 3,6 for rune 2901, got %d, %d", cat, pos)
	}
}
