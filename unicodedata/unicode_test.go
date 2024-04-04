package unicodedata

import (
	"reflect"
	"testing"
	"unicode"

	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

func TestUnicodeNormalization(t *testing.T) {
	assertCompose := func(a, b rune, okExp bool, abExp rune) {
		ab, ok := Compose(a, b)
		if ok != okExp || ab != abExp {
			t.Errorf("expected %d, %v got %d, %v", abExp, okExp, ab, ok)
		}
	}

	// Not composable
	assertCompose(0x0041, 0x0042, false, 0)
	assertCompose(0x0041, 0, false, 0)
	assertCompose(0x0066, 0x0069, false, 0)

	// Singletons should not compose
	assertCompose(0x212B, 0, false, 0)
	assertCompose(0x00C5, 0, false, 0)
	assertCompose(0x2126, 0, false, 0)
	assertCompose(0x03A9, 0, false, 0)

	// Non-starter pairs should not compose
	assertCompose(0x0308, 0x0301, false, 0) // !0x0344
	assertCompose(0x0F71, 0x0F72, false, 0) // !0x0F73

	// Pairs
	assertCompose(0x0041, 0x030A, true, 0x00C5)
	assertCompose(0x006F, 0x0302, true, 0x00F4)
	assertCompose(0x1E63, 0x0307, true, 0x1E69)
	assertCompose(0x0073, 0x0323, true, 0x1E63)
	assertCompose(0x0064, 0x0307, true, 0x1E0B)
	assertCompose(0x0064, 0x0323, true, 0x1E0D)

	// Hangul
	assertCompose(0xD4CC, 0x11B6, true, 0xD4DB)
	assertCompose(0x1111, 0x1171, true, 0xD4CC)
	assertCompose(0xCE20, 0x11B8, true, 0xCE31)
	assertCompose(0x110E, 0x1173, true, 0xCE20)

	assertCompose(0xAC00, 0x11A7, false, 0)
	assertCompose(0xAC00, 0x11A8, true, 0xAC01)
	assertCompose(0xAC01, 0x11A8, false, 0)

	assertDecompose := func(ab rune, expOk bool, expA, expB rune) {
		a, b, ok := Decompose(ab)
		if ok != expOk || a != expA || b != expB {
			t.Errorf("decompose: expected 0x%x, 0x%x, %v got 0x%x, 0x%x, %v", expA, expB, expOk, a, b, ok)
		}
	}

	// Not decomposable
	assertDecompose(0x0041, false, 0x0041, 0)
	assertDecompose(0xFB01, false, 0xFB01, 0)
	assertDecompose(0x1F1EF, false, 0x1F1EF, 0)

	// Singletons
	assertDecompose(0x212B, true, 0x00C5, 0)
	assertDecompose(0x2126, true, 0x03A9, 0)

	// Non-starter pairs decompose, but not compose
	assertDecompose(0x0344, true, 0x0308, 0x0301)
	assertDecompose(0x0F73, true, 0x0F71, 0x0F72)

	// Pairs
	assertDecompose(0x00C5, true, 0x0041, 0x030A)
	assertDecompose(0x00F4, true, 0x006F, 0x0302)
	assertDecompose(0x1E69, true, 0x1E63, 0x0307)
	assertDecompose(0x1E63, true, 0x0073, 0x0323)
	assertDecompose(0x1E0B, true, 0x0064, 0x0307)
	assertDecompose(0x1E0D, true, 0x0064, 0x0323)

	// Hangul
	assertDecompose(0xD4DB, true, 0xD4CC, 0x11B6)
	assertDecompose(0xD4CC, true, 0x1111, 0x1171)
	assertDecompose(0xCE31, true, 0xCE20, 0x11B8)
	assertDecompose(0xCE20, true, 0x110E, 0x1173)
}

func TestBreakClass(t *testing.T) {
	if LookupLineBreakClass('\u2024') != BreakIN {
		t.Fatal("invalid break class for 0x2024")
	}
}

func TestLookupType(t *testing.T) {
	// some manual test cases
	tests := []struct {
		args rune
		want *unicode.RangeTable
	}{
		{'a', unicode.Ll},
		{'.', unicode.Po},
		{'カ', unicode.Lo},
		{'🦳', unicode.So},
		{-1, nil},
	}
	for _, tt := range tests {
		if got := LookupType(tt.args); got != tt.want {
			t.Errorf("LookupType(%s) = %v, want %v", string(tt.args), got, tt.want)
		}
	}
}

func TestLookupCombiningClass(t *testing.T) {
	// reference values are from https://www.compart.com/en/unicode/combining/
	tests := []struct {
		args rune
		want uint8
	}{
		{-1, 0},
		{'a', 0},
		{'\u093F', 0},
		{'\u0E40', 0},
		{'\u093c', 7},
		{'\u1039', 9},
		{'\u0f7b', 130},
		{'\u1cdd', 220},
		{'\u0369', 230},
	}
	for _, tt := range tests {
		if got := LookupCombiningClass(tt.args); got != tt.want {
			t.Errorf("LookupCombiningClass(%s) = %v, want %v", string(tt.args), got, tt.want)
		}
	}
}

func TestLookupLineBreakClass(t *testing.T) {
	// See https://www.unicode.org/reports/tr14/#DescriptionOfProperties
	tests := []struct {
		args rune
		want *unicode.RangeTable
	}{
		{-1, BreakXX},

		// AI: Ambiguous (Alphabetic or Ideograph)
		{'\u24EA', BreakAI},
		{'\u2780', BreakAI},
		// AL: Ordinary Alphabetic and Symbol Characters (XP)
		{'\u0600', BreakNU},     //   ARABIC NUMBER SIGN
		{'\u06DD', BreakNU},     //  	ARABIC END OF AYAH
		{'\u070F', BreakAL},     //  	SYRIAC ABBREVIATION MARK
		{'\u2061', BreakAL},     //   	FUNCTION APPLICATION
		{'\U000110BD', BreakNU}, //  	KAITHI NUMBER SIGN
		// BA: Break After (A)
		{'\u1680', BreakBA},     // OGHAM SPACE MARK
		{'\u2000', BreakBA},     // EN QUAD
		{'\u2001', BreakBA},     // EM QUAD
		{'\u2002', BreakBA},     // EN SPACE
		{'\u2003', BreakBA},     // EM SPACE
		{'\u2004', BreakBA},     // THREE-PER-EM SPACE
		{'\u2005', BreakBA},     // FOUR-PER-EM SPACE
		{'\u2006', BreakBA},     // SIX-PER-EM SPACE
		{'\u2008', BreakBA},     // PUNCTUATION SPACE
		{'\u2009', BreakBA},     // THIN SPACE
		{'\u200A', BreakBA},     // HAIR SPACE
		{'\u205F', BreakBA},     // MEDIUM MATHEMATICAL SPACE
		{'\u3000', BreakBA},     // IDEOGRAPHIC SPACE
		{'\u0009', BreakBA},     //	TAB
		{'\u00AD', BreakBA},     //	SOFT HYPHEN (SHY)
		{'\u058A', BreakBA},     //	ARMENIAN HYPHEN
		{'\u2010', BreakBA},     //	HYPHEN
		{'\u2012', BreakBA},     //	FIGURE DASH
		{'\u2013', BreakBA},     //	EN DASH
		{'\u05BE', BreakBA},     //	HEBREW PUNCTUATION MAQAF
		{'\u0F0B', BreakBA},     //	TIBETAN MARK INTERSYLLABIC TSHEG
		{'\u1361', BreakBA},     //	ETHIOPIC WORDSPACE
		{'\u17D8', BreakBA},     //	KHMER SIGN BEYYAL
		{'\u17DA', BreakBA},     //	KHMER SIGN KOOMUUT
		{'\u2027', BreakBA},     //	HYPHENATION POINT
		{'\u007C', BreakBA},     //	VERTICAL LINE
		{'\u16EB', BreakBA},     //	RUNIC SINGLE PUNCTUATION
		{'\u16EC', BreakBA},     //	RUNIC MULTIPLE PUNCTUATION
		{'\u16ED', BreakBA},     //	RUNIC CROSS PUNCTUATION
		{'\u2056', BreakBA},     //	THREE DOT PUNCTUATION
		{'\u2058', BreakBA},     //	FOUR DOT PUNCTUATION
		{'\u2059', BreakBA},     //	FIVE DOT PUNCTUATION
		{'\u205A', BreakBA},     //	TWO DOT PUNCTUATION
		{'\u205B', BreakBA},     //	FOUR DOT MARK
		{'\u205D', BreakBA},     //	TRICOLON
		{'\u205E', BreakBA},     //	VERTICAL FOUR DOTS
		{'\u2E19', BreakBA},     //	PALM BRANCH
		{'\u2E2A', BreakBA},     //	TWO DOTS OVER ONE DOT PUNCTUATION
		{'\u2E2B', BreakBA},     //	ONE DOT OVER TWO DOTS PUNCTUATION
		{'\u2E2C', BreakBA},     //	SQUARED FOUR DOT PUNCTUATION
		{'\u2E2D', BreakBA},     //	FIVE DOT MARK
		{'\u2E30', BreakBA},     //	RING POINT
		{'\U00010100', BreakBA}, //	AEGEAN WORD SEPARATOR LINE
		{'\U00010101', BreakBA}, //	AEGEAN WORD SEPARATOR DOT
		{'\U00010102', BreakBA}, //	AEGEAN CHECK MARK
		{'\U0001039F', BreakBA}, //	UGARITIC WORD DIVIDER
		{'\U000103D0', BreakBA}, //	OLD PERSIAN WORD DIVIDER
		{'\U0001091F', BreakBA}, //	PHOENICIAN WORD SEPARATOR
		{'\U00012470', BreakBA}, //	CUNEIFORM PUNCTUATION SIGN OLD ASSYRIAN WORD DIVIDER
		{'\u0964', BreakBA},     //	DEVANAGARI DANDA
		{'\u0965', BreakBA},     //	DEVANAGARI DOUBLE DANDA
		{'\u0E5A', BreakBA},     //	THAI CHARACTER ANGKHANKHU
		{'\u0E5B', BreakBA},     //	THAI CHARACTER KHOMUT
		{'\u104A', BreakBA},     //	MYANMAR SIGN LITTLE SECTION
		{'\u104B', BreakBA},     //	MYANMAR SIGN SECTION
		{'\u1735', BreakBA},     //	PHILIPPINE SINGLE PUNCTUATION
		{'\u1736', BreakBA},     //	PHILIPPINE DOUBLE PUNCTUATION
		{'\u17D4', BreakBA},     //	KHMER SIGN KHAN
		{'\u17D5', BreakBA},     //	KHMER SIGN BARIYOOSAN
		{'\u1B5E', BreakBA},     //	BALINESE CARIK SIKI
		{'\u1B5F', BreakBA},     //	BALINESE CARIK PAREREN
		{'\uA8CE', BreakBA},     //	SAURASHTRA DANDA
		{'\uA8CF', BreakBA},     //	SAURASHTRA DOUBLE DANDA
		{'\uAA5D', BreakBA},     //	CHAM PUNCTUATION DANDA
		{'\uAA5E', BreakBA},     //	CHAM PUNCTUATION DOUBLE DANDA
		{'\uAA5F', BreakBA},     //	CHAM PUNCTUATION TRIPLE DANDA
		{'\U00010A56', BreakBA}, //	KHAROSHTHI PUNCTUATION DANDA
		{'\U00010A57', BreakBA}, //	KHAROSHTHI PUNCTUATION DOUBLE DANDA
		{'\u0F34', BreakBA},     //	TIBETAN MARK BSDUS RTAGS
		{'\u0F7F', BreakBA},     //	TIBETAN SIGN RNAM BCAD
		{'\u0F85', BreakBA},     //	TIBETAN MARK PALUTA
		{'\u0FBE', BreakBA},     //	TIBETAN KU RU KHA
		{'\u0FBF', BreakBA},     //	TIBETAN KU RU KHA BZHI MIG CAN
		{'\u0FD2', BreakBA},     //	TIBETAN MARK NYIS TSHEG
		{'\u1804', BreakBA},     //	MONGOLIAN COLON
		{'\u1805', BreakBA},     //	MONGOLIAN FOUR DOTS
		{'\u1B5A', BreakBA},     //	BALINESE PANTI
		{'\u1B5B', BreakBA},     //	BALINESE PAMADA
		{'\u1B5D', BreakBA},     //	BALINESE CARIK PAMUNGKAH
		{'\u1B60', BreakBA},     //	BALINESE PAMENENG
		{'\u1C3B', BreakBA},     //	LEPCHA PUNCTUATION TA-ROL
		{'\u1C3C', BreakBA},     //	LEPCHA PUNCTUATION NYET THYOOM TA-ROL
		{'\u1C3D', BreakBA},     //	LEPCHA PUNCTUATION CER-WA
		{'\u1C3E', BreakBA},     //	LEPCHA PUNCTUATION TSHOOK CER-WA
		{'\u1C3F', BreakBA},     //	LEPCHA PUNCTUATION TSHOOK
		{'\u1C7E', BreakBA},     //	OL CHIKI PUNCTUATION MUCAAD
		{'\u1C7F', BreakBA},     //	OL CHIKI PUNCTUATION DOUBLE MUCAAD
		{'\u2CFA', BreakBA},     //	COPTIC OLD NUBIAN DIRECT QUESTION MARK
		{'\u2CFB', BreakBA},     //	COPTIC OLD NUBIAN INDIRECT QUESTION MARK
		{'\u2CFC', BreakBA},     //	COPTIC OLD NUBIAN VERSE DIVIDER
		{'\u2CFF', BreakBA},     //	COPTIC MORPHOLOGICAL DIVIDER
		{'\u2E0E', BreakBA},     // EDITORIAL CORONIS
		{'\u2E17', BreakBA},     //	DOUBLE OBLIQUE HYPHEN
		{'\uA60D', BreakBA},     //	VAI COMMA
		{'\uA60F', BreakBA},     //	VAI QUESTION MARK
		{'\uA92E', BreakBA},     //	KAYAH LI SIGN CWI
		{'\uA92F', BreakBA},     //	KAYAH LI SIGN SHYA
		{'\U00010A50', BreakBA}, //	KHAROSHTHI PUNCTUATION DOT
		{'\U00010A51', BreakBA}, //	KHAROSHTHI PUNCTUATION SMALL CIRCLE
		{'\U00010A52', BreakBA}, //	KHAROSHTHI PUNCTUATION CIRCLE
		{'\U00010A53', BreakBA}, //	KHAROSHTHI PUNCTUATION CRESCENT BAR
		{'\U00010A54', BreakBA}, //	KHAROSHTHI PUNCTUATION MANGALAM
		{'\U00010A55', BreakBA}, //	KHAROSHTHI PUNCTUATION LOTUS
		// BB: Break Before (B)
		{'\u00B4', BreakBB}, //	ACUTE ACCENT
		{'\u1FFD', BreakBB}, //	GREEK OXIA
		{'\u02DF', BreakBB}, //	MODIFIER LETTER CROSS ACCENT
		{'\u02C8', BreakBB}, //	MODIFIER LETTER VERTICAL LINE
		{'\u02CC', BreakBB}, //	MODIFIER LETTER LOW VERTICAL LINE
		{'\u0F01', BreakBB}, //	TIBETAN MARK GTER YIG MGO TRUNCATED A
		{'\u0F02', BreakBB}, //	TIBETAN MARK GTER YIG MGO -UM RNAM BCAD MA
		{'\u0F03', BreakBB}, //	TIBETAN MARK GTER YIG MGO -UM GTER TSHEG MA
		{'\u0F04', BreakBB}, //	TIBETAN MARK INITIAL YIG MGO MDUN MA
		{'\u0F06', BreakBB}, //	TIBETAN MARK CARET YIG MGO PHUR SHAD MA
		{'\u0F07', BreakBB}, //	TIBETAN MARK YIG MGO TSHEG SHAD MA
		{'\u0F09', BreakBB}, //	TIBETAN MARK BSKUR YIG MGO
		{'\u0F0A', BreakBB}, //	TIBETAN MARK BKA- SHOG YIG MGO
		{'\u0FD0', BreakBB}, //	TIBETAN MARK BSKA- SHOG GI MGO RGYAN
		{'\u0FD1', BreakBB}, //	TIBETAN MARK MNYAM YIG GI MGO RGYAN
		{'\u0FD3', BreakBB}, //	TIBETAN MARK INITIAL BRDA RNYING YIG MGO MDUN MA
		{'\uA874', BreakBB}, //	PHAGS-PA SINGLE HEAD MARK
		{'\uA875', BreakBB}, //	PHAGS-PA DOUBLE HEAD MARK
		{'\u1806', BreakBB}, //	MONGOLIAN TODO SOFT HYPHEN

		//	B2: Break Opportunity Before and After (B/A/XP)
		{'\u2014', BreakB2}, //	EM DASH

		// BK: Mandatory Break (A) (Non-tailorable)

		{'\u000C', BreakBK}, //	FORM FEED (FF)
		{'\u000B', BreakBK}, //	LINE TABULATION (VT)
		{'\u2028', BreakBK}, //	LINE SEPARATOR
		{'\u2029', BreakBK}, //	PARAGRAPH SEPARATOR

		// CB: Contingent Break Opportunity (B/A)

		{'\uFFFC', BreakCB}, //	OBJECT REPLACEMENT CHARACTER

		// CJ: Conditional Japanese Starter

		{'\u3041', BreakCJ}, // 	Small hiragana
		{'\u30A1', BreakCJ}, // 	Small katakana
		{'\u30FC', BreakCJ}, //	KATAKANA-HIRAGANA PROLONGED SOUND MARK
		{'\uFF67', BreakCJ}, // 	Halfwidth variants

		// CL: Close Punctuation (XB)

		{'\u3001', BreakCL}, // 	IDEOGRAPHIC COMMA..IDEOGRAPHIC FULL STOP
		{'\uFE11', BreakCL}, //	PRESENTATION FORM FOR VERTICAL IDEOGRAPHIC COMMA
		{'\uFE12', BreakCL}, //	PRESENTATION FORM FOR VERTICAL IDEOGRAPHIC FULL STOP
		{'\uFE50', BreakCL}, //	SMALL COMMA
		{'\uFE52', BreakCL}, //	SMALL FULL STOP
		{'\uFF0C', BreakCL}, //	FULLWIDTH COMMA
		{'\uFF0E', BreakCL}, //	FULLWIDTH FULL STOP
		{'\uFF61', BreakCL}, //	HALFWIDTH IDEOGRAPHIC FULL STOP
		{'\uFF64', BreakCL}, //	HALFWIDTH IDEOGRAPHIC COMMA

		// CP: Closing Parenthesis (XB)

		{'\u0029', BreakCP}, //	RIGHT PARENTHESIS
		{'\u005D', BreakCP}, //	RIGHT SQUARE BRACKET

		// CR: Carriage Return (A) (Non-tailorable)
		{'\u000D', BreakCR}, //	CARRIAGE RETURN (CR)

		// EB: Emoji Base (B/A)
		{'\U0001F466', BreakEB}, //	BOY
		{'\U0001F478', BreakEB}, //	PRINCESS
		{'\U0001F6B4', BreakEB}, //	BICYCLIST

		// EM: Emoji Modifier (A)

		{'\U0001F3FB', BreakEM}, // EMOJI MODIFIER FITZPATRICK TYPE-1-2
		{'\U0001F3FF', BreakEM}, // EMOJI MODIFIER FITZPATRICK TYPE-6

		// EX: Exclamation/Interrogation (XB)

		{'\u0021', BreakEX}, //	EXCLAMATION MARK
		{'\u003F', BreakEX}, //	QUESTION MARK
		{'\u05C6', BreakEX}, //	HEBREW PUNCTUATION NUN HAFUKHA
		{'\u061B', BreakEX}, //	ARABIC SEMICOLON
		{'\u061E', BreakEX}, //	ARABIC TRIPLE DOT PUNCTUATION MARK
		{'\u061F', BreakEX}, //	ARABIC QUESTION MARK
		{'\u06D4', BreakEX}, //	ARABIC FULL STOP
		{'\u07F9', BreakEX}, //	NKO EXCLAMATION MARK
		{'\u0F0D', BreakEX}, //	TIBETAN MARK SHAD
		{'\uFF01', BreakEX}, //	FULLWIDTH EXCLAMATION MARK
		{'\uFF1F', BreakEX}, //	FULLWIDTH QUESTION MARK

		// GL: Non-breaking (“Glue”) (XB/XA) (Non-tailorable)

		{'\u00A0', BreakGL}, //	NO-BREAK SPACE (NBSP)
		{'\u202F', BreakGL}, //	NARROW NO-BREAK SPACE (NNBSP)
		{'\u180E', BreakGL}, //	MONGOLIAN VOWEL SEPARATOR (MVS)
		{'\u034F', BreakGL}, //	COMBINING GRAPHEME JOINER
		{'\u2007', BreakGL}, //	FIGURE SPACE
		{'\u2011', BreakGL}, //	NON-BREAKING HYPHEN
		{'\u0F08', BreakGL}, //	TIBETAN MARK SBRUL SHAD
		{'\u0F0C', BreakGL}, //	TIBETAN MARK DELIMITER TSHEG BSTAR
		{'\u0F12', BreakGL}, //	TIBETAN MARK RGYA GRAM SHAD
		{'\u035C', BreakGL}, // COMBINING DOUBLE BREVE BELOW
		{'\u0362', BreakGL}, //	 COMBINING DOUBLE RIGHTWARDS ARROW BELOW

		// HY: Hyphen (XA)
		{'\u002D', BreakHY}, //	HYPHEN-MINUS

		// ID: Ideographic (B/A)

		{'\u2E80', BreakID},     // 	CJK, Kangxi Radicals, Ideographic Description Symbols
		{'\u30A2', BreakID},     // 	Katakana (except small characters)
		{'\u3400', BreakID},     // 	CJK Unified Ideographs Extension A
		{'\u4E00', BreakID},     // 	CJK Unified Ideographs
		{'\uF900', BreakID},     // 	CJK Compatibility Ideographs
		{'\u3400', BreakID},     // 	CJK Unified Ideographs Extension A
		{'\u4E00', BreakID},     // 	CJK Unified Ideographs
		{'\uF900', BreakID},     // 	CJK Compatibility Ideographs
		{'\U00020000', BreakID}, // 	Plane 2
		{'\U00030000', BreakID}, // 	Plane 3
		{'\U0001F000', BreakID}, // 	Plane 1 range

		// IN: Inseparable Characters (XP)

		{'\u2024', BreakIN}, //	ONE DOT LEADER
		{'\u2025', BreakIN}, //	TWO DOT LEADER
		{'\u2026', BreakIN}, //	HORIZONTAL ELLIPSIS
		{'\uFE19', BreakIN}, //	PRESENTATION FORM FOR VERTICAL HORIZONTAL ELLIPSIS

		// IS: Infix Numeric Separator (XB)

		{'\u002C', BreakIS}, //	COMMA
		{'\u002E', BreakIS}, //	FULL STOP
		{'\u003A', BreakIS}, //	COLON
		{'\u003B', BreakIS}, //	SEMICOLON
		{'\u037E', BreakIS}, //	GREEK QUESTION MARK (canonically equivalent to 003B)
		{'\u0589', BreakIS}, //	ARMENIAN FULL STOP
		{'\u060C', BreakIS}, //	ARABIC COMMA
		{'\u060D', BreakIS}, //	ARABIC DATE SEPARATOR
		{'\u07F8', BreakIS}, //	NKO COMMA
		{'\u2044', BreakIS}, //	FRACTION SLASH
		{'\uFE10', BreakIS}, //	PRESENTATION FORM FOR VERTICAL COMMA
		{'\uFE13', BreakIS}, //	PRESENTATION FORM FOR VERTICAL COLON
		{'\uFE14', BreakIS}, //	PRESENTATION FORM FOR VERTICAL SEMICOLON

		// LF: Line Feed (A) (Non-tailorable)
		{'\u000A', BreakLF}, //	LINE FEED (LF)

		// NL: Next Line (A) (Non-tailorable)
		{'\u0085', BreakNL}, //	NEXT LINE (NEL)

		// NS: Nonstarters (XB)

		{'\u17D6', BreakNS}, //	KHMER SIGN CAMNUC PII KUUH
		{'\u203C', BreakNS}, //	DOUBLE EXCLAMATION MARK
		{'\u203D', BreakNS}, //	INTERROBANG
		{'\u2047', BreakNS}, //	DOUBLE QUESTION MARK
		{'\u2048', BreakNS}, //	QUESTION EXCLAMATION MARK
		{'\u2049', BreakNS}, //	EXCLAMATION QUESTION MARK
		{'\u3005', BreakNS}, //	IDEOGRAPHIC ITERATION MARK
		{'\u301C', BreakNS}, //	WAVE DASH
		{'\u303C', BreakNS}, //	MASU MARK
		{'\u303B', BreakNS}, //	VERTICAL IDEOGRAPHIC ITERATION MARK
		{'\u309B', BreakNS}, // 	KATAKANA-HIRAGANA VOICED SOUND MARK..HIRAGANA VOICED ITERATION MARK
		{'\u30A0', BreakNS}, //	KATAKANA-HIRAGANA DOUBLE HYPHEN
		{'\u30FB', BreakNS}, //	KATAKANA MIDDLE DOT
		{'\u30FD', BreakNS}, // 	KATAKANA ITERATION MARK..KATAKANA VOICED ITERATION MARK
		{'\uFE54', BreakNS}, // 	SMALL SEMICOLON..SMALL COLON
		{'\uFF1A', BreakNS}, // 	FULLWIDTH COLON.. FULLWIDTH SEMICOLON
		{'\uFF65', BreakNS}, //	HALFWIDTH KATAKANA MIDDLE DOT
		{'\uFF9E', BreakNS}, // 	HALFWIDTH KATAKANA VOICED SOUND MARK..HALFWIDTH KATAKANA SEMI-VOICED SOUND MARK

		// NU: Numeric (XP)

		{'\u066B', BreakNU}, //	ARABIC DECIMAL SEPARATOR
		{'\u066C', BreakNU}, //	ARABIC THOUSANDS SEPARATOR

		// OP: Open Punctuation (XA)

		{'\u00A1', BreakOP}, //	INVERTED EXCLAMATION MARK
		{'\u00BF', BreakOP}, //	INVERTED QUESTION MARK
		{'\u2E18', BreakOP}, //	INVERTED INTERROBANG

		// PO: Postfix Numeric (XB)

		{'\u0025', BreakPO}, //	PERCENT SIGN
		{'\u00A2', BreakPO}, //	CENT SIGN
		{'\u00B0', BreakPO}, //	DEGREE SIGN
		{'\u060B', BreakPO}, //	AFGHANI SIGN
		{'\u066A', BreakPO}, //	ARABIC PERCENT SIGN
		{'\u2030', BreakPO}, //	PER MILLE SIGN
		{'\u2031', BreakPO}, //	PER TEN THOUSAND SIGN
		{'\u2032', BreakPO}, // 	PRIME
		{'\u20A7', BreakPO}, //	PESETA SIGN
		{'\u2103', BreakPO}, //	DEGREE CELSIUS
		{'\u2109', BreakPO}, //	DEGREE FAHRENHEIT
		{'\uFDFC', BreakPO}, //	RIAL SIGN
		{'\uFE6A', BreakPO}, //	SMALL PERCENT SIGN
		{'\uFF05', BreakPO}, //	FULLWIDTH PERCENT SIGN
		{'\uFFE0', BreakPO}, //	FULLWIDTH CENT SIGN

		// PR: Prefix Numeric (XA)

		{'\u002B', BreakPR}, //	PLUS SIGN
		{'\u005C', BreakPR}, //	REVERSE SOLIDUS
		{'\u00B1', BreakPR}, //	PLUS-MINUS SIGN
		{'\u2116', BreakPR}, //	NUMERO SIGN
		{'\u2212', BreakPR}, //	MINUS SIGN
		{'\u2213', BreakPR}, //	MINUS-OR-PLUS SIGN

		// QU: Quotation (XB/XA)

		{'\u0022', BreakQU}, //	QUOTATION MARK
		{'\u0027', BreakQU}, //	APOSTROPHE
		{'\u275B', BreakQU}, //	HEAVY SINGLE TURNED COMMA QUOTATION MARK ORNAMENT
		{'\u275C', BreakQU}, //	HEAVY SINGLE COMMA QUOTATION MARK ORNAMENT
		{'\u275D', BreakQU}, //	HEAVY DOUBLE TURNED COMMA QUOTATION MARK ORNAMENT
		{'\u275E', BreakQU}, //	HEAVY DOUBLE COMMA QUOTATION MARK ORNAMENT
		{'\u2E00', BreakQU}, //	RIGHT ANGLE SUBSTITUTION MARKER
		{'\u2E06', BreakQU}, //	RAISED INTERPOLATION MARKER
		{'\u2E0B', BreakQU}, //	RAISED SQUARE
		// RI: Regional Indicator (B/A/XP)

		{'\U0001F1E6', BreakRI}, // REGIONAL INDICATOR SYMBOL LETTER A
		{'\U0001F1FF', BreakRI}, // REGIONAL INDICATOR SYMBOL LETTER Z

		// SA: Complex-Context Dependent (South East Asian) (P)

		{'\u1000', BreakSA},     //	Myanmar
		{'\u1780', BreakSA},     //	Khmer
		{'\u1950', BreakSA},     //	Tai Le
		{'\u1980', BreakSA},     //	New Tai Lue
		{'\u1A20', BreakSA},     //	Tai Tham
		{'\uA9E0', BreakSA},     //	Myanmar Extended-B
		{'\uAA60', BreakSA},     //	Myanmar Extended-A
		{'\uAA80', BreakSA},     //	Tai Viet
		{'\U00011700', BreakSA}, // 	Ahom

		// SP: Space (A) (Non-tailorable)

		{'\u0020', BreakSP}, //	SPACE (SP)

		// SY: Symbols Allowing Break After (A)

		{'\u002F', BreakSY}, //	SOLIDUS

		// WJ: Word Joiner (XB/XA) (Non-tailorable)

		{'\u2060', BreakWJ}, //	WORD JOINER (WJ)
		{'\uFEFF', BreakWJ}, //	ZERO WIDTH NO-BREAK SPACE (ZWNBSP)

		// ZW: Zero Width Space (A) (Non-tailorable)
		{'\u200B', BreakZW}, //	ZERO WIDTH SPACE (ZWSP)

		// ZWJ: Zero Width Joiner (XA/XB) (Non-tailorable)
		{'\u200D', BreakZWJ}, //	ZERO WIDTH JOINER (ZWJ)

	}
	for _, tt := range tests {
		if got := LookupLineBreakClass(tt.args); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("LookupLineBreakClass(U+%x) = %p, want %p", tt.args, got, tt.want)
		}
	}
}

