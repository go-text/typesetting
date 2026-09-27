package fontscan

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

func TestFontPageFromOs2(t *testing.T) {
	// minimal OS/2 table, version 0, with the simplified Arabic font page in fsSelection
	raw := make([]byte, 78)
	binary.BigEndian.PutUint16(raw[62:], uint16(tables.FPSimpArabic))
	tu.Assert(t, fontPageFromOs2(raw) == tables.FPSimpArabic)
	tu.Assert(t, fontPageFromOs2(nil) == tables.FPNone)
}

func TestLegacyArabicFootprintIncludesUnicodeCoverage(t *testing.T) {
	for _, page := range []tables.FontPage{tables.FPSimpArabic, tables.FPTradArabic} {
		t.Run(fmt.Sprintf("%x", page), func(t *testing.T) {
			os2 := make([]byte, 78)
			binary.BigEndian.PutUint16(os2[62:], uint16(page))
			// Microsoft symbol cmap, format 6, covering the legacy Arabic PUA glyphs.
			var cmap bytes.Buffer
			tu.AssertNoErr(t, binary.Write(&cmap, binary.BigEndian, []uint16{
				0, 1, 3, 0, 0, 12, // cmap header and encoding record
				6, 10 + 2*0x300, 0, 0xF000, 0x300, // subtable header
			}))
			for i := 0; i < 0x300; i++ {
				tu.AssertNoErr(t, binary.Write(&cmap, binary.BigEndian, uint16(i+1)))
			}
			data := ot.WriteOpentype([]ot.Table{
				{Tag: ot.MustNewTag("OS/2"), Content: os2},
				{Tag: ot.MustNewTag("cmap"), Content: cmap.Bytes()},
			}, ot.TrueType)
			ld, err := ot.NewLoader(bytes.NewReader(data))
			tu.AssertNoErr(t, err)
			fp, _, err := newFootprintFromLoader(ld, false, scanBuffer{})
			tu.AssertNoErr(t, err)
			tu.Assert(t, fp.Runes.Contains(0x0627))
			tu.Assert(t, fp.Runes.Contains(0xF000))
			tu.Assert(t, fp.Scripts.contains(language.Arabic))
		})
	}
}

func TestDefaultDirs(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	dirs, err := DefaultFontDirectories(logger)
	tu.AssertNoErr(t, err)
	fmt.Printf("Valid font directories:\n%v\n", dirs)
}

func TestScanFontFootprints(t *testing.T) {
	ti := time.Now()

	logger := log.New(io.Discard, "", 0)
	directories, err := DefaultFontDirectories(logger)
	tu.AssertNoErr(t, err)

	fontset, err := scanFontFootprints(logger, nil, directories...)
	tu.AssertNoErr(t, err)

	// Show some basic stats
	families := map[string]bool{}
	for _, font := range fontset.flatten() {
		if font.Runes.Len() == 0 {
			t.Fatalf("unexpected empty rune coverage for %s", font.Location.File)
		}
		families[font.Family] = true
	}

	fmt.Printf("Found %d fonts (%d families) in %s\n",
		len(fontset), len(families), time.Since(ti))
}

func BenchmarkScanFonts(b *testing.B) {
	logger := log.New(io.Discard, "", 0)
	directories, err := DefaultFontDirectories(logger)
	tu.AssertNoErr(b, err)

	for i := 0; i < b.N; i++ {
		_, _ = scanFontFootprints(logger, nil, directories...)
	}
}

func TestScanIncrementalNoOp(t *testing.T) {
	ti := time.Now()

	logger := log.New(io.Discard, "", 0)
	directories, err := DefaultFontDirectories(logger)
	tu.AssertNoErr(t, err)

	// first scan
	fontset, err := scanFontFootprints(logger, nil, directories...)
	tu.AssertNoErr(t, err)
	fmt.Printf("Initial scan time: %s\n", time.Since(ti))

	ti = time.Now()
	incremental, err := scanFontFootprints(logger, fontset, directories...)
	tu.AssertNoErr(t, err)
	fmt.Printf("Second scan time: %s\n", time.Since(ti))

	if err = assertFontsetEquals(fontset.flatten(), incremental.flatten()); err != nil {
		t.Fatalf("incremental scan not consistent with initial scan: %s", err)
	}
}

func copyFile(t *testing.T, srcName, dstName string) {
	t.Helper()

	dst, err := os.Create(dstName)
	tu.AssertNoErr(t, err)

	src, err := os.Open(srcName)
	tu.AssertNoErr(t, err)
	defer src.Close()

	_, err = io.Copy(dst, src)
	tu.AssertNoErr(t, err)
	if err = dst.Sync(); err != nil {
		t.Fatal(err)
	}

	if err = dst.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestScanIncrementalUpdate(t *testing.T) {
	dir := t.TempDir()
	copyFile(t, filepath.Join("..", "font", "testdata", "Amiri-Regular.ttf"), filepath.Join(dir, "font1.ttf"))

	// first scan
	logger := log.New(io.Discard, "", 0)
	fontset, err := scanFontFootprints(logger, nil, dir)
	tu.AssertNoErr(t, err)
	if len(fontset) != 1 {
		t.Fatalf("unexpected font set: %v", fontset)
	}

	time.Sleep(time.Millisecond * 10)

	// test adding a new file
	copyFile(t, filepath.Join("..", "font", "testdata", "Roboto-Regular.ttf"), filepath.Join(dir, "font2.ttf"))

	fontset2, err := scanFontFootprints(logger, fontset, dir)
	tu.AssertNoErr(t, err)
	if len(fontset2) != 2 {
		t.Fatalf("unexpected font set: %v", fontset)
	}

	time.Sleep(time.Millisecond * 10)

	// test updating an existing file
	copyFile(t, filepath.Join("..", "font", "testdata", "Roboto-Regular.ttf"), filepath.Join(dir, "font1.ttf"))

	fontset3, err := scanFontFootprints(logger, nil, dir)
	tu.AssertNoErr(t, err)
	if len(fontset3) != 2 {
		t.Fatalf("unexpected font set: %v", fontset)
	}
	if family := fontset3.flatten()[0].Family; family != "roboto" {
		t.Fatalf("unexpected family %s", family)
	}

	time.Sleep(time.Millisecond * 10)

	incremental, err := scanFontFootprints(logger, fontset2, dir)
	tu.AssertNoErr(t, err)
	if err = assertFontsetEquals(fontset3.flatten(), incremental.flatten()); err != nil {
		t.Fatalf("incremental scan not consistent with initial scan: %s", err)
	}

	// test removing a file
	if err = os.Remove(filepath.Join(dir, "font1.ttf")); err != nil {
		t.Fatal(err)
	}
	fontset4, err := scanFontFootprints(logger, fontset3, dir)
	tu.AssertNoErr(t, err)
	if len(fontset4) != 1 {
		t.Fatalf("unexpected font set: %v", fontset)
	}
}
