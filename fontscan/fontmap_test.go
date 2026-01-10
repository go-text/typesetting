package fontscan

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/unidoc/typesetting/font"
	ot "github.com/unidoc/typesetting/font/opentype"
	"github.com/unidoc/typesetting/language"
	"github.com/unidoc/typesetting/shaping"
	tu "github.com/unidoc/typesetting/testutils"
)

func ExampleFontMap_UseSystemFonts() {
	fontMap := NewFontMap(log.Default())
	fontMap.UseSystemFonts("cachdir") // error handling omitted

	// set the font description
	fontMap.SetQuery(Query{Families: []string{"Arial", "serif"}}) // regular Aspect
	// set the script, if known in advance
	fontMap.SetScript(language.Latin)
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
	// set the script, if known in advance
	fontMap.SetScript(language.Latin)

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
	// set the script, if known in advance
	fontMap.SetScript(language.Latin)

	// `fontMap` is now ready for text shaping, using the `ResolveFace` method
}

var _ shaping.FontmapScript = (*FontMap)(nil)

func TestResolveFont(t *testing.T) {
	var logOutput bytes.Buffer
	logger := log.New(&logOutput, "", 0)
	fm := NewFontMap(logger)

	tu.AssertC(t, fm.ResolveFace(0x20) == nil, "expected no face found in an empty FontMap")
	tu.AssertC(t, fm.ResolveFaceForLang(language.LangEn) == nil, "expected no face found in an empty FontMap")

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
	face := fm.ResolveFaceForLang(language.LangEn)
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
	fm.SetScript(language.Latin)
	face := fm.ResolveFace('c')
	tu.Assert(t, fm.FontLocation(face.Font).File == "user:Amiri")

	face = fm.ResolveFaceForLang(language.LangEn)
	tu.Assert(t, face != nil && fm.FontLocation(face.Font).File == "user:Amiri")
}

func TestResolveLang(t *testing.T) {
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

	// with fallback
	face := fm.ResolveFaceForLang(language.LangEn)
	tu.Assert(t, face != nil && fm.FontLocation(face.Font).File == "user:Amiri")

	// exact
	fm.SetQuery(Query{Families: []string{"Roboto"}})
	face = fm.ResolveFaceForLang(language.LangEn)
	tu.Assert(t, face != nil && fm.FontLocation(face.Font).File == "user:Roboto")
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
	fm.SetScript(language.Latin)
	tu.Assert(t, fm.FontLocation(fm.ResolveFace('a').Font).File == "user:amiri")
}

