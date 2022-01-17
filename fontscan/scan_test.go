package fontscan

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanFamilies(t *testing.T) {
	ti := time.Now()

	directories, err := DefaultFontDirs()
	if err != nil {
		t.Fatal(err)
	}

	// simulate a duplicate directory entry
	directories = append(directories, directories...)

	got, err := ScanFamilies(directories...)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Found %d fonts in %s\n", len(got), time.Since(ti))
}

func TestScanFonts(t *testing.T) {
	Warning.SetOutput(io.Discard)
	ti := time.Now()

	directories, err := DefaultFontDirs()
	if err != nil {
		t.Fatal(err)
	}

	fontset, err := ScanFonts(nil, directories...)
	if err != nil {
		t.Fatal(err)
	}

	// Show some basic stats
	distribution := map[Format]int{}
	for _, font := range fontset {
		distribution[font.Format]++
	}

	fmt.Printf("Found %d fonts in %s ( distribution: %v)\n", len(fontset), time.Since(ti), distribution)
}

func TestScanIncrementalNoOp(t *testing.T) {
	Warning.SetOutput(io.Discard)

	ti := time.Now()

	directories, err := DefaultFontDirs()
	if err != nil {
		t.Fatal(err)
	}

	// first scan
	fontset, err := ScanFonts(nil, directories...)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Initial scan time: %s\n", time.Since(ti))

	ti = time.Now()
	incremental, err := ScanFonts(fontset, directories...)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Second scan time: %s\n", time.Since(ti))

	if err = assertFontsetEquals(fontset, incremental); err != nil {
		t.Fatalf("incremental scan not consistent with initial scan: %s", err)
	}
}

func copyFile(t *testing.T, srcName, dstName string) {
	t.Helper()

	dst, err := os.Create(dstName)
	if err != nil {
		t.Fatal(err)
	}

	src, err := os.Open(srcName)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		t.Fatal(err)
	}
	if err = dst.Sync(); err != nil {
		t.Fatal(err)
	}

	if err = dst.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestScanIncrementalUpdate(t *testing.T) {
	Warning.SetOutput(io.Discard)

	dir := t.TempDir()
	copyFile(t, filepath.Join("..", "font", "testdata", "Amiri-Regular.ttf"), filepath.Join(dir, "font1.ttf"))

	// first scan
	fontset, err := ScanFonts(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset) != 1 {
		t.Fatalf("unexpected font set: %v", fontset)
	}

	// test adding a new file
	copyFile(t, filepath.Join("..", "font", "testdata", "Roboto-Regular.ttf"), filepath.Join(dir, "font2.ttf"))

	fontset2, err := ScanFonts(fontset, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset2) != 2 {
		t.Fatalf("unexpected font set: %v", fontset)
	}

	// test updating an existing file
	copyFile(t, filepath.Join("..", "font", "testdata", "Roboto-Regular.ttf"), filepath.Join(dir, "font1.ttf"))

	fontset3, err := ScanFonts(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset3) != 2 {
		t.Fatalf("unexpected font set: %v", fontset)
	}
	if family := fontset3[0].Family; family != "Roboto" {
		t.Fatalf("unexpected family %s", family)
	}

	incremental, err := ScanFonts(fontset2, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err = assertFontsetEquals(fontset3, incremental); err != nil {
		t.Fatalf("incremental scan not consistent with initial scan: %s", err)
	}

	// test removing a file
	if err = os.Remove(filepath.Join(dir, "font1.ttf")); err != nil {
		t.Fatal(err)
	}
	fontset4, err := ScanFonts(fontset3, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset4) != 1 {
		t.Fatalf("unexpected font set: %v", fontset)
	}
}
