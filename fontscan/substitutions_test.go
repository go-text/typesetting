package fontscan

import (
	"reflect"
	"testing"

	"github.com/go-text/typesetting/font"
	tu "github.com/go-text/typesetting/testutils"
)

func Test_familyList_insertStart(t *testing.T) {
	tests := []struct {
		start    []string
		families []string
		want     familyCrible
	}{
		{nil, nil, familyCrible{}},
		{nil, []string{"aa", "bb"}, familyCrible{"aa": 0, "bb": 1}},
		{[]string{"a"}, []string{"aa", "bb"}, familyCrible{"aa": 0, "bb": 1, "a": 2}},
	}
	for _, tt := range tests {
		l := newFamilyList(tt.start)
		l.insertStart(tt.families)
		c := make(familyCrible)
		if l.compileTo(c); !reflect.DeepEqual(c, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, c)
		}
	}
}

func Test_familyList_insertEnd(t *testing.T) {
	tests := []struct {
		start    []string
		families []string
		want     familyCrible
	}{
		{nil, nil, familyCrible{}},
		{nil, []string{"aa", "bb"}, familyCrible{"aa": 0, "bb": 1}},
		{[]string{"a"}, []string{"aa", "bb"}, familyCrible{"a": 0, "aa": 1, "bb": 2}},
	}
	for _, tt := range tests {
		l := newFamilyList(tt.start)
		l.insertEnd(tt.families)
		c := make(familyCrible)
		if l.compileTo(c); !reflect.DeepEqual(c, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, c)
		}
	}
}

func Test_familyList_insertAfter(t *testing.T) {
	tests := []struct {
		start    []string
		element  string
		families []string
		want     familyCrible
	}{
		{[]string{"f1"}, "f1", []string{"aa", "bb"}, familyCrible{"f1": 0, "aa": 1, "bb": 2}},
		{[]string{"f1", "f2"}, "f1", []string{"aa", "bb"}, familyCrible{"f1": 0, "aa": 1, "bb": 2, "f2": 3}},
	}
	for _, tt := range tests {
		l := newFamilyList(tt.start)
		mark := l.elementEquals(tt.element)
		if mark < 0 {
			t.Fatalf("element %s not found in %v", tt.element, l)
		}
		l.insertAfter(mark, tt.families)
		c := make(familyCrible)
		if l.compileTo(c); !reflect.DeepEqual(c, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, c)
		}
	}
}

func Test_familyList_insertBefore(t *testing.T) {
	tests := []struct {
		start    []string
		element  string
		families []string
		want     familyCrible
	}{
		{[]string{"f1"}, "f1", []string{"aa", "bb"}, familyCrible{"aa": 0, "bb": 1, "f1": 2}},
		{[]string{"f1", "f2"}, "f2", []string{"aa", "bb"}, familyCrible{"f1": 0, "aa": 1, "bb": 2, "f2": 3}},
	}
	for _, tt := range tests {
		l := newFamilyList(tt.start)
		mark := l.elementEquals(tt.element)
		if mark < 0 {
			t.Fatalf("element %s not found in %v", tt.element, l)
		}
		l.insertBefore(mark, tt.families)
		c := make(familyCrible)
		if l.compileTo(c); !reflect.DeepEqual(c, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, c)
		}
	}
}

func Test_familyList_replace(t *testing.T) {
	tests := []struct {
		start    []string
		element  string
		families []string
		want     familyCrible
	}{
		{[]string{"f1"}, "f1", []string{"aa", "bb"}, familyCrible{"aa": 0, "bb": 1}},
		{[]string{"f1", "f2"}, "f2", []string{"aa", "bb"}, familyCrible{"f1": 0, "aa": 1, "bb": 2}},
	}
	for _, tt := range tests {
		l := newFamilyList(tt.start)
		mark := l.elementEquals(tt.element)
		if mark < 0 {
			t.Fatalf("element %s not found in %v", tt.element, l)
		}
		l.replace(mark, tt.families)
		c := make(familyCrible)
		if l.compileTo(c); !reflect.DeepEqual(c, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, c)
		}
	}
}

func Test_familyList_execute(t *testing.T) {
	tests := []struct {
		start []string
		args  substitution
		want  familyCrible
	}{
		{nil, substitution{familyEquals("f2"), []string{"aa", "bb"}, opReplace}, familyCrible{}},                                  // no match
		{[]string{"f1", "f2"}, substitution{familyEquals("f4"), []string{"aa", "bb"}, opReplace}, familyCrible{"f1": 0, "f2": 1}}, // no match
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opReplace}, familyCrible{"f1": 0, "aa": 1, "bb": 2}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opAppend}, familyCrible{"f1": 0, "f2": 1, "aa": 2, "bb": 3}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opAppendLast}, familyCrible{"f1": 0, "f2": 1, "aa": 2, "bb": 3}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opPrepend}, familyCrible{"f1": 0, "aa": 1, "bb": 2, "f2": 3}},
		{[]string{"f1", "f2"}, substitution{familyEquals("f2"), []string{"aa", "bb"}, opPrependFirst}, familyCrible{"aa": 0, "bb": 1, "f1": 2, "f2": 3}},
	}
	for _, tt := range tests {
		fl := newFamilyList(tt.start)
		fl.execute(tt.args)
		c := make(familyCrible)
		if fl.compileTo(c); !reflect.DeepEqual(c, tt.want) {
			t.Fatalf("expected %v, got %v", tt.want, c)
		}
	}
}

