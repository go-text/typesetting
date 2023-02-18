package harfbuzz

import (
	"testing"
)

func TestLanguageOrder(t *testing.T) {
	for i, l := range otLanguages {
		if i == 0 {
			continue
		}
		c := l.compare(otLanguages[i-1].language)
		if c > 0 {
			t.Fatalf("ot_languages not sorted at index %d: %s %d %s\n",
				i, otLanguages[i-1].language, c, l.language)
		}
	}
}

func TestFindLanguage(t *testing.T) {
	for _, l := range otLanguages {
		j := bfindLanguage(l.language)
		if j == -1 {
			t.Errorf("can't find back language %v", l)
		}
		// since there is some duplicate, we won't have i == j
		if otLanguages[j].language != l.language {
			t.Errorf("unexpected %s", otLanguages[j].language)
		}
	}
}
