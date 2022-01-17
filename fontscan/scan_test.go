package fontscan

import (
	"fmt"
	"io"
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

	ti = time.Now()
	buildFootprintMods(fontset)
	fmt.Printf("tree build in %s\n", time.Since(ti))
}

func TestScanIncremental(t *testing.T) {
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

	if len(incremental) != 0 {
		t.Fatalf("expected empty result from incremental scan, got %v", incremental)
	}
}
