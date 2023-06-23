package fontscan

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/go-text/typesetting/opentype/api"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

// newRuneSet builds a set containing the given runes.
func newRuneSet(runes ...rune) runeSet {
	var rs runeSet
	for _, r := range runes {
		rs.Add(r)
	}
	return rs
}

func randomRunes() []rune {
	out := make([]rune, 1000)
	for i := range out {
		out[i] = rand.Int31()
	}
	return out
}

// runes returns a copy of the runes in the set.
func (rs runeSet) runes() (out []rune) {
	for _, page := range rs {
		pageLow := rune(page.ref) << 8
		for j, set := range page.set {
			for k := rune(0); k < 32; k++ {
				if set&uint32(1<<k) != 0 {
					out = append(out, pageLow|rune(j)<<5|k)
				}
			}
		}
	}
	return out
}

func TestRuneSet(t *testing.T) {
	tests := []struct {
		start    []rune
		expected []rune
		r        rune
	}{
		{
			nil,
			[]rune{0},
			0,
		},
		{
			nil,
			[]rune{1},
			1,
		},
		{
			nil,
			[]rune{32},
			32,
		},
		{
			[]rune{0, 32, 257},
			[]rune{0, 32, 257, 512},
			512,
		},
		{
			[]rune{0, 32, 257, 1000, 2000, 1500},
			[]rune{0, 32, 257, 1000, 1500, 2000, 10000},
			10000,
		},
	}
	for _, tt := range tests {
		cov := newRuneSet(tt.start...)
		cov.Add(tt.r)
		if runes := cov.runes(); !reflect.DeepEqual(runes, tt.expected) {
			t.Fatalf("expected %v, got %v (%v)", tt.expected, runes, cov)
		}

		for _, r := range tt.expected {
			if !cov.Contains(r) {
				t.Fatalf("missing char %d", r)
			}
		}

		cov.Delete(tt.r)
		sort.Slice(tt.start, func(i, j int) bool { return tt.start[i] < tt.start[j] })
		if runes := cov.runes(); !reflect.DeepEqual(runes, tt.start) {
			t.Fatalf("expected %v, got %v (%v)", tt.start, runes, cov)
		}

		for _, r := range tt.start {
			if !cov.Contains(r) {
				t.Fatalf("missing char %d", r)
			}
		}
		if cov.Contains(tt.r) {
			t.Fatalf("rune %d should be deleted", tt.r)
		}
		if cov.Contains(1_000_000) {
			t.Fatalf("rune %d should be missing", 1_000_000)
		}

		cov.Delete(1_000_000) // no op

		if cov.Len() != len(tt.start) {
			t.Fatalf("unexpected length %d", cov.Len())
		}
	}
}

func TestBinaryFormat(t *testing.T) {
	for range [50]int{} {
		cov := newRuneSet(randomRunes()...)
		b := cov.serialize()

		var got runeSet
		n, err := got.deserializeFrom(b)
		if err != nil {
			t.Fatalf("Coverage.deserializeFrom: %s", err)
		}

		if n != len(b) {
			t.Fatalf("unexpected number of bytes read: %d", n)
		}

		if !reflect.DeepEqual(cov, got) {
			t.Fatalf("expected %v, got %v", cov, got)
		}
	}
}

func TestDeserializeFrom(t *testing.T) {
	var cov runeSet

	if _, err := cov.deserializeFrom(nil); err == nil {
		t.Fatal("exepcted error on invalid input")
	}
	if _, err := cov.deserializeFrom([]byte{0, 5}); err == nil {
		t.Fatal("exepcted error on invalid input")
	}
}

// CmapSimple is a map based Cmap implementation.
type CmapSimple map[rune]api.GID

type cmap0Iter struct {
	data CmapSimple
	keys []rune
	pos  int
}

func (it *cmap0Iter) Next() bool {
	return it.pos < len(it.keys)
}

func (it *cmap0Iter) Char() (rune, api.GID) {
	r := it.keys[it.pos]
	it.pos++
	return r, it.data[r]
}

func (s CmapSimple) Iter() api.CmapIter {
	keys := make([]rune, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return &cmap0Iter{data: s, keys: keys}
}

func (s CmapSimple) Lookup(r rune) (api.GID, bool) {
	v, ok := s[r] // will be 0 if r is not in s
	return v, ok
}

func TestNewRuneSetFromCmap(t *testing.T) {
	tests := []struct {
		args api.Cmap
		want runeSet
	}{
		{CmapSimple{0: 0, 1: 0, 2: 0, 0xfff: 0}, newRuneSet(0, 1, 2, 0xfff)},
		{CmapSimple{0: 0, 1: 0, 2: 0, 800: 0, 801: 0, 1000: 0}, newRuneSet(0, 1, 2, 800, 801, 1000)},
	}
	for _, tt := range tests {
		if got, _ := newRuneSetFromCmap(tt.args, nil); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("NewRuneSetFromCmap() = %v, want %v", got, tt.want)
		}
	}
}

func TestBits(t *testing.T) {
	a, b := 2, 13
	var total uint32
	for i := a; i <= b; i++ {
		total |= 1 << i
	}

	alt := (uint32(1)<<(b-a+1) - 1) << a // mask for bits from a to b (included)
	tu.Assert(t, total == alt)
}

type runeRange [][2]rune

func (rr runeRange) RuneRanges(_ [][2]rune) [][2]rune { return rr }

func (rr runeRange) runes() (out []rune) {
	for _, ra := range rr {
		for r := ra[0]; r <= ra[1]; r++ {
			out = append(out, r)
		}
	}
	return out
}

func TestRuneRanges(t *testing.T) {
	for _, source := range []runeRange{
		{
			{0, 10}, {12, 14}, {15, 15}, {18, 2000}, {2100, 0xFFFFFF},
		},
		{
			{0, 30}, {0xFF, 0xFF * 2},
		},
	} {
		got, _ := newRuneSetFromCmapRange(source, nil)
		exp := newRuneSet(source.runes()...)
		tu.Assert(t, reflect.DeepEqual(got, exp))
	}
}
