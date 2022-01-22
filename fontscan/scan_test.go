package fontscan

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultDirs(t *testing.T) {
	dirs, err := DefaultFontDirectories()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Valid font directories:\n%v\n", dirs)
}

func TestScanFontFiles(t *testing.T) {
	ti := time.Now()

	directories, err := DefaultFontDirectories()
	if err != nil {
		t.Fatal(err)
	}

	fontpaths, err := scanFontFiles(directories...)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Found %d fonts in %s\n", len(fontpaths), time.Since(ti))
}

func TestScanFontFootprints(t *testing.T) {
	ti := time.Now()

	directories, err := DefaultFontDirectories()
	if err != nil {
		t.Fatal(err)
	}

	fontset, err := scanFontFootprints(nil, directories...)
	if err != nil {
		t.Fatal(err)
	}

	// Show some basic stats
	distribution := map[fontFormat]int{}
	for _, font := range fontset.flatten() {
		if font.Runes.Len() == 0 {
			t.Fatalf("unexpected empty rune coverage for %s", font.Location.File)
		}
		distribution[font.Format]++
	}

	fmt.Printf("Found %d fonts in %s (distribution: %v)\n", len(fontset), time.Since(ti), distribution)
}

func TestScanIncrementalNoOp(t *testing.T) {
	ti := time.Now()

	directories, err := DefaultFontDirectories()
	if err != nil {
		t.Fatal(err)
	}

	// first scan
	fontset, err := scanFontFootprints(nil, directories...)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Initial scan time: %s\n", time.Since(ti))

	ti = time.Now()
	incremental, err := scanFontFootprints(fontset, directories...)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Second scan time: %s\n", time.Since(ti))

	if err = assertFontsetEquals(fontset.flatten(), incremental.flatten()); err != nil {
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
	dir := t.TempDir()
	copyFile(t, filepath.Join("..", "font", "testdata", "Amiri-Regular.ttf"), filepath.Join(dir, "font1.ttf"))

	// first scan
	fontset, err := scanFontFootprints(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset) != 1 {
		t.Fatalf("unexpected font set: %v", fontset)
	}

	// test adding a new file
	copyFile(t, filepath.Join("..", "font", "testdata", "Roboto-Regular.ttf"), filepath.Join(dir, "font2.ttf"))

	fontset2, err := scanFontFootprints(fontset, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset2) != 2 {
		t.Fatalf("unexpected font set: %v", fontset)
	}

	// test updating an existing file
	copyFile(t, filepath.Join("..", "font", "testdata", "Roboto-Regular.ttf"), filepath.Join(dir, "font1.ttf"))

	fontset3, err := scanFontFootprints(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset3) != 2 {
		t.Fatalf("unexpected font set: %v", fontset)
	}
	if family := fontset3.flatten()[0].Family; family != "roboto" {
		t.Fatalf("unexpected family %s", family)
	}

	incremental, err := scanFontFootprints(fontset2, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err = assertFontsetEquals(fontset3.flatten(), incremental.flatten()); err != nil {
		t.Fatalf("incremental scan not consistent with initial scan: %s", err)
	}

	// test removing a file
	if err = os.Remove(filepath.Join(dir, "font1.ttf")); err != nil {
		t.Fatal(err)
	}
	fontset4, err := scanFontFootprints(fontset3, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(fontset4) != 1 {
		t.Fatalf("unexpected font set: %v", fontset)
	}
}
