package fontscan

import (
	"reflect"
	"testing"
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
		if mark == nil {
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
		if mark == nil {
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
		if mark == nil {
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
