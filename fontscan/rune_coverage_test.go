package fontscan

import (
	"bytes"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

// newRuneSet builds a set containing the given runes.
func newRuneSet(runes ...rune) RuneSet {
	var rs RuneSet
	for _, r := range runes {
		rs.Add(r)
	}
	return rs
}

func randomRunes() []rune {
	out := make([]rune, 1000)
	const maxRuneScript = 0xe01ef
	for i := range out {
		out[i] = rand.Int31n(maxRuneScript + 10) // allow some invalid runes
	}
	return out
}

func randomRanges() [][2]rune {
	L := 50 + rand.Intn(300)
	out := make([][2]rune, L)

	if rand.Intn(2) == 0 {
		lastEnd := 0
		for i := range out {
			start := lastEnd + rand.Intn(2)
			end := start + 1 + rand.Intn(300)
			lastEnd = end
			out[i] = [2]rune{rune(start), rune(end)}
		}
	} else {
		lastIndex := 0
		for i := range out {
			index := lastIndex + rand.Intn(2)
			item := language.ScriptRanges[index]
			out[i] = [2]rune{item.Start, item.End}
			lastIndex = index
		}
	}
	return out
}

func runesFromRanges(ranges [][2]rune) []rune {
	var out []rune
	for _, ra := range ranges {
		for r := ra[0]; r <= ra[1]; r++ {
			out = append(out, r)
		}
	}
	return out
}

// runes returns a copy of the runes in the set.
func (rs RuneSet) runes() (out []rune) {
	for _, page := range rs {
		pageLow := rune(page.ref) << 8
		for j, set := range page.set {
			for k := rune(0); k < 32; k++ {
				if set&uint32(1<<k) != 0 {
					out = append(out, pageLow|rune(j)<<5|k)
				}
			}
		}
	}
	return out
}

func TestRuneSet(t *testing.T) {
	tests := []struct {
		start    []rune
		expected []rune
		r        rune
	}{
		{
			nil,
			[]rune{0},
			0,
		},
		{
			nil,
			[]rune{1},
			1,
		},
		{
			nil,
			[]rune{32},
			32,
		},
		{
			[]rune{0, 32, 257},
			[]rune{0, 32, 257, 512},
			512,
		},
		{
			[]rune{0, 32, 257, 1000, 2000, 1500},
			[]rune{0, 32, 257, 1000, 1500, 2000, 10000},
			10000,
		},
	}
	for _, tt := range tests {
		cov := newRuneSet(tt.start...)
		cov.Add(tt.r)
		if runes := cov.runes(); !reflect.DeepEqual(runes, tt.expected) {
			t.Fatalf("expected %v, got %v (%v)", tt.expected, runes, cov)
		}

		for _, r := range tt.expected {
			if !cov.Contains(r) {
				t.Fatalf("missing char %d", r)
			}
		}

		cov.Delete(tt.r)
		sort.Slice(tt.start, func(i, j int) bool { return tt.start[i] < tt.start[j] })
		if runes := cov.runes(); !reflect.DeepEqual(runes, tt.start) {
			t.Fatalf("expected %v, got %v (%v)", tt.start, runes, cov)
		}

		for _, r := range tt.start {
			if !cov.Contains(r) {
				t.Fatalf("missing char %d", r)
			}
		}
		if cov.Contains(tt.r) {
			t.Fatalf("rune %d should be deleted", tt.r)
		}
		if cov.Contains(1_000_000) {
			t.Fatalf("rune %d should be missing", 1_000_000)
		}

		cov.Delete(1_000_000) // no op

		if cov.Len() != len(tt.start) {
			t.Fatalf("unexpected length %d", cov.Len())
		}
	}
}

func TestBinaryFormat(t *testing.T) {
	for range [50]int{} {
		cov := newRuneSet(randomRunes()...)
		b := cov.serialize()

		var got RuneSet
		n, err := got.deserializeFrom(b)
		if err != nil {
			t.Fatalf("Coverage.deserializeFrom: %s", err)
		}

		if n != len(b) {
			t.Fatalf("unexpected number of bytes read: %d", n)
		}

		if !reflect.DeepEqual(cov, got) {
			t.Fatalf("expected %v, got %v", cov, got)
		}
	}
}

func TestDeserializeFrom(t *testing.T) {
	var cov RuneSet

	if _, err := cov.deserializeFrom(nil); err == nil {
		t.Fatal("exepcted error on invalid input")
	}
	if _, err := cov.deserializeFrom([]byte{0, 5}); err == nil {
		t.Fatal("exepcted error on invalid input")
	}
}

// CmapSimple is a map based Cmap implementation.
type CmapSimple map[rune]font.GID

type cmap0Iter struct {
	data CmapSimple
	keys []rune
	pos  int
}

func (it *cmap0Iter) Next() bool {
	return it.pos < len(it.keys)
}

func (it *cmap0Iter) Char() (rune, font.GID) {
	r := it.keys[it.pos]
	it.pos++
	return r, it.data[r]
}

func (s CmapSimple) Iter() font.CmapIter {
	keys := make([]rune, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return &cmap0Iter{data: s, keys: keys}
}

func (s CmapSimple) Lookup(r rune) (font.GID, bool) {
	v, ok := s[r] // will be 0 if r is not in s
	return v, ok
}

func TestNewRuneSetFromCmap(t *testing.T) {
	tests := []struct {
		args font.Cmap
		want RuneSet
	}{
		{CmapSimple{0: 0, 1: 0, 2: 0, 0xfff: 0}, newRuneSet(0, 1, 2, 0xfff)},
		{CmapSimple{0: 0, 1: 0, 2: 0, 800: 0, 801: 0, 1000: 0}, newRuneSet(0, 1, 2, 800, 801, 1000)},
	}
	for _, tt := range tests {
		if got, _, _ := newCoveragesFromCmap(tt.args, nil); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("NewRuneSetFromCmap() = %v, want %v", got, tt.want)
		}
	}
}

func TestBits(t *testing.T) {
	a, b := 2, 13
	var total uint32
	for i := a; i <= b; i++ {
		total |= 1 << i
	}

	alt := (uint32(1)<<(b-a+1) - 1) << a // mask for bits from a to b (included)
	tu.Assert(t, total == alt)
}

type runeRange [][2]rune

func (rr runeRange) RuneRanges(_ [][2]rune) [][2]rune { return rr }

func (rr runeRange) runes() (out []rune) {
	for _, ra := range rr {
		for r := ra[0]; r <= ra[1]; r++ {
			out = append(out, r)
		}
	}
	return out
}

func TestRuneRanges(t *testing.T) {
	for _, source := range []runeRange{
		{
			{0, 10}, {12, 14}, {15, 15}, {18, 2000}, {2100, 0xFFFFFF},
		},
		{
			{0, 30}, {0xFF, 0xFF * 2},
		},
	} {
		got, _, _ := newCoveragesFromCmapRange(source, nil)
		exp := newRuneSet(source.runes()...)
		tu.Assert(t, reflect.DeepEqual(got, exp))
	}
}

func TestScriptSet(t *testing.T) {
	type testcase struct {
		name     string
		toInsert []language.Script
		expected []language.Script
	}
	for _, tc := range []testcase{
		{
			name:     "empty add nothing",
			expected: []language.Script{},
		},
		{
			name:     "add latin",
			toInsert: []language.Script{language.Latin},
			expected: []language.Script{language.Latin},
		},
		{
			name:     "add latin twice",
			toInsert: []language.Script{language.Latin, language.Latin},
			expected: []language.Script{language.Latin},
		},
		{
			name:     "add latin and arabic",
			toInsert: []language.Script{language.Latin, language.Arabic},
			expected: []language.Script{language.Latin, language.Arabic},
		},
		{
			name:     "add latin and arabic twice",
			toInsert: []language.Script{language.Latin, language.Arabic, language.Latin, language.Arabic},
			expected: []language.Script{language.Latin, language.Arabic},
		},
		{
			name:     "add latin and arabic twice in a row",
			toInsert: []language.Script{language.Latin, language.Latin, language.Arabic, language.Arabic},
			expected: []language.Script{language.Latin, language.Arabic},
		},
		{
			name: "add many scripts",
			toInsert: func() []language.Script {
				scripts := make([]language.Script, 0, len(testScripts)*10)
				for i := 0; i < 10; i++ {
					scripts = append(scripts, testScripts...)
				}
				rand.Seed(0)
				rand.Shuffle(len(scripts), func(i, j int) {
					scripts[i], scripts[j] = scripts[j], scripts[i]
				})
				return scripts
			}(),
			expected: testScripts,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			set := make(ScriptSet, 0, 1)
			for _, s := range tc.toInsert {
				set.insert(s)
			}
			assertSetsMatch(t, tc.expected, set)
		})
	}
}

