package fontscan

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestScanFamilies(t *testing.T) {
	ti := time.Now()

	directories, err := DefaultFontDirs()
	if err != nil {
		log.Fatal(err)
	}

	// simulate a duplicate directory entry
	directories = append(directories, directories...)

	got, err := ScanFamilies(directories...)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d fonts in %s\n", len(got), time.Since(ti))
}

func TestScanFonts(t *testing.T) {
	ti := time.Now()

	directories, err := DefaultFontDirs()
	if err != nil {
		log.Fatal(err)
	}

	fontset, err := ScanFonts(directories...)
	if err != nil {
		log.Fatal(err)
	}

	// Show some basic stats
	repartition := map[Format]int{}
	for _, font := range fontset {
		repartition[font.Format]++
	}

	fmt.Printf("Found %d fonts in %s (repartition: %v)\n", len(fontset), time.Since(ti), repartition)
}
