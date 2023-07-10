package harfbuzz

import (
	"testing"

	"github.com/go-text/typesetting/language"
)

func TestNumArabicLookup(t *testing.T) {
	if len(arabicFallbackFeatures) > arabicFallbackMaxLookups {
		t.Error()
	}
}

func TestHasArabicJoining(t *testing.T) {
	if !hasArabicJoining(language.Arabic) {
		t.Fatal()
	}
	if hasArabicJoining(language.Linear_A) {
		t.Fatal()
	}
}
