package cff

import (
	"encoding/binary"
	"testing"
)

// cff2IndexHeader builds the 5 byte CFF2 INDEX header (count, then offSize)
// that indexStart.mustParse consumes, followed by the given body.
func cff2IndexHeader(count uint32, offSize uint8, body ...byte) []byte {
	out := make([]byte, 5, 5+len(body))
	binary.BigEndian.PutUint32(out, count)
	out[4] = offSize
	return append(out, body...)
}

// A CFF2 INDEX header is decoded by generated code which does not validate
// its fields, so parseIndexContent must reject implausible ones itself
// instead of trusting count to size an allocation. Each case below made
// v0.3.4 allocate up to count*24 bytes (about 103 GB for a count of
// 0xFFFFFFFF) before panicking in bigEndian.
func TestParseIndex2RejectsUnusableHeader(t *testing.T) {
	for _, test := range []struct {
		name    string
		count   uint32
		offSize uint8
	}{
		// offSize 0 makes offsetArraySize 0, so the EOF check passes
		// whatever count says.
		{"zero offSize, huge count", 0xFFFFFFFF, 0},
		{"zero offSize, small count", 1, 0},
		// count+1 computed in uint32 wraps to 0, so offsetArraySize is 0
		// again even though offSize is valid.
		{"valid offSize, wrapping count", 0xFFFFFFFF, 4},
		// offSize above the 1-4 range the spec allows.
		{"out of range offSize", 1, 5},
		// count is plausible on its own but cannot fit in src.
		{"count larger than src", 1000, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			src := cff2IndexHeader(test.count, test.offSize, 1, 2, 3, 4, 5, 6, 7, 8)
			if _, err := parseIndex2(src, 0); err == nil {
				t.Fatalf("parseIndex2(count=%d, offSize=%d): expected an error, got nil",
					test.count, test.offSize)
			}
		})
	}
}

// Guard against the validation above rejecting well formed input.
func TestParseIndex2AcceptsValidIndex(t *testing.T) {
	// Two objects, 1 byte offsets. Offsets are stored off by one, so
	// {1, 3, 5} delimits "ab" and "cd".
	src := cff2IndexHeader(2, 1, 1, 3, 5, 'a', 'b', 'c', 'd')

	got, err := parseIndex2(src, 0)
	if err != nil {
		t.Fatalf("parseIndex2: unexpected error: %s", err)
	}
	want := []string{"ab", "cd"}
	if len(got) != len(want) {
		t.Fatalf("parseIndex2: expected %d objects, got %d", len(want), len(got))
	}
	for i, w := range want {
		if string(got[i]) != w {
			t.Fatalf("parseIndex2: object %d: expected %q, got %q", i, w, got[i])
		}
	}
}
