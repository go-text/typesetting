package harfbuzz

import (
	"testing"

	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/unicodedata"
)

// ported from harfbuzz/test/api/test-unicode.c Copyright © 2011  Codethink Limited, Google, Inc. Ryan Lortie, Behdad Esfahbod

func TestUnicodeProp(t *testing.T) {
	runes := []rune{6176, 6155, 0x70f}
	exps := []unicodeProp{7, 236, 1}
	for i, r := range runes {
		got, _ := computeUnicodeProps(r)
		exp := exps[i]
		if got != exp {
			t.Fatalf("for rune 0x%x, expected %d, got %d", r, exp, got)
		}
	}
}

func TestGeneralCategory(t *testing.T) {
	if got := uni.generalCategory(0x70f); got != 1 {
		t.Errorf("for rune 0x%x, expected 1, got %d", 0x70f, got)
	}
}

/* Check all properties */

/* Some of the following tables where adapted from glib/glib/tests/utf8-misc.c.
 * The license is compatible. */

type testPairT struct {
	unicode rune
	value   uint
}

var combiningClassTests = []testPairT{
	{0x0020, 0},
	{0x0334, 1},
	{0x093C, 7},
	{0x3099, 8},
	{0x094D, 9},
	{0x05B0, 10},
	{0x05B1, 11},
	{0x05B2, 12},
	{0x05B3, 13},
	{0x05B4, 14},
	{0x05B5, 15},
	{0x05B6, 16},
	{0x05B7, 17},
	{0x05B8, 18},
	{0x05B9, 19},
	{0x05BB, 20},
	{0x05BC, 21},
	{0x05BD, 22},
	{0x05BF, 23},
	{0x05C1, 24},
	{0x05C2, 25},
	{0xFB1E, 26},
	{0x064B, 27},
	{0x064C, 28},
	{0x064D, 29},
	/* ... */
	{0x05AE, 228},
	{0x0300, 230},
	{0x302C, 232},
	{0x0362, 233},
	{0x0360, 234},
	{0x0345, 240},

	{0x111111, 0},
}

var combiningClassTestsMore = []testPairT{
	/* Unicode-5.1 character additions */
	{0x1DCD, 234},

	/* Unicode-5.2 character additions */
	{0xA8E0, 230},

	/* Unicode-6.0 character additions */
	{0x135D, 230},

	/* Unicode-6.1 character additions */
	{0xA674, 230},

	/* Unicode-7.0 character additions */
	{0x1AB0, 230},

	/* Unicode-8.0 character additions */
	{0xA69E, 230},

	/* Unicode-9.0 character additions */
	{0x1E000, 230},

	/* Unicode-10.0 character additions */
	{0x1DF6, 232},

	/* Unicode-11.0 character additions */
	{0x07FD, 220},

	/* Unicode-12.0 character additions */
	{0x0EBA, 9},

	/* Unicode-13.0 character additions */
	{0x1ABF, 220},

	/* Unicode-14.0 character additions */
	{0x1DFA, 218},

	/* Unicode-15.0 character additions */
	{0x10EFD, 220},

	{0x111111, 0},
}

var generalCategoryTests = []testPairT{
	{0x000D, uint(control)},
	{0x200E, uint(format)},
	{0x0378, uint(unassigned)},
	{0xE000, uint(privateUse)},
	{0xD800, uint(surrogate)},
	{0x0061, uint(lowercaseLetter)},
	{0x02B0, uint(modifierLetter)},
	{0x3400, uint(otherLetter)},
	{0x01C5, uint(titlecaseLetter)},
	{0xFF21, uint(uppercaseLetter)},
	{0x0903, uint(spacingMark)},
	{0x20DD, uint(enclosingMark)},
	{0xA806, uint(nonSpacingMark)},
	{0xFF10, uint(decimalNumber)},
	{0x16EE, uint(letterNumber)},
	{0x17F0, uint(otherNumber)},
	{0x005F, uint(connectPunctuation)},
	{0x058A, uint(dashPunctuation)},
	{0x0F3B, uint(closePunctuation)},
	{0x2019, uint(finalPunctuation)},
	{0x2018, uint(initialPunctuation)},
	{0x2016, uint(otherPunctuation)},
	{0x0F3A, uint(openPunctuation)},
	{0x20A0, uint(currencySymbol)},
	{0x309B, uint(modifierSymbol)},
	{0xFB29, uint(mathSymbol)},
	{0x00A6, uint(otherSymbol)},
	{0x2028, uint(lineSeparator)},
	{0x2029, uint(paragraphSeparator)},
	{0x202F, uint(spaceSeparator)},

	{0x111111, uint(unassigned)},
}

