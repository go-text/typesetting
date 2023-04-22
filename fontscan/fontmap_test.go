package fontscan

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-text/typesetting/font"
	meta "github.com/go-text/typesetting/opentype/api/metadata"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

func TestResolveFont(t *testing.T) {
	fm := NewFontMap()

	tu.AssertC(t, fm.ResolveFont(0x20) == nil, "expected no face found in an empty FontMap")

	err := fm.UseSystemFonts()
	tu.AssertNoErr(t, err)

	var logOutput bytes.Buffer
	log.Default().SetOutput(&logOutput)

	fm.SetQuery(Query{Families: []string{"helvetica"}, Aspect: meta.Aspect{Weight: meta.WeightBold}})
	foundFace := map[font.Font]bool{}
	for _, r := range "Hello " + "تثذرزسشص" + "world" + "لمنهويء" {
		face := fm.ResolveFont(r)
		if face == nil {
			t.Fatalf("missing font for rune 0x%X", r)
		}
		foundFace[face] = true
	}
	fmt.Println(len(foundFace), "faces used")

	if logOutput.Len() != 0 {
		t.Fatalf("unexpected logs %s", logOutput.String())
	}

	logOutput.Reset()
	existingFamily := fm.database[0].Family
	fm.SetQuery(Query{Families: []string{existingFamily}})
	for _, r := range "Hello world" {
		face := fm.ResolveFont(r)
		if face == nil {
			t.Fatalf("missing font for rune 0x%X", r)
		}
	}
	if logOutput.Len() != 0 {
		t.Fatalf("unexpected logs %s", logOutput.String())
	}
}

func BenchmarkResolveFont(b *testing.B) {
	fm := NewFontMap()

	err := fm.UseSystemFonts()
	tu.AssertNoErr(b, err)

	fm.SetQuery(Query{Families: []string{"helvetica"}, Aspect: meta.Aspect{Weight: meta.WeightBold}})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, r := range "Hello " + "تثذرزسشص" + "world" + "لمنهويء" {
			font := fm.ResolveFont(r)
			tu.AssertC(b, font != nil, fmt.Sprintf("missing font for rune 0x%X", r))
		}
	}
}

func BenchmarkSetQuery(b *testing.B) {
	fm := NewFontMap()

	err := fm.UseSystemFonts()
	tu.AssertNoErr(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fm.SetQuery(Query{
			Families: []string{"helvetica", "DejaVu", "monospace"},
			Aspect:   meta.Aspect{Style: meta.StyleItalic, Weight: meta.WeightBold},
		})
	}
}

func Test_refreshSystemFontsIndex(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "fonts.cache")

	_, err := refreshSystemFontsIndex(cachePath)
	tu.AssertNoErr(t, err)

	ti := time.Now()
	_, err = refreshSystemFontsIndex(cachePath)
	tu.AssertNoErr(t, err)

	fmt.Printf("cache refresh in %s\n", time.Since(ti))
}

func TestInitSystemFonts(t *testing.T) {
	err := initSystemFonts()
	tu.AssertNoErr(t, err)

	tu.AssertC(t, len(systemFonts.flatten()) != 0, "SystemFonts should not be empty")
}

func TestFontMap_AddFont_FaceLocation(t *testing.T) {
	file1, err := os.Open("../font/testdata/Amiri-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file1.Close()

	file2, err := os.Open("../font/testdata/Roboto-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file2.Close()

	fm := NewFontMap()

	if err = fm.AddFont(file1, "Amiri", ""); err != nil {
		t.Fatal(err)
	}
	tu.AssertC(t, len(fm.fonts) == 1, fmt.Sprintf("unexpected face cache %d", len(fm.fonts)))

	if err = fm.AddFont(file2, "Roboto", ""); err != nil {
		t.Fatal(err)
	}
	tu.AssertC(t, len(fm.fonts) == 2, fmt.Sprintf("unexpected face cache %d", len(fm.fonts)))

	loc1, loc2 := Location{File: "Amiri"}, Location{File: "Roboto"}
	face1, face2 := fm.fonts[loc1], fm.fonts[loc2]
	tu.Assert(t, fm.FontLocation(face1) == loc1)
	tu.Assert(t, fm.FontLocation(face2) == loc2)

	// try with an "invalid" face
	tu.Assert(t, fm.FontLocation(nil) == Location{})

	err = fm.AddFont(file1, "Roboto2", "MyRoboto")
	tu.AssertNoErr(t, err)

	fm.SetQuery(Query{Families: []string{"MyRoboto"}})
	face := fm.ResolveFont(0x20)
	tu.Assert(t, fm.FontLocation(face).File == "Roboto2")
}
