package fontscan

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	meta "github.com/go-text/typesetting/opentype/api/metadata"
)

func Test_serializeFootprints(t *testing.T) {
	input := []footprint{
		{
			Family: "a strange one",
			Runes:  newRuneSet(1, 0, 2, 0x789, 0xfffee),
			Aspect: meta.Aspect{1, 200, 0.45},
		},
		{
			Runes: runeSet{},
		},
	}
	dump := serializeFootprintsTo(input, nil)

	got, err := deserializeFootprints(dump)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(input, got) {
		t.Fatalf("expected %v, got %v", input, got)
	}
}

func assertFontsetEquals(expected, got []footprint) error {
	if len(expected) != len(got) {
		return fmt.Errorf("invalid length: expected %d, got %d", len(expected), len(got))
	}
	for i := range got {
		expectedFootprint, gotFootprint := expected[i], got[i]
		if !reflect.DeepEqual(expectedFootprint, gotFootprint) {
			return fmt.Errorf("expected Footprint \n %v \n got \n %v", expectedFootprint, gotFootprint)
		}
	}
	return nil
}

func TestSerializeSystemFonts(t *testing.T) {
	directories, err := DefaultFontDirectories()
	if err != nil {
		t.Fatal(err)
	}

	fontset, err := scanFontFootprints(nil, directories...)
	if err != nil {
		t.Fatal(err)
	}

	ti := time.Now()
	var b bytes.Buffer
	err = fontset.serializeTo(&b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%d fonts serialized (into memory) in %s; size: %dKB\n", len(fontset), time.Since(ti), b.Len()/1000)

	fontset2, err := deserializeIndex(&b)
	if err != nil {
		t.Fatal(err)
	}
	if err = assertFontsetEquals(fontset.flatten(), fontset2.flatten()); err != nil {
		t.Fatalf("inconsistent serialization %s", err)
	}
}
