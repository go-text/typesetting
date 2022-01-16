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

	got, err := ScanFamilies(directories...)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d fonts in %s\n", len(got), time.Since(ti))
}