func TestLookupGraphemeBreakClass(t *testing.T) {
	// See https://www.unicode.org/reports/tr29/#Grapheme_Cluster_Break_Property_Values
	tests := []struct {
		args rune
		want *unicode.RangeTable
	}{
		{-1, nil},
		{'a', nil},
		// CR
		{'\u000D', GraphemeBreakCR}, // CARRIAGE RETURN (CR)
		// LF
		{'\u000A', GraphemeBreakLF}, // LINE FEED (LF)
		// Extend
		{'\u200C', GraphemeBreakExtend}, // ZERO WIDTH NON-JOINER
		// ZWJ
		{'\u200D', GraphemeBreakZWJ}, // ZERO WIDTH JOINER
		// RI
		{'\U0001F1E6', GraphemeBreakRegional_Indicator}, // REGIONAL INDICATOR SYMBOL LETTER A
		// SpacingMark
		{'\u0E33', GraphemeBreakSpacingMark}, // ( ำ ) THAI CHARACTER SARA AM
		{'\u0EB3', GraphemeBreakSpacingMark}, // ( ຳ ) LAO VOWEL SIGN AM

		// L
		{'\u1100', GraphemeBreakL}, // ( ᄀ ) HANGUL CHOSEONG KIYEOK
		{'\u115F', GraphemeBreakL}, // ( ᅟ ) HANGUL CHOSEONG FILLER
		{'\uA960', GraphemeBreakL}, // ( ꥠ ) HANGUL CHOSEONG TIKEUT-MIEUM
		{'\uA97C', GraphemeBreakL}, // ( ꥼ ) HANGUL CHOSEONG SSANGYEORINHIEUH
		// V
		{'\u1160', GraphemeBreakV}, // ( ᅠ ) HANGUL JUNGSEONG FILLER
		{'\u11A2', GraphemeBreakV}, // ( ᆢ ) HANGUL JUNGSEONG SSANGARAEA
		{'\uD7B0', GraphemeBreakV}, // ( ힰ ) HANGUL JUNGSEONG O-YEO
		{'\uD7C6', GraphemeBreakV}, // ( ퟆ ) HANGUL JUNGSEONG ARAEA-E
		// T
		{'\u11A8', GraphemeBreakT}, // ( ᆨ ) HANGUL JONGSEONG KIYEOK
		{'\u11F9', GraphemeBreakT}, // ( ᇹ ) HANGUL JONGSEONG YEORINHIEUH
		{'\uD7CB', GraphemeBreakT}, // ( ퟋ ) HANGUL JONGSEONG NIEUN-RIEUL
		{'\uD7FB', GraphemeBreakT}, // ( ퟻ ) HANGUL JONGSEONG PHIEUPH-THIEUTH
		// LV
		{'\uAC00', GraphemeBreakLV}, // ( 가 ) HANGUL SYLLABLE GA
		{'\uAC1C', GraphemeBreakLV}, // ( 개 ) HANGUL SYLLABLE GAE
		{'\uAC38', GraphemeBreakLV}, // ( 갸 ) HANGUL SYLLABLE GYA
		// LVT
		{'\uAC01', GraphemeBreakLVT}, // ( 각 ) HANGUL SYLLABLE GAG
		{'\uAC02', GraphemeBreakLVT}, // ( 갂 ) HANGUL SYLLABLE GAGG
		{'\uAC03', GraphemeBreakLVT}, // ( 갃 ) HANGUL SYLLABLE GAGS
		{'\uAC04', GraphemeBreakLVT}, // ( 간 ) HANGUL SYLLABLE GAN
	}
	for _, tt := range tests {
		if got := LookupGraphemeBreakClass(tt.args); got != tt.want {
			t.Errorf("LookupGraphemeBreakClass(%x) = %p, want %p", tt.args, got, tt.want)
		}
	}
}

