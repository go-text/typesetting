package fontscan

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
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

	got, err := deserializeFootprints(w)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(input, got) {
		t.Fatalf("expected %v, got %v", input, got)
	}
}

func TestSerializeSystemFonts(t *testing.T) {
	directories, err := DefaultFontDirs()
	if err != nil {
		t.Fatal(err)
	}

	fontset, err := ScanFonts(directories...)
	if err != nil {
		t.Fatal(err)
	}

	ti := time.Now()
	var b bytes.Buffer
	err = serializeFootprints(fontset, &b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%d fonts serialized (into memory) in %s; size: %dKB\n", len(fontset), time.Since(ti), b.Len()/1000)
}
