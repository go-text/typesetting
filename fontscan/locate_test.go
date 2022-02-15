package fontscan

import (
	"fmt"
	"testing"
	"time"

	"github.com/benoitkugler/textlayout/fonts"
)

func Test_FindFont(t *testing.T) {
	for _, family := range [...]string{
		"arial", "times", "deja vu",
	} {
		ti := time.Now()
		_, loc, err := FindFont(family, Aspect{})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("found", loc.File, "in", time.Since(ti))

		_, loc, err = FindFont(family, Aspect{Style: fonts.StyleItalic})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(loc.File)
	}
}
