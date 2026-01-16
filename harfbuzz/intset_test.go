package harfbuzz

import (
	"reflect"
	"sort"
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

// newIntSet builds a set containing the given runes.
func newIntSet(runes ...uint32) intSet {
	var rs intSet
	for _, r := range runes {
		rs.add(r)
	}
	return rs
}

// ints is an helper method returning a copy of the runes in the set.
func (rs intSet) ints() (out []uint32) {
	for _, page := range rs {
		pageLow := uint32(page.ref) << 8
		for j, set := range page.set {
			for k := uint32(0); k < 32; k++ {
				if set&uint32(1<<k) != 0 {
					out = append(out, pageLow|uint32(j)<<5|k)
				}
			}
		}
	}
	return out
}

func TestRuneSet(t *testing.T) {
	tests := []struct {
		start    []uint32
		expected []uint32
		r        uint32
	}{
		{
			nil,
			[]uint32{0},
			0,
		},
		{
			nil,
			[]uint32{1},
			1,
		},
		{
			nil,
			[]uint32{32},
			32,
		},
		{
			[]uint32{0, 32, 257},
			[]uint32{0, 32, 257, 512},
			512,
		},
		{
			[]uint32{0, 32, 257, 1000, 2000, 1500},
			[]uint32{0, 32, 257, 1000, 1500, 2000, 10000},
			10000,
		},
	}
	for _, tt := range tests {
		cov := newIntSet(tt.start...)
		cov.add(tt.r)
		if runes := cov.ints(); !reflect.DeepEqual(runes, tt.expected) {
			t.Fatalf("expected %v, got %v (%v)", tt.expected, runes, cov)
		}

		for _, r := range tt.expected {
			if !cov.has(r) {
				t.Fatalf("missing char %d", r)
			}
		}

		cov.delete(tt.r)
		sort.Slice(tt.start, func(i, j int) bool { return tt.start[i] < tt.start[j] })
		if runes := cov.ints(); !reflect.DeepEqual(runes, tt.start) {
			t.Fatalf("expected %v, got %v (%v)", tt.start, runes, cov)
		}

		for _, r := range tt.start {
			if !cov.has(r) {
				t.Fatalf("missing char %d", r)
			}
		}
		if cov.has(tt.r) {
			t.Fatalf("rune %d should be deleted", tt.r)
		}
		if cov.has(1_000_000) {
			t.Fatalf("rune %d should be missing", 1_000_000)
		}

		cov.delete(1_000_000) // no op

		if cov.Len() != len(tt.start) {
			t.Fatalf("unexpected length %d", cov.Len())
		}
	}
}

func TestIntersects(t *testing.T) {
	tu.Assert(t, newIntSet(0, 1, 2, 3, 4).Intersects(newIntSet(0, 1, 2, 3)))
	tu.Assert(t, newIntSet(0, 1, 2, 3).Intersects(newIntSet(0, 1, 2, 3)))
	tu.Assert(t, newIntSet(0, 1, 2, 3, 12, 24).Intersects(newIntSet(0, 1, 12)))
	tu.Assert(t, newIntSet(0, 1, 2, 3, 12000, 13000).Intersects(newIntSet(0, 1, 12000)))
	tu.Assert(t, newIntSet(0, 1, 2, 50000).Intersects(newIntSet(3, 50000)))
	tu.Assert(t, newIntSet(0, 1, 2, 1000, 50000).Intersects(newIntSet(10000, 50000)))
	tu.Assert(t, !newIntSet(0, 1, 2).Intersects(newIntSet(3, 12000)))
	tu.Assert(t, !newIntSet(4, 5, 6, 8).Intersects(newIntSet(1, 9, 10, 1000)))
}

func TestAddRange(t *testing.T) {
	for _, test := range []struct {
		ranges   [][2]uint32
		expected []uint32
	}{
		{[][2]uint32{{0, 1}}, []uint32{0, 1}},
		{[][2]uint32{{0, 1}, {3, 9}}, []uint32{0, 1, 3, 4, 5, 6, 7, 8, 9}},
		{[][2]uint32{{255, 258}, {1024, 1024}}, []uint32{255, 256, 257, 258, 1024}},
	} {
		var s intSet
		for _, r := range test.ranges {
			s.addRange(r[0], r[1])
		}
		tu.Assert(t, reflect.DeepEqual(s.ints(), test.expected))
	}
}
