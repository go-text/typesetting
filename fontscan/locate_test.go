package fontscan

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func Test_FindFont(t *testing.T) {
	for _, family := range [...]string{
		"arial", "times", "deja vu",
	} {
		_, loc, err := FindFont(family, Aspect{})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(loc.File)

		_, loc, err = FindFont(family, Aspect{Style: fonts.StyleItalic})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(loc.File)
	}
}
