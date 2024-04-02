package fontscan

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
)

func Test_serializeFootprints(t *testing.T) {
	input := []Footprint{
		{
			Family:  "a strange one",
			Runes:   newRuneSet(1, 0, 2, 0x789, 0xfffee),
			Scripts: ScriptSet{0, 1, 5, 0xffffff, language.Nabataean, language.Unknown},
			Aspect:  font.Aspect{Style: 1, Weight: 200, Stretch: 0.45},
		},
		{
			Runes:   RuneSet{},
			Scripts: ScriptSet{},
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

// Test_serializeEmpty ensures that serializing an empty index is safe.
func Test_serializeEmpty(t *testing.T) {
	input := []Footprint{}
	dump := serializeFootprintsTo(input, nil)

	got, err := deserializeFootprints(dump)
	if err != nil {
		t.Fatal(err)
	}

	if len(input) != len(got) {
		t.Errorf("expected %d footprints, got %d", len(input), len(got))
	}
}

func assertFontsetEquals(expected, got []Footprint) error {
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

func TestSerializeDeserialize(t *testing.T) {
	for _, fp := range []Footprint{
		{
			Family:  "a strange one",
			Runes:   newRuneSet(1, 0, 2, 0x789, 0xfffee),
			Scripts: ScriptSet{0, 1, 5, 0xffffff},
			Aspect:  font.Aspect{Style: 1, Weight: 200, Stretch: 0.45},
		},
		{
			Runes:   RuneSet{},
			Scripts: ScriptSet{},
		},
	} {
		b := fp.serializeTo(nil)

		var got Footprint
		n, err := got.deserializeFrom(b)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(b) {
			t.Fatalf("unexpected number of bytes read: %d", n)
		}

		if !reflect.DeepEqual(got, fp) {
			t.Fatalf("unexepected Footprint: %v, expected %v", got, fp)
		}
	}
}

func randomBytes() []byte {
	out := make([]byte, 1000)
	rand.Read(out)
	return out
}

func TestDeserializeInvalid(t *testing.T) {
	for range [50]int{} {
		src := randomBytes()
		if rand.Intn(2) == 0 { // indicate a small string
			binary.BigEndian.PutUint16(src, 10)
		}
		if rand.Intn(2) == 0 { // indicate no string and no rune set
			binary.BigEndian.PutUint16(src, 0)
			binary.BigEndian.PutUint32(src[2:], 0)
			src = src[:8] // truncate to simulate a broken input
		}
		var fp Footprint
		_, err := fp.deserializeFrom(src)
		if err == nil {
			t.Fatal("expected error on random input")
		}
	}
}

func TestSerializeSystemFonts(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	directories, err := DefaultFontDirectories(logger)
	if err != nil {
		t.Fatal(err)
	}

	fontset, err := scanFontFootprints(logger, nil, directories...)
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
