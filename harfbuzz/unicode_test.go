package harfbuzz

import (
	"testing"

	ucd "github.com/go-text/typesetting/internal/unicodedata"
	"github.com/go-text/typesetting/language"
)

// ported from harfbuzz/test/api/test-unicode.c Copyright © 2011  Codethink Limited, Google, Inc. Ryan Lortie, Behdad Esfahbod

func TestUnicodeProp(t *testing.T) {
	runes := []rune{6176, 6155, 0x70f}
	exps := []unicodeProp{7, 236, unicodeProp(ucd.Cf)}
	for i, r := range runes {
		got, _ := computeUnicodeProps(r)
		exp := exps[i]
		if got != exp {
			t.Fatalf("for rune 0x%x, expected %d, got %d", r, exp, got)
		}
	}
}

/* Check all properties */

/* Some of the following tables where adapted from glib/glib/tests/utf8-misc.c.
 * The license is compatible. */

type testPairT struct {
	unicode rune
	value   uint
}

var scriptTests = []testPairT{
	{0x002A, uint(language.Common)},
	{0x0670, uint(language.Inherited)},
	{0x060D, uint(language.Arabic)},
	{0x0559, uint(language.Armenian)},
	{0x09CD, uint(language.Bengali)},
	{0x31B6, uint(language.Bopomofo)},
	{0x13A2, uint(language.Cherokee)},
	{0x2CFD, uint(language.Coptic)},
	{0x0482, uint(language.Cyrillic)},
	{0x10401, uint(language.Deseret)},
	{0x094D, uint(language.Devanagari)},
	{0x1258, uint(language.Ethiopic)},
	{0x10FC, uint(language.Georgian)},
	{0x10341, uint(language.Gothic)},
	{0x0375, uint(language.Greek)},
	{0x0A83, uint(language.Gujarati)},
	{0x0A3C, uint(language.Gurmukhi)},
	{0x3005, uint(language.Han)},
	{0x1100, uint(language.Hangul)},
	{0x05BF, uint(language.Hebrew)},
	{0x309F, uint(language.Hiragana)},
	{0x0CBC, uint(language.Kannada)},
	{0x30FF, uint(language.Katakana)},
	{0x17DD, uint(language.Khmer)},
	{0x0EDD, uint(language.Lao)},
	{0x0061, uint(language.Latin)},
	{0x0D3D, uint(language.Malayalam)},
	{0x1843, uint(language.Mongolian)},
	{0x1031, uint(language.Myanmar)},
	{0x169C, uint(language.Ogham)},
	{0x10322, uint(language.Old_Italic)},
	{0x0B3C, uint(language.Oriya)},
	{0x16EF, uint(language.Runic)},
	{0x0DBD, uint(language.Sinhala)},
	{0x0711, uint(language.Syriac)},
	{0x0B82, uint(language.Tamil)},
	{0x0C03, uint(language.Telugu)},
	{0x07B1, uint(language.Thaana)},
	{0x0E31, uint(language.Thai)},
	{0x0FD4, uint(language.Tibetan)},
	// {0x1401, uint(language.Canadian_Syllabics)},
	{0xA015, uint(language.Yi)},
	{0x1700, uint(language.Tagalog)},
	{0x1720, uint(language.Hanunoo)},
	{0x1740, uint(language.Buhid)},
	{0x1760, uint(language.Tagbanwa)},

	/* Unicode-4.0 additions */
	{0x2800, uint(language.Braille)},
	{0x10808, uint(language.Cypriot)},
	{0x1932, uint(language.Limbu)},
	{0x10480, uint(language.Osmanya)},
	{0x10450, uint(language.Shavian)},
	{0x10000, uint(language.Linear_B)},
	{0x1950, uint(language.Tai_Le)},
	{0x1039F, uint(language.Ugaritic)},

	/* Unicode-4.1 additions */
	{0x1980, uint(language.New_Tai_Lue)},
	{0x1A1F, uint(language.Buginese)},
	{0x2C00, uint(language.Glagolitic)},
	{0x2D6F, uint(language.Tifinagh)},
	{0xA800, uint(language.Syloti_Nagri)},
	{0x103D0, uint(language.Old_Persian)},
	{0x10A3F, uint(language.Kharoshthi)},

	/* Unicode-5.0 additions */
	{0x0378, uint(language.Unknown)},
	{0x1B04, uint(language.Balinese)},
	{0x12000, uint(language.Cuneiform)},
	{0x10900, uint(language.Phoenician)},
	{0xA840, uint(language.Phags_Pa)},
	{0x07C0, uint(language.Nko)},

	/* Unicode-5.1 additions */
	{0xA900, uint(language.Kayah_Li)},
	{0x1C00, uint(language.Lepcha)},
	{0xA930, uint(language.Rejang)},
	{0x1B80, uint(language.Sundanese)},
	{0xA880, uint(language.Saurashtra)},
	{0xAA00, uint(language.Cham)},
	{0x1C50, uint(language.Ol_Chiki)},
	{0xA500, uint(language.Vai)},
	{0x102A0, uint(language.Carian)},
	{0x10280, uint(language.Lycian)},
	{0x1093F, uint(language.Lydian)},

	{0x111111, uint(language.Unknown)},
}

