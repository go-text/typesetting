package fontscan

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/shaping"
	tu "github.com/go-text/typesetting/testutils"
)

func ExampleFontMap_UseSystemFonts() {
	fontMap := NewFontMap(log.Default())
	fontMap.UseSystemFonts("cachdir") // error handling omitted

	// set the font description
	fontMap.SetQuery(Query{Families: []string{"Arial", "serif"}}) // regular Aspect
	// `fontMap` is now ready for text shaping, using the `ResolveFace` method
}

func ExampleFontMap_AddFont() {
	// Open an on-disk font file. Do not close it, as the fontMap will need to parse
	// it on-demand. If you need to close it, read all of the bytes into a bytes.Reader
	// first.
	fontFile, _ := os.Open("myFont.ttf") // error handling omitted

	fontMap := NewFontMap(log.Default())
	fontMap.AddFont(fontFile, "myFont.ttf", "My Font") // error handling omitted

	// set the font description
	fontMap.SetQuery(Query{Families: []string{"Arial", "serif"}}) // regular Aspect

	// `fontMap` is now ready for text shaping, using the `ResolveFace` method
}

func ExampleFontMap_AddFace() {
	// Open an on-disk font file.
	fontFile, _ := os.Open("myFont.ttf") // error handling omitted
	defer fontFile.Close()

	// Load it and its metadata.
	ld, _ := ot.NewLoader(fontFile) // error handling omitted
	f, _ := font.NewFont(ld)        // error handling omitted
	md := f.Describe()
	fontMap := NewFontMap(log.Default())
	fontMap.AddFace(font.NewFace(f), Location{File: fmt.Sprint(md)}, md)

	// set the font description
	fontMap.SetQuery(Query{Families: []string{"Arial", "serif"}}) // regular Aspect

	// `fontMap` is now ready for text shaping, using the `ResolveFace` method
}

var _ shaping.Fontmap = (*FontMap)(nil)

func TestResolveFont(t *testing.T) {
	en, _ := NewLangID("en")

	var logOutput bytes.Buffer
	logger := log.New(&logOutput, "", 0)
	fm := NewFontMap(logger)

	tu.AssertC(t, fm.ResolveFace(0x20) == nil, "expected no face found in an empty FontMap")
	tu.AssertC(t, fm.ResolveFaceForLang(en) == nil, "expected no face found in an empty FontMap")

	err := fm.UseSystemFonts(t.TempDir())
	tu.AssertNoErr(t, err)

	logOutput.Reset()

	fm.SetQuery(Query{Families: []string{"helvetica"}, Aspect: font.Aspect{Weight: font.WeightBold}})
	foundFace := map[*font.Face]bool{}
	for _, r := range "Hello " + "تثذرزسشص" + "world" + "لمنهويء" {
		face := fm.ResolveFace(r)
		if face == nil {
			t.Fatalf("missing font for rune 0x%X", r)
		}
		foundFace[face] = true
	}
	fmt.Println(len(foundFace), "faces used")

	if logOutput.Len() != 0 {
		t.Fatalf("unexpected logs\n%s", logOutput.String())
	}

	logOutput.Reset()
	existingFamily := fm.database[0].Family
	fm.SetQuery(Query{Families: []string{existingFamily}})
	for _, r := range "Hello world" {
		face := fm.ResolveFace(r)
		if face == nil {
			t.Fatalf("missing font for rune 0x%X", r)
		}
	}
	if logOutput.Len() != 0 {
		t.Fatalf("unexpected logs\n%s", logOutput.String())
	}
}

func TestResolveForLang(t *testing.T) {
	fm := NewFontMap(log.New(io.Discard, "", 0))

	err := fm.UseSystemFonts(t.TempDir())
	tu.AssertNoErr(t, err)

	fm.SetQuery(Query{Families: []string{"helvetica"}})

	// all system fonts should have support for english
	en, _ := NewLangID("en")
	face := fm.ResolveFaceForLang(en)
	tu.AssertC(t, face != nil, "expected EN to be supported by system fonts")
}

