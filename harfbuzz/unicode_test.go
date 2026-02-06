package harfbuzz

import (
	"testing"

	ucd "github.com/go-text/typesetting/internal/unicodedata"
)

// ported from harfbuzz/test/api/test-unicode.c Copyright © 2011  Codethink Limited, Google, Inc. Ryan Lortie, Behdad Esfahbod

func TestUnicodeProp(t *testing.T) {
	runes := []rune{6176, 6155, 0x70f}
	exps := []unicodeProp{7, 236, unicodeProp(ucd.Cf)}
	for i, r := range runes {
		got, _ := computeUnicodeProps(r)
		exp := exps[i]
		if got != exp {
			t.Fatalf("for rune 0x%x, expected %d, got %d", r, exp, got)
		}
	}
}