var scriptTestsMore = []testPairT{
	/* Unicode-5.2 additions */
	{0x10B00, uint(language.Avestan)},
	{0xA6A0, uint(language.Bamum)},
	{0x1400, uint(language.Canadian_Aboriginal)},
	{0x13000, uint(language.Egyptian_Hieroglyphs)},
	{0x10840, uint(language.Imperial_Aramaic)},
	{0x1CED, uint(language.Inherited)},
	{0x10B60, uint(language.Inscriptional_Pahlavi)},
	{0x10B40, uint(language.Inscriptional_Parthian)},
	{0xA980, uint(language.Javanese)},
	{0x11082, uint(language.Kaithi)},
	{0xA4D0, uint(language.Lisu)},
	{0xABE5, uint(language.Meetei_Mayek)},
	{0x10A60, uint(language.Old_South_Arabian)},
	{0x10C00, uint(language.Old_Turkic)},
	{0x0800, uint(language.Samaritan)},
	{0x1A20, uint(language.Tai_Tham)},
	{0xAA80, uint(language.Tai_Viet)},

	/* Unicode-6.0 additions */
	{0x1BC0, uint(language.Batak)},
	{0x11000, uint(language.Brahmi)},
	{0x0840, uint(language.Mandaic)},

	/* Unicode-6.1 additions */
	{0x10980, uint(language.Meroitic_Hieroglyphs)},
	{0x109A0, uint(language.Meroitic_Cursive)},
	{0x110D0, uint(language.Sora_Sompeng)},
	{0x11100, uint(language.Chakma)},
	{0x11180, uint(language.Sharada)},
	{0x11680, uint(language.Takri)},
	{0x16F00, uint(language.Miao)},

	/* Unicode-6.2 additions */
	{0x20BA, uint(language.Common)},

	/* Unicode-6.3 additions */
	{0x2066, uint(language.Common)},

	/* Unicode-7.0 additions */
	{0x10350, uint(language.Old_Permic)},
	{0x10500, uint(language.Elbasan)},
	{0x10530, uint(language.Caucasian_Albanian)},
	{0x10600, uint(language.Linear_A)},
	{0x10860, uint(language.Palmyrene)},
	{0x10880, uint(language.Nabataean)},
	{0x10A80, uint(language.Old_North_Arabian)},
	{0x10AC0, uint(language.Manichaean)},
	{0x10B80, uint(language.Psalter_Pahlavi)},
	{0x11150, uint(language.Mahajani)},
	{0x11200, uint(language.Khojki)},
	{0x112B0, uint(language.Khudawadi)},
	{0x11300, uint(language.Grantha)},
	{0x11480, uint(language.Tirhuta)},
	{0x11580, uint(language.Siddham)},
	{0x11600, uint(language.Modi)},
	{0x118A0, uint(language.Warang_Citi)},
	{0x11AC0, uint(language.Pau_Cin_Hau)},
	{0x16A40, uint(language.Mro)},
	{0x16AD0, uint(language.Bassa_Vah)},
	{0x16B00, uint(language.Pahawh_Hmong)},
	{0x1BC00, uint(language.Duployan)},
	{0x1E800, uint(language.Mende_Kikakui)},

	/* Unicode-8.0 additions */
	{0x108E0, uint(language.Hatran)},
	{0x10C80, uint(language.Old_Hungarian)},
	{0x11280, uint(language.Multani)},
	{0x11700, uint(language.Ahom)},
	{0x14400, uint(language.Anatolian_Hieroglyphs)},
	{0x1D800, uint(language.SignWriting)},

	/* Unicode-9.0 additions */
	{0x104B0, uint(language.Osage)},
	{0x11400, uint(language.Newa)},
	{0x11C00, uint(language.Bhaiksuki)},
	{0x11C70, uint(language.Marchen)},
	{0x17000, uint(language.Tangut)},
	{0x1E900, uint(language.Adlam)},

	/* Unicode-10.0 additions */
	{0x11A00, uint(language.Zanabazar_Square)},
	{0x11A50, uint(language.Soyombo)},
	{0x11D00, uint(language.Masaram_Gondi)},
	{0x1B170, uint(language.Nushu)},

	/* Unicode-11.0 additions */
	{0x10D00, uint(language.Hanifi_Rohingya)},
	{0x10F00, uint(language.Old_Sogdian)},
	{0x10F30, uint(language.Sogdian)},
	{0x11800, uint(language.Dogra)},
	{0x11D60, uint(language.Gunjala_Gondi)},
	{0x11EE0, uint(language.Makasar)},
	{0x16E40, uint(language.Medefaidrin)},

	/* Unicode-12.0 additions */
	{0x10FE0, uint(language.Elymaic)},
	{0x119A0, uint(language.Nandinagari)},
	{0x1E100, uint(language.Nyiakeng_Puachue_Hmong)},
	{0x1E2C0, uint(language.Wancho)},

	/* Unicode-12.1 additions */
	{0x32FF, uint(language.Common)},

	/* Unicode-13.0 additions */
	{0x10E80, uint(language.Yezidi)},
	{0x10FB0, uint(language.Chorasmian)},
	{0x11900, uint(language.Dives_Akuru)},
	{0x18B00, uint(language.Khitan_Small_Script)},

	/* Unicode-14.0 additions */
	{0x10570, uint(language.Vithkuqi)},
	{0x10F70, uint(language.Old_Uyghur)},
	{0x12F90, uint(language.Cypro_Minoan)},
	{0x16A70, uint(language.Tangsa)},
	{0x1E290, uint(language.Toto)},

	/* Unicode-15.0 additions */
	{0x11F00, uint(language.Kawi)},
	{0x1E4D0, uint(language.Nag_Mundari)},

	/* Unicode-16.0 additions */
	{0x105C0, uint(language.Todhri)},
	{0x10D40, uint(language.Garay)},
	{0x11380, uint(language.Tulu_Tigalari)},
	{0x11BC0, uint(language.Sunuwar)},
	{0x16100, uint(language.Gurung_Khema)},
	{0x16D40, uint(language.Kirat_Rai)},
	{0x1E5D0, uint(language.Ol_Onal)},

	/* Unicode-16.0 additions */
	{0x10940, uint(language.Sidetic)},
	{0x11DB0, uint(language.Tolong_Siki)},
	{0x16EA0, uint(language.Beria_Erfe)},
	{0x1E6C0, uint(language.Tai_Yo)},

	{0x111111, uint(language.Unknown)},
}

type propertyTest struct {
	name             string
	getter           func(unicode rune) uint
	tests, testsMore []testPairT
}

var properties = [...]propertyTest{
	{"script", func(u rune) uint { return uint(language.LookupScript(u)) }, scriptTests, scriptTestsMore},
}

func TestUnicodeProperties(t *testing.T) {
	for _, p := range properties {
		tests := append(p.tests, p.testsMore...)
		for _, te := range tests {
			if v := p.getter(te.unicode); v != te.value {
				t.Errorf("for property %s, for rune 0x%x, expected 0x%x, got 0x%x", p.name, te.unicode, te.value, v)
			}
		}
	}
}