func TestLookupWordBreakClass(t *testing.T) {
	// these runes have changed from Unicode v15.0 to 15.1
	tu.Assert(t, LookupWordBreakClass(0x6DD) == WordBreakNumeric)
	tu.Assert(t, LookupWordBreakClass(0x661) == WordBreakNumeric)
}

func TestLookupMirrorChar(t *testing.T) {
	tests := []struct {
		args  rune
		want  rune
		want1 bool
	}{
		{'a', 'a', false},
		{'\u2231', '\u2231', false},
		{'\u0028', '\u0029', true},
		{'\u0029', '\u0028', true},
		{'\u22E0', '\u22E1', true},
	}
	for _, tt := range tests {
		got, got1 := LookupMirrorChar(tt.args)
		if got != tt.want {
			t.Errorf("LookupMirrorChar() got = %v, want %v", got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("LookupMirrorChar() got1 = %v, want %v", got1, tt.want1)
		}
	}
}

func TestLookupVerticalOrientation(t *testing.T) {
	tests := []struct {
		s              language.Script
		r              rune
		wantIsSideways bool
	}{
		{language.Cyrillic, '\u0400', true},
		{language.Latin, 'A', true},
		{language.Latin, '\uFF21', false},
		{language.Katakana, 'も', false},
		{language.Katakana, '\uFF89', true},
		{language.Hangul, '\uFFAB', true},
	}
	for _, tt := range tests {
		if gotIsSideways := LookupVerticalOrientation(tt.s).Orientation(tt.r); gotIsSideways != tt.wantIsSideways {
			t.Errorf("LookupVerticalOrientation(%s) = %v, want %v", string(tt.r), gotIsSideways, tt.wantIsSideways)
		}
	}
}
