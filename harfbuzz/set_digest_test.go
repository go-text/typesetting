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
