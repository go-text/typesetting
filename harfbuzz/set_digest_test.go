package harfbuzz

import (
	"fmt"
	"testing"

	"github.com/go-text/typesetting/font"
	tu "github.com/go-text/typesetting/testutils"
)

func TestDigest(t *testing.T) {
	const (
		setTypeSize = 2
		numBits     = 3 + 1 + 1
	)
	if shift0 >= setTypeSize*8 {
		t.Error()
	}
	if shift0+numBits > setTypeSize*8 {
		t.Error()
	}
	if shift1 >= setTypeSize*8 {
		t.Error()
	}
	if shift1+numBits > setTypeSize*8 {
		t.Error()
	}
	if shift2 >= setTypeSize*8 {
		t.Error()
	}
	if shift2+numBits > setTypeSize*8 {
		t.Error()
	}
}

func TestDigestHas(t *testing.T) {
	var d setDigest
	for i := setType(10); i < 65_000; i += 7 {
		d.add(i)
	}
	for i := setType(10); i < 65_000; i += 7 {
		if !d.mayHave(i) {
			t.Errorf("expected <may have> for %d", i)
		}
	}
	for i := setType(0); i < 0xFFFF; i++ { // care with overflow
		// if the filter is negative, then the glyph must not be in the set
		if !d.mayHave(i) {
			if (i-10)%7 == 0 {
				t.Errorf("<not have> for glyph %d present in set", i)
			}
		}
	}
}

func TestDigestRangeAfterFlood(t *testing.T) {
	// A range spanning 63 glyphs or more fills the shift-0 sub-digest; the ranges
	// added after it must still reach the other two (Noto Sans Arabic's ccmp
	// coverage: [577-718] first, the noon at 759 in a later range).
	var d setDigest
	d.addRange(577, 718)
	d.addRange(737, 762)
	for _, g := range []setType{577, 700, 718, 737, 759, 762} {
		if !d.mayHave(g) {
			t.Errorf("expected <may have> for %d, which is in an added range", g)
		}
	}
	if d.mayHave(730) {
		t.Errorf("<may have> for 730, which no range covers, means the filter flooded")
	}
}

// TestDigestKerningAfterFlood shapes pairs whose first glyph a lookup's coverage
// lists after a range of 63 glyphs or more: Raleway's kern lookup covers f-s (and
// their accented forms) that way, Commissioner's N-T and o-u. The digest dropped
// those ranges and the pairs went unkerned, while AV, LT and fo, whose first glyph
// is in the coverage's first range, were kerned all along. The advances are
// HarfBuzz's (14.5), in font units.
func TestDigestKerningAfterFlood(t *testing.T) {
	for _, test := range []struct {
		file    string
		text    string
		advance Position // of the first glyph
	}{
		{"common/Raleway-v4020-Regular.otf", "ro", 335},
		{"common/Raleway-v4020-Regular.otf", "ke", 503},
		{"common/Raleway-v4020-Regular.otf", "ly", 255},
		{"common/Raleway-v4020-Regular.otf", "ov", 577},
		{"common/Raleway-v4020-Regular.otf", "AV", 617},
		{"common/Raleway-v4020-Regular.otf", "LT", 471},
		{"common/Commissioner-VF.ttf", "To", 1114},
		{"common/Commissioner-VF.ttf", "Ta", 1074},
		{"common/Commissioner-VF.ttf", "Ro", 1246},
		{"common/Commissioner-VF.ttf", "ov", 1164},
		{"common/Commissioner-VF.ttf", "fo", 708},
	} {
		ft := openFontFileTT(t, test.file)
		fnt := NewFont(font.NewFace(ft))
		buffer := NewBuffer()
		buffer.AddRunes([]rune(test.text), 0, -1)
		buffer.GuessSegmentProperties()
		buffer.Shape(fnt, nil)
		tu.AssertC(t, buffer.Pos[0].XAdvance == test.advance,
			fmt.Sprintf("%s %q: first advance %d, expected %d (kerned)", test.file, test.text, buffer.Pos[0].XAdvance, test.advance))
	}
}