func TestResolveFallbackManual(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	fm := NewFontMap(logger)

	file1, err := os.Open("../font/testdata/Amiri-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file1.Close()

	file2, err := os.Open("../font/testdata/Roboto-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file2.Close()

	err = fm.AddFont(file1, "user:Amiri", "")
	tu.AssertNoErr(t, err)
	err = fm.AddFont(file2, "user:Roboto", "")
	tu.AssertNoErr(t, err)

	fm.SetQuery(Query{}) // no families
	face := fm.ResolveFace('c')
	tu.Assert(t, fm.FontLocation(face.Font).File == "user:Amiri")

	en, _ := NewLangID("en")
	face = fm.ResolveFaceForLang(en)
	tu.Assert(t, face != nil && fm.FontLocation(face.Font).File == "user:Amiri")
}

func TestRevolveFamilyConflict(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	fm := NewFontMap(logger)

	err := fm.UseSystemFonts(t.TempDir())
	tu.AssertNoErr(t, err)

	file1, err := os.Open("../font/testdata/Amiri-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file1.Close()

	// This test is effective on platforms with an Arimo font
	fm.AddFont(file1, "user:amiri", "Arimo")

	fm.SetQuery(Query{Families: []string{"Arimo"}})
	tu.Assert(t, fm.FontLocation(fm.ResolveFace('a').Font).File == "user:amiri")
}

func BenchmarkResolveFont(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	fm := NewFontMap(logger)

	err := fm.UseSystemFonts(b.TempDir())
	tu.AssertNoErr(b, err)

	fm.SetQuery(Query{Families: []string{"helvetica"}, Aspect: font.Aspect{Weight: font.WeightBold}})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, r := range "Hello " + "تثذرزسشص" + "world" + "لمنهويء" {
			font := fm.ResolveFace(r)
			tu.AssertC(b, font != nil, fmt.Sprintf("missing font for rune 0x%X", r))
		}
	}
}

func BenchmarkSetQuery(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	fm := NewFontMap(logger)

	err := fm.UseSystemFonts(b.TempDir())
	tu.AssertNoErr(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fm.SetQuery(Query{
			Families: []string{"helvetica", "DejaVu", "monospace"},
			Aspect:   font.Aspect{Style: font.StyleItalic, Weight: font.WeightBold},
		})
	}
}

func Test_refreshSystemFontsIndex(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "fonts.cache")

	logger := log.New(io.Discard, "", 0)
	_, err := refreshSystemFontsIndex(logger, cachePath)
	tu.AssertNoErr(t, err)

	ti := time.Now()
	_, err = refreshSystemFontsIndex(logger, cachePath)
	tu.AssertNoErr(t, err)

	fmt.Printf("cache refresh in %s\n", time.Since(ti))
}

func TestInitSystemFonts(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	err := initSystemFonts(logger, t.TempDir())
	tu.AssertNoErr(t, err)

	tu.AssertC(t, len(systemFonts.flatten()) != 0, "systemFonts should not be empty")
}

func TestSystemFonts(t *testing.T) {
	fonts, err := SystemFonts(nil, t.TempDir())
	tu.AssertNoErr(t, err)

	tu.AssertC(t, len(fonts) != 0, "systemFonts should not be empty")
}

func TestFontMap_AddFont_FaceLocation(t *testing.T) {
	file1, err := os.Open("../font/testdata/Amiri-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file1.Close()

	file2, err := os.Open("../font/testdata/Roboto-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file2.Close()

	logger := log.New(io.Discard, "", 0)
	fm := NewFontMap(logger)

	if err = fm.AddFont(file1, "Amiri", ""); err != nil {
		t.Fatal(err)
	}
	tu.AssertC(t, len(fm.faceCache) == 1, fmt.Sprintf("unexpected face cache %d", len(fm.faceCache)))

	if err = fm.AddFont(file2, "Roboto", ""); err != nil {
		t.Fatal(err)
	}
	tu.AssertC(t, len(fm.faceCache) == 2, fmt.Sprintf("unexpected face cache %d", len(fm.faceCache)))

	loc1, loc2 := Location{File: "Amiri"}, Location{File: "Roboto"}
	face1, face2 := fm.faceCache[loc1], fm.faceCache[loc2]
	tu.Assert(t, fm.FontLocation(face1.Font) == loc1)
	tu.Assert(t, fm.FontLocation(face2.Font) == loc2)

	// try with an "invalid" face
	tu.Assert(t, fm.FontLocation(nil) == Location{})

	err = fm.AddFont(file1, "Roboto2", "MyRoboto")
	tu.AssertNoErr(t, err)

	fm.SetQuery(Query{Families: []string{"MyRoboto"}})
	face := fm.ResolveFace(0x20)
	tu.Assert(t, fm.FontLocation(face.Font).File == "Roboto2")
}

func TestQueryHelveticaLinux(t *testing.T) {
	// This is a regression test which asserts that
	// our behavior is similar than fontconfig

	file1, err := os.Open("../font/testdata/Amiri-Regular.ttf")
	tu.AssertNoErr(t, err)
	defer file1.Close()

	fm := NewFontMap(nil)
	err = fm.AddFont(file1, "file1", "Nimbus Sans")
	tu.AssertNoErr(t, err)

	err = fm.AddFont(file1, "file2", "Bitstream Vera Sans")
	tu.AssertNoErr(t, err)

	fm.SetQuery(Query{Families: []string{
		"BlinkMacSystemFont", // 'unknown' family
		"Helvetica",
	}})
	family, _ := fm.FontMetadata(fm.ResolveFace('x').Font)
	tu.Assert(t, family == font.NormalizeFamily("Nimbus Sans")) // prefered Helvetica replacement
}

func TestFindSytemFont(t *testing.T) {
	fm := NewFontMap(log.New(io.Discard, "", 0))
	_, ok := fm.FindSystemFont("Nimbus")
	tu.Assert(t, !ok) // no match on an empty fontmap

	// simulate system fonts
	fm.appendFootprints(Footprint{
		Family:   font.NormalizeFamily("Nimbus"),
		Location: Location{File: "nimbus.ttf"},
	},
		Footprint{
			Family:         font.NormalizeFamily("Noto Sans"),
			Location:       Location{File: "noto.ttf"},
			isUserProvided: true,
		},
	)

	nimbus, ok := fm.FindSystemFont("Nimbus")
	tu.Assert(t, ok && nimbus.File == "nimbus.ttf")

	_, ok = fm.FindSystemFont("nimbus ")
	tu.Assert(t, ok)

	_, ok = fm.FindSystemFont("Arial")
	tu.Assert(t, !ok)

	_, ok = fm.FindSystemFont("Noto Sans")
	tu.Assert(t, !ok) // user provided font are ignored
}