func assertSetsMatch(t *testing.T, expected, actual []language.Script) {
	t.Helper()
	sort.Slice(expected, func(i, j int) bool {
		return expected[i] < expected[j]
	})
	if len(expected) != len(actual) {
		t.Errorf("expected %d scripts, got %d", len(expected), len(actual))
		t.Logf("expected: %q", expected)
		t.Logf("actual: %q", actual)
	}
	for i := 0; i < min(len(expected), len(actual)); i++ {
		if expected[i] != actual[i] {
			t.Errorf("mismatch at index %d, expected %s got %s", i, expected[i], actual[i])
		}
	}
}

func TestRuneSetScripts(t *testing.T) {
	type testcase struct {
		name     string
		fontdata []byte
		expected []language.Script
	}
	for _, tc := range []testcase{
		{
			name: "Roboto Regular",
			fontdata: func() []byte {
				data, err := os.ReadFile("../font/testdata/Roboto-Regular.ttf")
				tu.AssertNoErr(t, err)
				return data
			}(),
			expected: []language.Script{language.Cyrillic, language.Greek, language.Latin, language.Inherited, language.Common, language.Unknown},
		},
		{
			name: "Amiri Regular",
			fontdata: func() []byte {
				data, err := os.ReadFile("../font/testdata/Amiri-Regular.ttf")
				tu.AssertNoErr(t, err)
				return data
			}(),
			expected: []language.Script{language.Arabic, language.Latin, language.Inherited, language.Common},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			face, err := font.ParseTTF(bytes.NewReader(tc.fontdata))
			tu.AssertNoErr(t, err)
			rs, actualScripts, _ := newCoveragesFromCmap(face.Cmap, nil)
			approxScripts := rs.approximateScripts()
			assertSetsMatch(t, tc.expected, actualScripts)
			assertSetsMatch(t, tc.expected, approxScripts)
		})
	}
}

