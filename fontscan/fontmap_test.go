package fontscan

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/go-text/typesetting/font"
)

func TestResolveFace(t *testing.T) {
	fm := NewFontMap()
	if err := fm.UseSystemFonts(); err != nil {
		t.Fatal(err)
	}

	fm.SetQuery(FontQuery{Families: []string{"helvetica"}, Aspect: Aspect{Weight: fonts.WeightBold}})
	foundFace := map[font.Face]bool{}
	for _, r := range "Hello " + "تثذرزسشص" + "world" + "لمنهويء" {
		face := fm.ResolveFace(r)
		if face == nil {
			t.Fatalf("missing font for rune 0x%X", r)
		}
		foundFace[face] = true
	}
	fmt.Println(len(foundFace), "faces used")

	existingFamily := fm.database[0].Family
	fm.SetQuery(FontQuery{Families: []string{existingFamily}})
	for _, r := range "Hello world" {
		face := fm.ResolveFace(r)
		if face == nil {
			t.Fatalf("missing font for rune 0x%X", r)
		}
	}
}

func Test_refreshSystemFontsIndex(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "fonts.cache")

	_, err := refreshSystemFontsIndex(cachePath)
	if err != nil {
		t.Fatal(err)
	}

	ti := time.Now()
	_, err = refreshSystemFontsIndex(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("cache refresh in %s\n", time.Since(ti))
}

func TestInitSystemFonts(t *testing.T) {
	err := initSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	if len(systemFonts.flatten()) == 0 {
		t.Fatal("SystemFonts should not be empty")
	}
}

func TestFontMap_AddFont_FaceLocation(t *testing.T) {
	file1, err := os.Open("../font/testdata/Amiri-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer file1.Close()

	file2, err := os.Open("../font/testdata/Roboto-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()

	fm := NewFontMap()

	if err = fm.AddFont(file1, "Amiri"); err != nil {
		t.Fatal(err)
	}
	if len(fm.faces) != 1 {
		t.Fatalf("unexpected face cache %d", len(fm.faces))
	}

	if err = fm.AddFont(file2, "Roboto"); err != nil {
		t.Fatal(err)
	}
	if len(fm.faces) != 2 {
		t.Fatalf("unexpected face cache %d", len(fm.faces))
	}

	loc1, loc2 := Location{File: "Amiri"}, Location{File: "Roboto"}
	face1, face2 := fm.faces[loc1], fm.faces[loc2]

	if got := fm.FaceLocation(face1); got != loc1 {
		t.Fatalf("FaceLocation: expected %v, got %v", loc1, got)
	}
	if got := fm.FaceLocation(face2); got != loc2 {
		t.Fatalf("FaceLocation: expected %v, got %v", loc2, got)
	}

	// try with an "invalid" face
	if got := fm.FaceLocation(nil); got != (Location{}) {
		t.Fatalf("FaceLocation: expected %v, got %v", Location{}, got)
	}
}