var generalCategoryTestsMore = []testPairT{
	/* Unicode-5.2 character additions */
	{0x1F131, uint(otherSymbol)},

	/* Unicode-6.0 character additions */
	{0x0620, uint(otherLetter)},

	/* Unicode-6.1 character additions */
	{0x058F, uint(currencySymbol)},

	/* Unicode-6.2 character additions */
	{0x20BA, uint(currencySymbol)},

	/* Unicode-6.3 character additions */
	{0x061C, uint(format)},

	/* Unicode-7.0 character additions */
	{0x058D, uint(otherSymbol)},

	/* Unicode-8.0 character additions */
	{0x08E3, uint(nonSpacingMark)},

	/* Unicode-9.0 character additions */
	{0x08D4, uint(nonSpacingMark)},

	/* Unicode-10.0 character additions */
	{0x09FD, uint(otherPunctuation)},

	/* Unicode-11.0 character additions */
	{0x0560, uint(lowercaseLetter)},

	/* Unicode-12.0 character additions */
	{0x0C77, uint(otherPunctuation)},

	/* Unicode-12.1 character additions */
	{0x32FF, uint(otherSymbol)},

	/* Unicode-13.0 character additions */
	{0x08BE, uint(otherLetter)},

	/* Unicode-14.0 character additions */
	{0x20C0, uint(currencySymbol)},

	/* Unicode-15.0 character additions */
	{0x0CF3, uint(spacingMark)},

	{0x111111, uint(unassigned)},
}

var mirroringTests = []testPairT{
	/* Some characters that do NOT mirror */
	{0x0020, 0x0020},
	{0x0041, 0x0041},
	{0x00F0, 0x00F0},
	{0x27CC, 0x27CC},
	{0xE01EF, 0xE01EF},
	{0x1D7C3, 0x1D7C3},
	{0x100000, 0x100000},

	/* Some characters that do mirror */
	{0x0029, 0x0028},
	{0x0028, 0x0029},
	{0x003E, 0x003C},
	{0x003C, 0x003E},
	{0x005D, 0x005B},
	{0x005B, 0x005D},
	{0x007D, 0x007B},
	{0x007B, 0x007D},
	{0x00BB, 0x00AB},
	{0x00AB, 0x00BB},
	{0x226B, 0x226A},
	{0x226A, 0x226B},
	{0x22F1, 0x22F0},
	{0x22F0, 0x22F1},
	{0xFF60, 0xFF5F},
	{0xFF5F, 0xFF60},
	{0xFF63, 0xFF62},
	{0xFF62, 0xFF63},

	{0x111111, 0x111111},
}

var mirroringTestsMore = []testPairT{
	/* Unicode-6.1 character additions */
	{0x27CB, 0x27CD},

	/* Unicode-11.0 character additions */
	{0x2BFE, 0x221F},

	{0x111111, 0x111111},
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

	{0x111111, uint(language.Unknown)},
}

type propertyTest struct {
	name             string
	getter           func(unicode rune) uint
	tests, testsMore []testPairT
}

var properties = [...]propertyTest{
	{"combiningClass", func(u rune) uint { return uint(unicodedata.LookupCombiningClass(u)) }, combiningClassTests, combiningClassTestsMore},
	{"generalCategory", func(u rune) uint { return uint(uni.generalCategory(u)) }, generalCategoryTests, generalCategoryTestsMore},
	{"mirroring", func(u rune) uint { return uint(uni.mirroring(u)) }, mirroringTests, mirroringTestsMore},
	{"script", func(u rune) uint { return uint(language.LookupScript(u)) }, scriptTests, scriptTestsMore},
}

func TestUnicodeProperties(t *testing.T) {
	for _, p := range properties {
		tests := append(p.tests, p.testsMore...)
		for _, te := range tests {
			if v := p.getter(te.unicode); v != te.value {
				t.Errorf("for property %s for rune 0x%x, expected 0x%x, got %d", p.name, te.unicode, te.value, v)
			}
		}
	}
}
