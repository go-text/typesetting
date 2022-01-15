package fontscan

import (
	"bytes"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func randomRunes() []rune {
	out := make([]rune, 1000)
	for i := range out {
		out[i] = rand.Int31()
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
		cov := NewRuneSet(tt.start...)
		cov.Add(tt.r)
		if runes := cov.Runes(); !reflect.DeepEqual(runes, tt.expected) {
			t.Fatalf("expected %v, got %v (%v)", tt.expected, runes, cov)
		}

		for _, r := range tt.expected {
			if !cov.Contains(r) {
				t.Fatalf("missing char %d", r)
			}
		}

		cov.Delete(tt.r)
		sort.Slice(tt.start, func(i, j int) bool { return tt.start[i] < tt.start[j] })
		if runes := cov.Runes(); !reflect.DeepEqual(runes, tt.start) {
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
		cov := NewRuneSet(randomRunes()...)
		var b bytes.Buffer
		err := cov.serializeTo(&b)
		if err != nil {
			t.Fatalf("Coverage.serializeTo: %s", err)
		}

		var got RuneSet
		n, err := got.deserializeFrom(b.Bytes())
		if err != nil {
			t.Fatalf("Coverage.deserializeFrom: %s", err)
		}

		if n != b.Len() {
			t.Fatalf("unexpected number of bytes read: %d", n)
		}

		if !reflect.DeepEqual(cov, got) {
			t.Fatalf("expected %v, got %v", cov, got)
		}
	}
}

func TestDeserializeFrom(t *testing.T) {
	var cov RuneSet

	if _, err := cov.deserializeFrom(nil); err == nil {
		t.Fatal("exepcted error on invalid input")
	}
	if _, err := cov.deserializeFrom([]byte{0, 5}); err == nil {
		t.Fatal("exepcted error on invalid input")
	}
}

func TestCoverage_isSubset(t *testing.T) {
	tests := []struct {
		a    RuneSet
		b    RuneSet
		want bool
	}{
		{NewRuneSet(), NewRuneSet(), true},
		{NewRuneSet(1, 10, 0x78DD), NewRuneSet(1, 10, 0x78DD), true},
		{NewRuneSet(1, 10, 0x78DD), NewRuneSet(1, 10, 0x78DD, 13), true},
		{NewRuneSet(1, 10, 0x78DD, 12), NewRuneSet(1, 10, 0x78DD), false},
		{NewRuneSet(0x78DD), NewRuneSet(1, 10), false},
		{NewRuneSet(1, 10), NewRuneSet(0x78DD), false},
	}
	for _, tt := range tests {
		if got := tt.a.isSubset(tt.b); got != tt.want {
			t.Errorf("Coverage.isSubset() = %v, want %v", got, tt.want)
		}
	}
}
