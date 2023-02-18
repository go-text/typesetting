package language

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLanguage(t *testing.T) {
	fmt.Println(DefaultLanguage())
}

func TestSimpleInheritance(t *testing.T) {
	l := NewLanguage("en_US_someVariant")
	if sh := l.SimpleInheritance(); !reflect.DeepEqual(sh, []Language{l, "en-us", "en"}) {
		t.Fatalf("unexpected inheritance %v", sh)
	}

	l = NewLanguage("fr")
	if sh := l.SimpleInheritance(); !reflect.DeepEqual(sh, []Language{l}) {
		t.Fatalf("unexpected inheritance %v", sh)
	}
}

func TestLanguage_IsDerivedFrom(t *testing.T) {
	type args struct {
		root Language
	}
	tests := []struct {
		name string
		l    Language
		args args
		want bool
	}{
		{
			name: "",
			l:    "fr-FR",
			args: args{"fr"},
			want: true,
		},
		{
			name: "",
			l:    "ca",
			args: args{"cat"},
			want: false,
		},
		{
			name: "",
			l:    "ca",
			args: args{"ca"},
			want: true,
		},
	}
	for _, tt := range tests {
		if got := tt.l.IsDerivedFrom(tt.args.root); got != tt.want {
			t.Errorf("Language.IsDerivedFrom() = %v, want %v", got, tt.want)
		}
	}
}

func TestLanguage_IsUndefined(t *testing.T) {
	tests := []struct {
		l    Language
		want bool
	}{
		{NewLanguage("und"), true},
		{NewLanguage("uNd"), true},
		{NewLanguage("und-"), true},
		{NewLanguage("und-07"), true},
		{NewLanguage("fr"), false},
		{NewLanguage("und4"), false},
		{NewLanguage("un"), false},
		{NewLanguage("und-zmth"), true}, // maths
		{NewLanguage("und-zsye"), true}, // emojis
	}
	for _, tt := range tests {
		if got := tt.l.IsUndetermined(); got != tt.want {
			t.Errorf("Language.IsUndefined() = %v, want %v", got, tt.want)
		}
	}
}

func TestLanguage_Compare(t *testing.T) {
	tests := []struct {
		l     Language
		other Language
		want  LanguageComparison
	}{
		{"fr", "fr", LanguagesExactMatch},
		{"fr-be", "fr-be", LanguagesExactMatch},
		{"und-fr", "und-fr", LanguagesExactMatch},
		{"es-fr", "es-be", LanguagePrimaryMatch},
		{"es-fr-78", "es-be", LanguagePrimaryMatch},
		{"und-fr", "und-be", LanguagesDiffer},
		{"und-math", "fr-math", LanguagesDiffer},
		{"fr-math", "und-math", LanguagesDiffer},
	}
	for _, tt := range tests {
		if got := tt.l.Compare(tt.other); got != tt.want {
			t.Errorf("Language.Compare() = %v, want %v", got, tt.want)
		}
	}
}