func BenchmarkResolveFont(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	fm := NewFontMap(logger)

	err := fm.UseSystemFonts(b.TempDir())
	tu.AssertNoErr(b, err)

	fm.SetQuery(Query{Families: []string{"helvetica"}, Aspect: font.Aspect{Weight: font.WeightBold}})
	fm.SetScript(language.Latin)

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
	fm.SetScript(language.Latin)
	face := fm.ResolveFace(' ')
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
	fm.SetScript(language.Latin)
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

// the following tests use a "linux" font configuration
func newSampleFontmap() *FontMap {
	fm := NewFontMap(log.New(io.Discard, "", 0))
	fm.appendFootprints(linuxSampleFontSet...)
	for _, fp := range linuxSampleFontSet {
		fm.cache(fp, &font.Face{Font: new(font.Font)}) // we need a new pointer for each file
	}
	return fm
}

func TestDumpSystemFonts(t *testing.T) {
	t.Skip()
	fontset, err := SystemFonts(nil, os.TempDir())
	tu.AssertNoErr(t, err)

	var trimmed fontSet
	for _, fp := range fontset {
		switch fp.Family {
		case "nimbussans", "lohitbengali", "lohitdevanagari", "lohitodia",
			"notoloopedthai", "notosanskhmer", "khmeros", "khmerossystem",
			"freeserif", "freesans", "freemono", "dejavu", "dejavusans":
			trimmed = append(trimmed, fp)
		}
	}
	code := fmt.Sprintf(`
	package fontscan
	import "github.com/unidoc/typesetting/font"

	// extracted from a linux system
	var linuxSampleFontSet = 
	%#v`, trimmed)
	code = strings.ReplaceAll(code, "fontscan.", "")
	code = strings.ReplaceAll(code, "Footprint{", "\n{")
	code = strings.ReplaceAll(code, ", Index:0x0, Instance:0x0", "")
	code = strings.ReplaceAll(code, ", isUserProvided:false", "")
	code = strings.ReplaceAll(code, "Location:", "\nLocation:")
	code = strings.ReplaceAll(code, "Runes:", "\nRunes:")
	code = strings.ReplaceAll(code, "Langs:", "\nLangs:")
	code = strings.ReplaceAll(code, "Aspect:", "\nAspect:")

	err = os.WriteFile("fontmap_sample_test.go", []byte(code), os.ModePerm)
	tu.AssertNoErr(t, err)

	err = exec.Command("goimports", "-w", "fontmap_sample_test.go").Run()
	tu.AssertNoErr(t, err)
}

func TestResolve_ScriptBengali(t *testing.T) {
	fm := newSampleFontmap()

	// make sure the same font is selected for a given script, when possible
	text := []rune("হয় না।")
	fm.SetQuery(Query{Families: []string{"Nimbus Sans"}})
	runs := (&shaping.Segmenter{}).Split(shaping.Input{Text: text, RunEnd: len(text)}, fm)
	tu.Assert(t, len(runs) == 1)
	family, _ := fm.FontMetadata(runs[0].Face.Font)
	tu.Assert(t, family == "lohitbengali")
}

func TestResolve_ScriptThaana(t *testing.T) {
	fm := newSampleFontmap()

	// make sure the same font is selected for a given script, when possible
	text := []rune("އުފަންވަނީ، ދަރަޖަ")
	fm.SetQuery(Query{Families: []string{"Nimbus Sans"}})
	runs := (&shaping.Segmenter{}).Split(shaping.Input{Text: text, RunEnd: len(text)}, fm)
	tu.Assert(t, len(runs) == 1)
	family, _ := fm.FontMetadata(runs[0].Face.Font)
	tu.Assert(t, family == "freeserif")
	tu.Assert(t, strings.HasSuffix(fm.FontLocation(runs[0].Face.Font).File, "FreeSerif.ttf"))
}

func TestResolve_SciptGujarati(t *testing.T) {
	fm := newSampleFontmap()

	text := []rune("ମୁଁ କାଚ ଖାଇପାରେ ଏବଂ ତାହା ମୋର କ୍ଷତି କରିନଥାଏ।")
	fm.SetQuery(Query{Families: []string{"Nimbus Sans"}})
	runs := (&shaping.Segmenter{}).Split(shaping.Input{Text: text, RunEnd: len(text)}, fm)
	tu.Assert(t, len(runs) == 1)
	family, _ := fm.FontMetadata(runs[0].Face.Font)
	tu.Assert(t, family == "lohitodia")
}

func TestResolve_SciptArabic(t *testing.T) {
	fm := newSampleFontmap()

	text := []rune("میں کانچ کھا سکتا ہوں اور مجھے تکلیف نہیں ہوتی ۔")
	fm.SetQuery(Query{Families: []string{"Nimbus Sans"}})
	runs := (&shaping.Segmenter{}).Split(shaping.Input{Text: text, RunEnd: len(text)}, fm)
	tu.Assert(t, len(runs) == 10)
	family0, _ := fm.FontMetadata(runs[0].Face.Font)
	family1, _ := fm.FontMetadata(runs[1].Face.Font)
	tu.Assert(t, family0 == "dejavusans")
	tu.Assert(t, family1 == "freeserif")
}

func TestResolve_SciptKhmer(t *testing.T) {
	fm := newSampleFontmap()

	text := []rune("ខ្ញុំអាចញុំកញ្ចក់បាន ដោយគ្មានបញ្ហារ")
	fm.SetQuery(Query{Families: []string{"Nimbus Sans"}})
	runs := (&shaping.Segmenter{}).Split(shaping.Input{Text: text, RunEnd: len(text)}, fm)
	tu.Assert(t, len(runs) == 1)
	family, _ := fm.FontMetadata(runs[0].Face.Font)
	tu.Assert(t, family == "khmeros")
}
