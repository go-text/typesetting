package fontscan

import (
	"reflect"
	"testing"

	"github.com/benoitkugler/textlayout/language"
)

func Test_findExactReferenceLang(t *testing.T) {
	tests := []struct {
		args language.Language
		want int
	}{
		{"a", -1},       // -p-1 = 0
		{"ac", -3},      // -p-1 = 2
		{"ag", -4},      // -p-1 = 3
		{"ber-da", -16}, // -p-1 = 15
		{"zh-hl", -245}, // -p-1 = 244
		{"zz", -249},    // -p-1 = 248
	}
	// add all the exact match
	for id, entry := range referenceLangs {
		tests = append(tests, struct {
			args language.Language
			want int
		}{
			entry.language, id,
		})
	}

	for _, tt := range tests {
		if got := findExactReferenceLang(tt.args); got != tt.want {
			t.Errorf("findExactReferenceLang(%s) = %v, want %v", tt.args, got, tt.want)
		}
	}
}

func TestLangSet(t *testing.T) {
	tests := []struct {
		l        language.Language
		start    []language.Language
		expected []language.Language
	}{
		{
			"fr",
			nil,
			[]language.Language{"fr"},
		},
		{
			"de",
			[]language.Language{"fr"},
			[]language.Language{"de", "fr"},
		},
		{
			"xxx-a",
			[]language.Language{"fr"},
			[]language.Language{"fr"},
		},
		{
			"fr-be",
			[]language.Language{"fr"},
			[]language.Language{"fr"},
		},
	}
	for _, tt := range tests {
		ls := NewLangSet(tt.start...)
		ls.Add(tt.l)
		if langs := ls.Langs(); !reflect.DeepEqual(langs, tt.expected) {
			t.Fatalf("expected %v, got %v (%v)", tt.expected, langs, ls)
		}

		for _, l := range tt.expected {
			if ls.Contains(l) != language.LanguagesExactMatch {
				t.Fatalf("missing lang %s", l)
			}
		}

		ls.Delete(tt.l)
		if langs := ls.Langs(); !reflect.DeepEqual(langs, tt.start) {
			t.Fatalf("expected %v, got %v (%v)", tt.start, langs, ls)
		}

		if ls.Contains(tt.l) == language.LanguagesExactMatch {
			t.Fatalf("lang %s should be deleted", tt.l)
		}

		if ls.Contains("xxx-fr") != 0 {
			t.Fatalf("lang %s should be missing", "xxx-fr")
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		l    language.Language
		ls   []language.Language
		want language.LanguageComparison
	}{
		{
			"fr",
			nil,
			language.LanguagesDiffer,
		},
		{
			"de",
			[]language.Language{"fr"},
			language.LanguagesDiffer,
		},
		{
			"fr",
			[]language.Language{"fr"},
			language.LanguagesExactMatch,
		},
		{
			"fr-be",
			[]language.Language{"fr"},
			language.LanguagePrimaryMatch,
		},
		{
			"hu-be",
			[]language.Language{"hu"},
			language.LanguagePrimaryMatch,
		},
		{
			"und-xxx",
			[]language.Language{"und-zsye"},
			language.LanguagesDiffer,
		},
	}
	for _, tt := range tests {
		ls := NewLangSet(tt.ls...)
		if got := ls.Contains(tt.l); got != tt.want {
			t.Errorf("LangSet.Contains(%s) = %v, want %v", tt.l, got, tt.want)
		}
	}
}
