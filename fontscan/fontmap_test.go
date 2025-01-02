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

func TestResolve_ScriptBengali(t *testing.T) {
	// we need a valid font file so that [ResolveFace] works;
	// the actual content of the file is not used
	dummyTestFont := "../font/testdata/Amiri-Regular.ttf"

	// this sample fontset is build from a "typical linux" system,
	// looking for Lohit-Bengali, NimbusSans, and Lohit-Devanagari
	// families
	bengaliFontSet := fontSet{
		{
			Location: Location{File: dummyTestFont},
			Family:   "nimbussans",
			Runes: RuneSet{
				{0x0, pageSet{0x0, 0xffffffff, 0xffffffff, 0x7fffffff, 0x0, 0xffffffff, 0xffffffff, 0xffffffff}},
				{0x1, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff, 0x40000, 0x0, 0x0, 0xfc000000}},
				{0x2, pageSet{0xf000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3f0002c0, 0x0}},
				{0x3, pageSet{0x0, 0x0, 0x0, 0x0, 0xffffd7f0, 0xfffffffb, 0x627fff, 0x0}},
				{0x4, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0x3c000c, 0x3fcf0000, 0xfcfcc0f, 0x3009801, 0xc30c}},
				{0x1e, pageSet{0x0, 0x0, 0x0, 0x0, 0x3f, 0x0, 0x0, 0xc0000}},
				{0x20, pageSet{0x7fb80004, 0x560d0047, 0x10, 0x83f10000, 0x0, 0x9098, 0x20000000, 0x0}},
				{0x21, pageSet{0x514e8020, 0xe0e145, 0x78000000, 0x0, 0x3ff0000, 0x200100, 0x3f0050, 0x0}},
				{0x22, pageSet{0xe6aeabed, 0xb04fa9, 0x120, 0xc37, 0x3e000fc, 0x800003c, 0x0, 0x0}},
				{0x23, pageSet{0x10004, 0x603, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0x25, pageSet{0x11111005, 0x10101010, 0xffff0000, 0x1ffff, 0xf1111, 0x96241c03, 0x3008cd8, 0x40}},
				{0x26, pageSet{0x0, 0x1c000000, 0x5, 0xc69, 0x0, 0x0, 0x0, 0x0}},
				{0x30, pageSet{0xc000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xef, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000, 0xffffffff, 0xfc001fff}},
				{0xfb, pageSet{0x1f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xff, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000}},
			},
			Scripts: ScriptSet{0x4379726c, 0x4772656b, 0x4c61746e, 0x5a696e68, 0x5a797979, 0x5a7a7a7a},
			Langs:   LangSet{0x7e2e743b4227384d, 0x9803d4f39f0115fc, 0x8b176fdad2678cfb, 0xa7b822ffbb0b91f0, 0x418f794, 0x0, 0x0, 0x0},
			Aspect:  font.Aspect{Style: 0x1, Weight: 700, Stretch: 1},
		},
		{
			Location: Location{File: dummyTestFont},
			Family:   "nimbussans",
			Runes: RuneSet{
				{0x0, pageSet{0x0, 0xffffffff, 0xffffffff, 0x7fffffff, 0x0, 0xffffffff, 0xffffffff, 0xffffffff}},
				{0x1, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff, 0x40000, 0x0, 0x0, 0xfc000000}},
				{0x2, pageSet{0xf000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3f0002c0, 0x0}},
				{0x3, pageSet{0x0, 0x0, 0x0, 0x0, 0xffffd7f0, 0xfffffffb, 0x627fff, 0x0}},
				{0x4, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0x3c000c, 0x3fcf0000, 0xfcfcc0f, 0x3009801, 0xc30c}},
				{0x1e, pageSet{0x0, 0x0, 0x0, 0x0, 0x3f, 0x0, 0x0, 0xc0000}},
				{0x20, pageSet{0x7fb80004, 0x560d0047, 0x10, 0x83f10000, 0x0, 0x9098, 0x20000000, 0x0}},
				{0x21, pageSet{0x514e8020, 0xe0e145, 0x78000000, 0x0, 0x3ff0000, 0x200100, 0x3f0050, 0x0}},
				{0x22, pageSet{0xe6aeabed, 0xb04fa9, 0x120, 0xc37, 0x3e000fc, 0x800003c, 0x0, 0x0}},
				{0x23, pageSet{0x10004, 0x603, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0x25, pageSet{0x11111005, 0x10101010, 0xffff0000, 0x1ffff, 0xf1111, 0x96241c03, 0x3008cd8, 0x40}},
				{0x26, pageSet{0x0, 0x1c000000, 0x5, 0xc69, 0x0, 0x0, 0x0, 0x0}},
				{0x30, pageSet{0xc000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xef, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000, 0xffffffff, 0xfc001fff}},
				{0xfb, pageSet{0x1f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xff, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000}},
			},
			Scripts: ScriptSet{0x4379726c, 0x4772656b, 0x4c61746e, 0x5a696e68, 0x5a797979, 0x5a7a7a7a},
			Langs:   LangSet{0x7e2e743b4227384d, 0x9803d4f39f0115fc, 0x8b176fdad2678cfb, 0xa7b822ffbb0b91f0, 0x418f794, 0x0, 0x0, 0x0},
			Aspect:  font.Aspect{Style: 0x2, Weight: 700, Stretch: 1},
		},
		{
			Location: Location{File: dummyTestFont},
			Family:   "nimbussans",
			Runes: RuneSet{
				{0x0, pageSet{0x0, 0xffffffff, 0xffffffff, 0x7fffffff, 0x0, 0xffffffff, 0xffffffff, 0xffffffff}},
				{0x1, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff, 0x40000, 0x0, 0x0, 0xfc000000}},
				{0x2, pageSet{0xf000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3f0002c0, 0x0}},
				{0x3, pageSet{0x0, 0x0, 0x0, 0x0, 0xffffd7f0, 0xfffffffb, 0x627fff, 0x0}},
				{0x4, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0x3c000c, 0x3fcf0000, 0xfcfcc0f, 0x3009801, 0xc30c}},
				{0x1e, pageSet{0x0, 0x0, 0x0, 0x0, 0x3f, 0x0, 0x0, 0xc0000}},
				{0x20, pageSet{0x7fb80004, 0x560d0047, 0x10, 0x83f10000, 0x0, 0x9098, 0x20000000, 0x0}},
				{0x21, pageSet{0x514e8020, 0xe0e145, 0x78000000, 0x0, 0x3ff0000, 0x200100, 0x3f0050, 0x0}},
				{0x22, pageSet{0xe6aeabed, 0xb04fa9, 0x120, 0xc37, 0x3e000fc, 0x800003c, 0x0, 0x0}},
				{0x23, pageSet{0x10004, 0x603, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0x25, pageSet{0x11111005, 0x10101010, 0xffff0000, 0x1ffff, 0xf1111, 0x96241c03, 0x3008cd8, 0x40}},
				{0x26, pageSet{0x0, 0x1c000000, 0x5, 0xc69, 0x0, 0x0, 0x0, 0x0}},
				{0x30, pageSet{0xc000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xef, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000, 0xffffffff, 0xfc001fff}},
				{0xfb, pageSet{0x1f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xff, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000}},
			},
			Scripts: ScriptSet{0x4379726c, 0x4772656b, 0x4c61746e, 0x5a696e68, 0x5a797979, 0x5a7a7a7a},
			Langs:   LangSet{0x7e2e743b4227384d, 0x9803d4f39f0115fc, 0x8b176fdad2678cfb, 0xa7b822ffbb0b91f0, 0x418f794, 0x0, 0x0, 0x0},
			Aspect:  font.Aspect{Style: 0x2, Weight: 400, Stretch: 1},
		},
		{
			Location: Location{File: dummyTestFont},
			Family:   "nimbussans",
			Runes: RuneSet{
				{0x0, pageSet{0x0, 0xffffffff, 0xffffffff, 0x7fffffff, 0x0, 0xffffffff, 0xffffffff, 0xffffffff}},
				{0x1, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff, 0x40000, 0x0, 0x0, 0xfc000000}},
				{0x2, pageSet{0xf000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3f0002c0, 0x0}},
				{0x3, pageSet{0x0, 0x0, 0x0, 0x0, 0xffffd7f0, 0xfffffffb, 0x627fff, 0x0}},
				{0x4, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0x3c000c, 0x3fcf0000, 0xfcfcc0f, 0x3009801, 0xc30c}},
				{0x1e, pageSet{0x0, 0x0, 0x0, 0x0, 0x3f, 0x0, 0x0, 0xc0000}},
				{0x20, pageSet{0x7fb80004, 0x560d0047, 0x10, 0x83f10000, 0x0, 0x9098, 0x20000000, 0x0}},
				{0x21, pageSet{0x514e8020, 0xe0e145, 0x78000000, 0x0, 0x3ff0000, 0x200100, 0x3f0050, 0x0}},
				{0x22, pageSet{0xe6aeabed, 0xb04fa9, 0x120, 0xc37, 0x3e000fc, 0x800003c, 0x0, 0x0}},
				{0x23, pageSet{0x10004, 0x603, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0x25, pageSet{0x11111005, 0x10101010, 0xffff0000, 0x1ffff, 0xf1111, 0x96241c03, 0x3008cd8, 0x40}},
				{0x26, pageSet{0x0, 0x1c000000, 0x5, 0xc69, 0x0, 0x0, 0x0, 0x0}},
				{0x30, pageSet{0xc000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xef, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000, 0xffffffff, 0xfc001fff}},
				{0xfb, pageSet{0x1f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xff, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000}},
			},
			Scripts: ScriptSet{0x4379726c, 0x4772656b, 0x4c61746e, 0x5a696e68, 0x5a797979, 0x5a7a7a7a},
			Langs:   LangSet{0x7e2e743b4227384d, 0x9803d4f39f0115fc, 0x8b176fdad2678cfb, 0xa7b822ffbb0b91f0, 0x418f794, 0x0, 0x0, 0x0},
			Aspect:  font.Aspect{Style: 0x1, Weight: 400, Stretch: 1},
		},
		{
			Location: Location{File: dummyTestFont},
			Family:   "lohitbengali",
			Runes: RuneSet{
				{0x0, pageSet{0x0, 0xffffffff, 0xf8000001, 0x78000001, 0x0, 0x4, 0x800000, 0x800000}},
				{0x9, pageSet{0x0, 0x0, 0x0, 0x30, 0xfff99fef, 0xf3fdfdff, 0xb0807f9f, 0xfffffcf}},
				{0x20, pageSet{0x3ff83000, 0x40, 0x0, 0x0, 0x0, 0x2000000, 0x0, 0x0}},
				{0x22, pageSet{0x40000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0x25, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1000, 0x0}},
				{0xff, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000}},
			},
			Scripts: ScriptSet{0x42656e67, 0x5a696e68, 0x5a797979, 0x5a7a7a7a},
			Langs:   LangSet{0x10000200, 0x0, 0x4000000, 0x0, 0x0, 0x0, 0x0, 0x0},
			Aspect:  font.Aspect{Style: 0x1, Weight: 400, Stretch: 1},
		},
		{
			Location: Location{File: dummyTestFont},
			Family:   "lohitdevanagari",
			Runes: RuneSet{
				{0x0, pageSet{0x0, 0xffffffff, 0xffffffff, 0x7fffffff, 0x0, 0xffffdffe, 0xffffffff, 0xffffffff}},
				{0x1, pageSet{0xcfcff0ff, 0xffffcf8f, 0xcfff31ff, 0x7f0fcc3f, 0x40000, 0x0, 0x0, 0x0}},
				{0x2, pageSet{0xf000000, 0x800000, 0x0, 0x0, 0x0, 0x10000000, 0x3f0000c0, 0x0}},
				{0x3, pageSet{0x80, 0x40, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0}},
				{0x9, pageSet{0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff, 0x0, 0x0, 0x0, 0x0}},
				{0x1c, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x600000}},
				{0x20, pageSet{0xfff83000, 0x601007f, 0x10, 0x0, 0x1f, 0x2001000, 0x0, 0x0}},
				{0x21, pageSet{0x80000, 0x407c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0x22, pageSet{0x7c07807c, 0x800, 0x100, 0x30, 0x0, 0x0, 0x0, 0x0}},
				{0x25, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1c00, 0x0}},
				{0xa8, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3fffffff}},
				{0xfb, pageSet{0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{0xff, pageSet{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80000000}},
			},
			Scripts: ScriptSet{0x44657661, 0x4772656b, 0x4c61746e, 0x5a696e68, 0x5a797979, 0x5a7a7a7a},
			Langs:   LangSet{0x743c7429c3c430c9, 0x882084e39fe115fc, 0x8b066ff86023ecc8, 0x13b880fb9b021b74, 0x418d794, 0x0, 0x0, 0x0},
			Aspect:  font.Aspect{Style: 0x1, Weight: 400, Stretch: 1},
		},
	}

	fm := NewFontMap(log.New(io.Discard, "", 0))
	fm.appendFootprints(bengaliFontSet...)

	// make sure the same font is selected for a given script, when possible
	text := []rune("হয় না।")
	fm.SetQuery(Query{Families: []string{"Nimbus Sans"}})
	runs := (&shaping.Segmenter{}).Split(shaping.Input{Text: text, RunEnd: len(text)}, fm)
	tu.Assert(t, len(runs) == 1)
	// only one font is loaded, so there is no clash
	family, _ := fm.FontMetadata(runs[0].Face.Font)
	tu.Assert(t, family == "lohitbengali")
}