func TestSubstituteHelvetica(t *testing.T) {
	fc := familyCrible{}
	fc.fillWithSubstitutions("helvetica")
	if f0 := fc.families()[0]; f0 != "helvetica" {
		t.Fatalf("unexpected family %s", f0)
	}
	if f4 := fc.families()[4]; f4 != "arial" {
		t.Fatalf("unexpected family %s", f4)
	}
}

func TestInsertAt(t *testing.T) {
	mkcap := func(cap int, strs ...string) []string {
		return append(make([]string, 0, cap), strs...)
	}
	clone := func(s []string) []string {
		r := make([]string, len(s), cap(s))
		copy(r, s)
		return r
	}

	tests := []struct {
		start  []string
		at     int
		add    []string
		result []string
	}{
		{[]string{}, 0, []string{"A"}, []string{"A"}},
		{[]string{"X"}, 0, []string{"A"}, []string{"A", "X"}},
		{[]string{"X"}, 1, []string{"A"}, []string{"X", "A"}},
		{mkcap(3, "X"), 0, []string{"A"}, []string{"A", "X"}},
		{mkcap(3, "X"), 1, []string{"A"}, []string{"X", "A"}},
		{mkcap(3, "X"), 0, []string{"A", "B"}, []string{"A", "B", "X"}},
		{mkcap(3, "X"), 1, []string{"A", "B"}, []string{"X", "A", "B"}},
		{mkcap(4, "X", "Y"), 0, []string{"A", "B"}, []string{"A", "B", "X", "Y"}},
		{mkcap(4, "X", "Y"), 1, []string{"A", "B"}, []string{"X", "A", "B", "Y"}},
		{mkcap(4, "X", "Y"), 2, []string{"A", "B"}, []string{"X", "Y", "A", "B"}},
	}
	for _, tt := range tests {
		result := insertAt(clone(tt.start), tt.at, tt.add)
		if !reflect.DeepEqual(tt.result, result) {
			t.Fatalf("insertAt: expected %v, got %v", tt.result, result)
		}

		// replaceAt functions like insertAt when start == end
		result = replaceAt(clone(tt.start), tt.at, tt.at, tt.add)
		if !reflect.DeepEqual(tt.result, result) {
			t.Fatalf("replaceAt: expected %v, got %v", tt.result, result)
		}
	}
}

func TestReplaceAt(t *testing.T) {
	mkcap := func(cap int, strs ...string) []string {
		return append(make([]string, 0, cap), strs...)
	}
	clone := func(s []string) []string {
		r := make([]string, len(s), cap(s))
		copy(r, s)
		return r
	}

	tests := []struct {
		start  []string
		at, to int
		add    []string
		result []string
	}{
		{mkcap(4, "X", "Y", "Z"), 0, 2, []string{"A"}, []string{"A", "Z"}},

		{mkcap(4, "X", "Y", "Z"), 0, 1, []string{"A"}, []string{"A", "Y", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 0, 1, []string{"A", "B"}, []string{"A", "B", "Y", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 0, 1, []string{"A", "B", "C"}, []string{"A", "B", "C", "Y", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 0, 2, []string{"A"}, []string{"A", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 0, 2, []string{"A", "B"}, []string{"A", "B", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 0, 2, []string{"A", "B", "C"}, []string{"A", "B", "C", "Z"}},

		{mkcap(4, "X", "Y", "Z"), 1, 2, []string{"A"}, []string{"X", "A", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 1, 2, []string{"A", "B"}, []string{"X", "A", "B", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 1, 2, []string{"A", "B", "C"}, []string{"X", "A", "B", "C", "Z"}},
		{mkcap(4, "X", "Y", "Z"), 1, 3, []string{"A"}, []string{"X", "A"}},
		{mkcap(4, "X", "Y", "Z"), 1, 3, []string{"A", "B"}, []string{"X", "A", "B"}},
		{mkcap(4, "X", "Y", "Z"), 1, 3, []string{"A", "B", "C"}, []string{"X", "A", "B", "C"}},
	}
	for _, tt := range tests {
		result := replaceAt(clone(tt.start), tt.at, tt.to, tt.add)
		if !reflect.DeepEqual(tt.result, result) {
			t.Fatalf("expected %v, got %v", tt.result, result)
		}
	}
}

func BenchmarkNewFamilyCrible(b *testing.B) {
	c := make(familyCrible)
	for i := 0; i < b.N; i++ {
		c.fillWithSubstitutions("Arial")
	}
}

func TestSubstituteHelveticaOrder(t *testing.T) {
	c := make(familyCrible)
	c.fillWithSubstitutionsList([]string{font.NormalizeFamily("BlinkMacSystemFont"), font.NormalizeFamily("Helvetica")})
	// BlinkMacSystemFont is not known by the library, so it is expanded with generic sans-serif,
	// but with lower priority then Helvetica
	l := c.families()
	tu.Assert(t, l[0] == font.NormalizeFamily("BlinkMacSystemFont"))
	tu.Assert(t, l[1] == font.NormalizeFamily("Helvetica"))
	tu.Assert(t, l[2] == font.NormalizeFamily("Nimbus Sans")) // from Helvetica
}