// slow but easy implementation, used as a reference
func (rs RuneSet) scriptsNaive() ScriptSet {
	tmp := make(map[language.Script]bool)
	for _, r := range rs.runes() {
		s := language.LookupScript(r)
		tmp[s] = true
	}
	out := make(ScriptSet, 0, len(tmp))
	for s := range tmp {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func TestScriptsFromRanges(t *testing.T) {
	var ranges [][][2]rune
	for range [400]int{} {
		ranges = append(ranges, randomRanges())
	}
	ranges = append(ranges, [][2]rune{
		{0, 20},
		{0xe01ef - 2, 0xe01ef + 2},
	}, [][2]rune{
		{0, 30},
		{0xe01ef + 3, 0xe01ef + 5},
	})

	for _, rans := range ranges {
		runes := runesFromRanges(rans)
		rs := newRuneSet(runes...)
		exp, got := rs.scriptsNaive(), scriptsFromRanges(rans)
		if !reflect.DeepEqual(exp, got) {
			t.Fatalf("for %v, expected %v, got %v", rans, exp, got)
		}
	}
}

// approximateScripts returns an approximation of the scripts that the runeSet has coverage for.
// It works by sampling the coverage set for the first covered rune in every page and
// mapping that to a supported script. This means that it can miss some supported
// scripts.
func (rs RuneSet) approximateScripts() []language.Script {
	scripts := make(ScriptSet, 0, 1)
	for _, pageSet := range rs {
	pageSearch:
		for pageIdx, page := range pageSet.set {
			if page == 0 {
				continue
			}
			for i := 0; i < 32; i++ {
				if (page & (1 << i)) == 0 {
					continue
				}
				firstRune := (rune(pageSet.ref) << 8) | (rune(pageIdx) << 5) | rune(i)

				scripts.insert(language.LookupScript(firstRune))
				continue pageSearch
			}
		}
	}
	return scripts
}

func BenchmarkScriptSet(b *testing.B) {
	type test struct {
		ranges [][2]rune
		set    RuneSet
	}
	var cases []test
	for range [200]int{} {
		ranges := randomRanges()
		cases = append(cases, test{ranges, newRuneSet(runesFromRanges(ranges)...)})
	}

	b.Run("naive", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, test := range cases {
				_ = test.set.scriptsNaive()
			}
		}
	})
	b.Run("from rune pages", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, test := range cases {
				_ = test.set.approximateScripts()
			}
		}
	})
	b.Run("from rune ranges", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, test := range cases {
				_ = scriptsFromRanges(test.ranges)
			}
		}
	})
}

