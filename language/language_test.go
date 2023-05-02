package language

import (
	"fmt"
	"reflect"
	"testing"
)

func TestLanguage(t *testing.T) {
	fmt.Println(DefaultLanguage())
}

func TestNonASCIILanguage(t *testing.T) {
	_ = NewLanguage("Δ") // should not panic
	if l1, l2 := NewLanguage("aΔ"), NewLanguage("a"); l1 != l2 {
		t.Fatalf("unexpected handling of non ASCII tags: %s != %s", l1, l2)
	}
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

var extensionTagTags = []struct {
	l           Language
	wantPrefix  Language
	wantPrivate Language
}{
	{"und", "und", ""},
	{"und-syrn", "und-syrn", ""},
	{"rm-ch-fonipa-sursilv-x-foobar", "rm-ch-fonipa-sursilv", "x-foobar"},
	{"fa-x-hbotabc-hbot-41686121-zxc", "fa", "x-hbotabc-hbot-41686121-zxc"},
	{"zh-x-hbotabc-hbot-41686121-zxc", "zh", "x-hbotabc-hbot-41686121-zxc"},
	{"fa-x-hbot-41686121-hbotabc-zxc", "fa", "x-hbot-41686121-hbotabc-zxc"},
	{"zh-x-hbot-41686121-hbotabc-zxc", "zh", "x-hbot-41686121-hbotabc-zxc"},
	{"fa-ir-x-hbotabc-hbot-41686121-zxc", "fa-ir", "x-hbotabc-hbot-41686121-zxc"},
	{"zh-cn-x-hbotabc-hbot-41686121-zxc", "zh-cn", "x-hbotabc-hbot-41686121-zxc"},
	{"zh-xy-x-hbotabc-hbot-41686121-zxc", "zh-xy", "x-hbotabc-hbot-41686121-zxc"},
	{"fa-ir-x-hbot-41686121-hbotabc-zxc", "fa-ir", "x-hbot-41686121-hbotabc-zxc"},
	{"zh-cn-x-hbot-41686121-hbotabc-zxc", "zh-cn", "x-hbot-41686121-hbotabc-zxc"},
	{"zh-xy-x-hbot-41686121-hbotabc-zxc", "zh-xy", "x-hbot-41686121-hbotabc-zxc"},
	{"xyz-xy-x-hbotabc-hbot-41686121-zxc", "xyz-xy", "x-hbotabc-hbot-41686121-zxc"},
	{"xyz-xy-x-hbot-41686121-hbotabc-zxc", "xyz-xy", "x-hbot-41686121-hbotabc-zxc"},
	{"x-hbscabc", "", "x-hbscabc"},
	{"x-hbscdeva", "", "x-hbscdeva"},
	{"x-hbscdev2", "", "x-hbscdev2"},
	{"x-hbscdev3", "", "x-hbscdev3"},
	{"x-hbsc-64657633", "", "x-hbsc-64657633"},
	{"x-hbotpap0-hbsccopt", "", "x-hbotpap0-hbsccopt"},
	{"en-x-hbsc", "en", "x-hbsc"},
	{"en-x-hbsc", "en", "x-hbsc"},
	{"en-x-hbscabc", "en", "x-hbscabc"},
	{"en-x-hbscdeva", "en", "x-hbscdeva"},
	{"en-x-hbscdev2", "en", "x-hbscdev2"},
	{"en-x-hbscdev3", "en", "x-hbscdev3"},
	{"en-x-fonipa", "en", "x-fonipa"},
	// extension tags
	{"en-a-fonipa", "en", ""},
	{"en-a-qwe-b-fonipa", "en", ""},
	{"en-a-qwe-x-fonipa", "en", "x-fonipa"},
}

func TestLanguage_SplitExtensionTags(t *testing.T) {
	for _, tt := range extensionTagTags {
		gotPrefix, gotPrivate := tt.l.SplitExtensionTags()
		if gotPrefix != tt.wantPrefix {
			t.Errorf("Language.SplitExtensionTags() gotPrefix = %v, want %v", gotPrefix, tt.wantPrefix)
		}
		if gotPrivate != tt.wantPrivate {
			t.Errorf("Language.SplitExtensionTags() gotPrivate = %v, want %v", gotPrivate, tt.wantPrivate)
		}
	}
}

func Benchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range extensionTagTags {
			_, _ = test.l.SplitExtensionTags()
		}
	}
}
