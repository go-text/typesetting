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
	if sh := l.SimpleInheritance(); !reflect.DeepEqual(sh, []Language{l, {"en-us"}, {"en"}}) {
		t.Fatalf("unexpected inheritance %v", sh)
	}

	l = NewLanguage("fr")
	if sh := l.SimpleInheritance(); !reflect.DeepEqual(sh, []Language{l}) {
		t.Fatalf("unexpected inheritance %v", sh)
	}
}

func TestLanguage_IsDerivedFrom(t *testing.T) {
	tests := []struct {
		name string
		l    Language
		root Language
		want bool
	}{
		{
			name: "",
			l:    Language{"fr-FR"},
			root: Language{"fr"},
			want: true,
		},
		{
			name: "",
			l:    Language{"ca"},
			root: Language{"cat"},
			want: false,
		},
		{
			name: "",
			l:    Language{"ca"},
			root: Language{"ca"},
			want: true,
		},
	}
	for _, tt := range tests {
		if got := tt.l.IsDerivedFrom(tt.root); got != tt.want {
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
		{Language{"fr"}, Language{"fr"}, LanguagesExactMatch},
		{Language{"fr-be"}, Language{"fr-be"}, LanguagesExactMatch},
		{Language{"und-fr"}, Language{"und-fr"}, LanguagesExactMatch},
		{Language{"es-fr"}, Language{"es-be"}, LanguagePrimaryMatch},
		{Language{"es-fr-78"}, Language{"es-be"}, LanguagePrimaryMatch},
		{Language{"und-fr"}, Language{"und-be"}, LanguagesDiffer},
		{Language{"und-math"}, Language{"fr-math"}, LanguagesDiffer},
		{Language{"fr-math"}, Language{"und-math"}, LanguagesDiffer},
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
	{Language{"und"}, Language{"und"}, Language{""}},
	{Language{"und-syrn"}, Language{"und-syrn"}, Language{""}},
	{Language{"rm-ch-fonipa-sursilv-x-foobar"}, Language{"rm-ch-fonipa-sursilv"}, Language{"x-foobar"}},
	{Language{"fa-x-hbotabc-hbot-41686121-zxc"}, Language{"fa"}, Language{"x-hbotabc-hbot-41686121-zxc"}},
	{Language{"zh-x-hbotabc-hbot-41686121-zxc"}, Language{"zh"}, Language{"x-hbotabc-hbot-41686121-zxc"}},
	{Language{"fa-x-hbot-41686121-hbotabc-zxc"}, Language{"fa"}, Language{"x-hbot-41686121-hbotabc-zxc"}},
	{Language{"zh-x-hbot-41686121-hbotabc-zxc"}, Language{"zh"}, Language{"x-hbot-41686121-hbotabc-zxc"}},
	{Language{"fa-ir-x-hbotabc-hbot-41686121-zxc"}, Language{"fa-ir"}, Language{"x-hbotabc-hbot-41686121-zxc"}},
	{Language{"zh-cn-x-hbotabc-hbot-41686121-zxc"}, Language{"zh-cn"}, Language{"x-hbotabc-hbot-41686121-zxc"}},
	{Language{"zh-xy-x-hbotabc-hbot-41686121-zxc"}, Language{"zh-xy"}, Language{"x-hbotabc-hbot-41686121-zxc"}},
	{Language{"fa-ir-x-hbot-41686121-hbotabc-zxc"}, Language{"fa-ir"}, Language{"x-hbot-41686121-hbotabc-zxc"}},
	{Language{"zh-cn-x-hbot-41686121-hbotabc-zxc"}, Language{"zh-cn"}, Language{"x-hbot-41686121-hbotabc-zxc"}},
	{Language{"zh-xy-x-hbot-41686121-hbotabc-zxc"}, Language{"zh-xy"}, Language{"x-hbot-41686121-hbotabc-zxc"}},
	{Language{"xyz-xy-x-hbotabc-hbot-41686121-zxc"}, Language{"xyz-xy"}, Language{"x-hbotabc-hbot-41686121-zxc"}},
	{Language{"xyz-xy-x-hbot-41686121-hbotabc-zxc"}, Language{"xyz-xy"}, Language{"x-hbot-41686121-hbotabc-zxc"}},
	{Language{"x-hbscabc"}, Language{""}, Language{"x-hbscabc"}},
	{Language{"x-hbscdeva"}, Language{""}, Language{"x-hbscdeva"}},
	{Language{"x-hbscdev2"}, Language{""}, Language{"x-hbscdev2"}},
	{Language{"x-hbscdev3"}, Language{""}, Language{"x-hbscdev3"}},
	{Language{"x-hbsc-64657633"}, Language{""}, Language{"x-hbsc-64657633"}},
	{Language{"x-hbotpap0-hbsccopt"}, Language{""}, Language{"x-hbotpap0-hbsccopt"}},
	{Language{"en-x-hbsc"}, Language{"en"}, Language{"x-hbsc"}},
	{Language{"en-x-hbsc"}, Language{"en"}, Language{"x-hbsc"}},
	{Language{"en-x-hbscabc"}, Language{"en"}, Language{"x-hbscabc"}},
	{Language{"en-x-hbscdeva"}, Language{"en"}, Language{"x-hbscdeva"}},
	{Language{"en-x-hbscdev2"}, Language{"en"}, Language{"x-hbscdev2"}},
	{Language{"en-x-hbscdev3"}, Language{"en"}, Language{"x-hbscdev3"}},
	{Language{"en-x-fonipa"}, Language{"en"}, Language{"x-fonipa"}},
	// extension tags
	{Language{"en-a-fonipa"}, Language{"en"}, Language{""}},
	{Language{"en-a-qwe-b-fonipa"}, Language{"en"}, Language{""}},
	{Language{"en-a-qwe-x-fonipa"}, Language{"en"}, Language{"x-fonipa"}},
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