func (a RuneSet) includesNaive(b RuneSet) bool {
	ar := a.runes()
	br := b.runes()
	aSet := make(map[rune]bool)
	for _, r := range ar {
		aSet[r] = true
	}
	for _, b := range br {
		if !aSet[b] {
			return false
		}
	}
	return true
}

func Test_isIncludedIn(t *testing.T) {
	tu.Assert(t, newRuneSet(0, 1, 2, 3, 4).includes(newRuneSet(0, 1, 2, 3)))
	tu.Assert(t, newRuneSet(0, 1, 2, 3).includes(newRuneSet(0, 1, 2, 3)))
	tu.Assert(t, newRuneSet(0, 1, 2, 3, 12, 24).includes(newRuneSet(0, 1, 12)))
	tu.Assert(t, newRuneSet(0, 1, 2, 3, 12000, 13000).includes(newRuneSet(0, 1, 12000)))
	tu.Assert(t, !newRuneSet(0, 1, 2).includes(newRuneSet(0, 1, 12000)))
	tu.Assert(t, !newRuneSet(4, 5, 6, 8).includes(newRuneSet(4, 5, 6, 7)))

	// compare against an easy to write implementation
	for range [300]byte{} {
		a := newRuneSet(randomRunes()...)
		b := newRuneSet(randomRunes()...)
		tu.Assert(t, a.includes(b) == a.includesNaive(b))
	}
}

