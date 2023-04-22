package fontscan

import (
	"fmt"
	"testing"
	"time"

	meta "github.com/go-text/typesetting/opentype/api/metadata"
)

func Test_FindFont(t *testing.T) {
	for _, family := range [...]string{
		"arial", "times", "deja vu",
	} {
		ti := time.Now()
		_, loc, err := FindFont(family, meta.Aspect{})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("found", loc.File, "in", time.Since(ti))

		_, loc, err = FindFont(family, meta.Aspect{Style: meta.StyleItalic})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(loc.File)
	}
}
