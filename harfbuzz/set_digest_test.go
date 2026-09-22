package harfbuzz

import "testing"

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