func BenchmarkIsIncludedIn(b *testing.B) {
	as := newRuneSet(randomRunes()...)
	bs := newRuneSet(append(randomRunes(), randomRunes()...)...)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = as.includes(bs)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var testScripts = []language.Script{
	language.Adlam,
	language.Afaka,
	language.Ahom,
	language.Anatolian_Hieroglyphs,
	language.Arabic,
	language.Armenian,
	language.Avestan,
	language.Balinese,
	language.Bamum,
	language.Bassa_Vah,
	language.Batak,
	language.Bengali,
	language.Bhaiksuki,
	language.Blissymbols,
	language.Book_Pahlavi,
	language.Bopomofo,
	language.Brahmi,
	language.Braille,
	language.Buginese,
	language.Buhid,
	language.Canadian_Aboriginal,
	language.Carian,
	language.Caucasian_Albanian,
	language.Chakma,
	language.Cham,
	language.Cherokee,
	language.Chorasmian,
	language.Cirth,
	language.Code_for_unwritten_documents,
	language.Common,
	language.Coptic,
	language.Cuneiform,
	language.Cypriot,
	language.Cypro_Minoan,
	language.Cyrillic,
	language.Deseret,
	language.Devanagari,
	language.Dives_Akuru,
	language.Dogra,
	language.Duployan,
	language.Egyptian_Hieroglyphs,
	language.Egyptian_demotic,
	language.Egyptian_hieratic,
	language.Elbasan,
	language.Elymaic,
	language.Ethiopic,
	language.Georgian,
	language.Glagolitic,
	language.Gothic,
	language.Grantha,
	language.Greek,
	language.Gujarati,
	language.Gunjala_Gondi,
	language.Gurmukhi,
	language.Han,
	language.Hangul,
	language.Hanifi_Rohingya,
	language.Hanunoo,
	language.Hatran,
	language.Hebrew,
	language.Hiragana,
	language.Imperial_Aramaic,
	language.Inherited,
	language.Inscriptional_Pahlavi,
	language.Inscriptional_Parthian,
	language.Javanese,
	language.Jurchen,
	language.Kaithi,
	language.Kannada,
	language.Katakana,
	language.Katakana_Or_Hiragana,
	language.Kawi,
	language.Kayah_Li,
	language.Kharoshthi,
	language.Khitan_Small_Script,
	language.Khitan_large_script,
	language.Khmer,
	language.Khojki,
	language.Khudawadi,
	language.Kpelle,
	language.Lao,
	language.Latin,
	language.Leke,
	language.Lepcha,
	language.Limbu,
	language.Linear_A,
	language.Linear_B,
	language.Lisu,
	language.Loma,
	language.Lycian,
	language.Lydian,
	language.Mahajani,
	language.Makasar,
	language.Malayalam,
	language.Mandaic,
	language.Manichaean,
	language.Marchen,
	language.Masaram_Gondi,
	language.Mathematical_notation,
	language.Mayan_hieroglyphs,
	language.Medefaidrin,
	language.Meetei_Mayek,
	language.Mende_Kikakui,
	language.Meroitic_Cursive,
	language.Meroitic_Hieroglyphs,
	language.Miao,
	language.Modi,
	language.Mongolian,
	language.Mro,
	language.Multani,
	language.Myanmar,
	language.Nabataean,
	language.Nag_Mundari,
	language.Nandinagari,
	language.New_Tai_Lue,
	language.Newa,
	language.Nko,
	language.Nushu,
	language.Nyiakeng_Puachue_Hmong,
	language.Ogham,
	language.Ol_Chiki,
	language.Old_Hungarian,
	language.Old_Italic,
	language.Old_North_Arabian,
	language.Old_Permic,
	language.Old_Persian,
	language.Old_Sogdian,
	language.Old_South_Arabian,
	language.Old_Turkic,
	language.Old_Uyghur,
	language.Oriya,
	language.Osage,
	language.Osmanya,
	language.Pahawh_Hmong,
	language.Palmyrene,
	language.Pau_Cin_Hau,
	language.Phags_Pa,
	language.Phoenician,
	language.Psalter_Pahlavi,
	language.Ranjana,
	language.Rejang,
	language.Rongorongo,
	language.Runic,
	language.Samaritan,
	language.Sarati,
	language.Saurashtra,
	language.Sharada,
	language.Shavian,
	language.Shuishu,
	language.Siddham,
	language.SignWriting,
	language.Sinhala,
	language.Sogdian,
	language.Sora_Sompeng,
	language.Soyombo,
	language.Sundanese,
	language.Sunuwar,
	language.Syloti_Nagri,
	language.Symbols,
	language.Syriac,
	language.Tagalog,
	language.Tagbanwa,
	language.Tai_Le,
	language.Tai_Tham,
	language.Tai_Viet,
	language.Takri,
	language.Tamil,
	language.Tangsa,
	language.Tangut,
	language.Telugu,
	language.Tengwar,
	language.Thaana,
	language.Thai,
	language.Tibetan,
	language.Tifinagh,
	language.Tirhuta,
	language.Toto,
	language.Ugaritic,
	language.Unknown,
	language.Vai,
	language.Visible_Speech,
	language.Vithkuqi,
	language.Wancho,
	language.Warang_Citi,
	language.Woleai,
	language.Yezidi,
	language.Yi,
	language.Zanabazar_Square,
}
