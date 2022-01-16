package fontscan

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_serializeFootprints(t *testing.T) {
	input := []Footprint{
		{
			Family: "a strange one",
			Runes:  NewRuneSet(1, 0, 2, 0x789, 0xfffee),
			Aspect: Aspect{1, 200, 0.45},
			Format: OpenType,
		},
		{
			Runes: RuneSet{},
		},
	}
	w := &bytes.Buffer{}
	if err := serializeFootprints(input, w); err != nil {
		t.Fatal(err)
	}

	got, err := deserializeFootprints(w.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(input, got) {
		t.Fatalf("expected %v, got %v", input, got)
	}
}
