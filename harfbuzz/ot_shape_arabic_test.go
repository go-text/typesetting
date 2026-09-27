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

func TestReorderMarksBelowClass(t *testing.T) {
	b := NewBuffer()
	b.AddRunes([]rune{0x0655 /* hamza below, MCM */, 0x0650 /* kasra */}, 0, -1)
	for i := range b.Info {
		b.Info[i].setUnicodeProps(b)
		b.Info[i].setModifiedCombiningClass(220)
	}
	(&complexShaperArabic{}).reorderMarks(nil, b, 0, 2)
	if got := b.Info[0].getModifiedCombiningClass(); got != mcc22 {
		t.Fatalf("expected %d, got %d", mcc22, got)
	}
}
