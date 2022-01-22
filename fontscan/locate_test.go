package fontscan

import "testing"

func Test_FindFont(t *testing.T) {
	for _, family := range [...]string{
		"arial", "times", "deja vu",
	} {
		_, err := FindFont(family)
		if err != nil {
			t.Fatal(err)
		}
	}
}
