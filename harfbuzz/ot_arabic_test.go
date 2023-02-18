package harfbuzz

import "testing"

func TestNumArabicLookup(t *testing.T) {
	if len(arabicFallbackFeatures) > arabicFallbackMaxLookups {
		t.Error()
	}
}
