package unicodedata

import (
	"testing"
	"unicode"

	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

var composeTests = []struct {
	a, b rune
	ok   bool
	ab   rune
}{
	{0x41, 0x300, true, 0xC0},

	// Not composable
	{0x0041, 0x0042, false, 0},
	{0x0041, 0, false, 0},
	{0x0066, 0x0069, false, 0},

	// Singletons should not compose
	{0x212B, 0, false, 0},
	{0x00C5, 0, false, 0},
	{0x2126, 0, false, 0},
	{0x03A9, 0, false, 0},

	// Non-starter pairs should not compose
	{0x0308, 0x0301, false, 0}, // !0x034
	{0x0F71, 0x0F72, false, 0}, // !0x0F7

	// Pairs
	{0x0041, 0x030A, true, 0x00C5},
	{0x006F, 0x0302, true, 0x00F4},
	{0x1E63, 0x0307, true, 0x1E69},
	{0x0073, 0x0323, true, 0x1E63},
	{0x0064, 0x0307, true, 0x1E0B},
	{0x0064, 0x0323, true, 0x1E0D},

	// Hangul
	{0xD4CC, 0x11B6, true, 0xD4DB},
	{0x1111, 0x1171, true, 0xD4CC},
	{0xCE20, 0x11B8, true, 0xCE31},
	{0x110E, 0x1173, true, 0xCE20},

	{0xAC00, 0x11A7, false, 0},
	{0xAC00, 0x11A8, true, 0xAC01},
	{0xAC01, 0x11A8, false, 0},
}

var decomposeTests = []struct {
	ab   rune
	ok   bool
	a, b rune
}{
	{0xC0, true, 0x41, 0x300},
	{0xAC00, true, 0x1100, 0x1161},
	{0x01C4, false, 0x01C4, 0},
	{0x320E, false, 0x320E, 0},

	// Not decomposable
	{0x0041, false, 0x0041, 0},
	{0xFB01, false, 0xFB01, 0},
	{0x1F1EF, false, 0x1F1EF, 0},

	// Singletons
	{0x212B, true, 0x00C5, 0},
	{0x2126, true, 0x03A9, 0},

	// Non-starter pairs decompose, but not compose
	{0x0344, true, 0x0308, 0x0301},
	{0x0F73, true, 0x0F71, 0x0F72},

	// Pairs
	{0x00C5, true, 0x0041, 0x030A},
	{0x00F4, true, 0x006F, 0x0302},
	{0x1E69, true, 0x1E63, 0x0307},
	{0x1E63, true, 0x0073, 0x0323},
	{0x1E0B, true, 0x0064, 0x0307},
	{0x1E0D, true, 0x0064, 0x0323},

	// Hangul
	{0xD4DB, true, 0xD4CC, 0x11B6},
	{0xD4CC, true, 0x1111, 0x1171},
	{0xCE31, true, 0xCE20, 0x11B8},
	{0xCE20, true, 0x110E, 0x1173},
}

func TestUnicodeNormalization(t *testing.T) {
	for _, test := range composeTests {
		ab, ok := Compose(test.a, test.b)
		tu.Assert(t, ab == test.ab && ok == test.ok)
	}
	for _, test := range decomposeTests {
		a, b, ok := Decompose(test.ab)
		tu.Assert(t, a == test.a && b == test.b && ok == test.ok)
	}
}

var generalCategoryTests = []struct {
	args rune
	want GeneralCategory
}{
	{0x70f, Cf},
	{'a', Ll},
	{'.', Po},
	{'カ', Lo},
	{'🦳', So},
	{'\U0001F3FF', Sk},
	{'\U0001F02C', 0},
	{-1, 0},
	// the following cases are taken from Harfbuzz
	{0x000D, Cc},
	{0x200E, Cf},
	{0x0378, 0},
	{0xE000, Co},
	{0xD800, Cs},
	{0x0061, Ll},
	{0x02B0, Lm},
	{0x3400, Lo},
	{0x01C5, Lt},
	{0xFF21, Lu},
	{0x0903, Mc},
	{0x20DD, Me},
	{0xA806, Mn},
	{0xFF10, Nd},
	{0x16EE, Nl},
	{0x17F0, No},
	{0x005F, Pc},
	{0x058A, Pd},
	{0x0F3B, Pe},
	{0x2019, Pf},
	{0x2018, Pi},
	{0x2016, Po},
	{0x0F3A, Ps},
	{0x20A0, Sc},
	{0x309B, Sk},
	{0xFB29, Sm},
	{0x00A6, So},
	{0x2028, Zl},
	{0x2029, Zp},
	{0x202F, Zs},
	{0x111111, 0},
	/* Unicode-5.2 character additions */
	{0x1F131, So},
	/* Unicode-6.0 character additions */
	{0x0620, Lo},
	/* Unicode-6.1 character additions */
	{0x058F, Sc},
	/* Unicode-6.2 character additions */
	{0x20BA, Sc},
	/* Unicode-6.3 character additions */
	{0x061C, Cf},
	/* Unicode-7.0 character additions */
	{0x058D, So},
	/* Unicode-8.0 character additions */
	{0x08E3, Mn},
	/* Unicode-9.0 character additions */
	{0x08D4, Mn},
	/* Unicode-10.0 character additions */
	{0x09FD, Po},
	/* Unicode-11.0 character additions */
	{0x0560, Ll},
	/* Unicode-12.0 character additions */
	{0x0C77, Po},
	/* Unicode-12.1 character additions */
	{0x32FF, So},
	/* Unicode-13.0 character additions */
	{0x08BE, Lo},
	/* Unicode-14.0 character additions */
	{0x20C0, Sc},
	/* Unicode-15.0 character additions */
	{0x0CF3, Mc},
	/* Unicode-15.1 character additions */
	{0x31EF, So},
	/* Unicode-16.0 character additions */
	{0x10D6E, Pd},
	/* Unicode-17.0 character additions */
	{0x11DE0, Nd},
}

func TestLookupType(t *testing.T) {
	for _, tt := range generalCategoryTests {
		if got := LookupType(tt.args); got != tt.want {
			t.Errorf("LookupType(%s = %0x) = %v, want %v", string(tt.args), tt.args, got, tt.want)
		}
	}
}

var cccTests = []struct {
	args rune
	want uint8
}{
	// reference values are from https://www.compart.com/en/unicode/combining/
	{-1, 0},
	{'a', 0},
	{'\u093F', 0},
	{'\u0E40', 0},
	{'\u093c', 7},
	{'\u1039', 9},
	{'\u0f7b', 130},
	{'\u1cdd', 220},
	{'\u0369', 230},
	// these are copied from Harfbuzz
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
	/* Unicode-16.0 character additions */
	{0x0897, 230},
	/* Unicode-17.0 character additions */
	{0x1ACF, 230},
	{0x111111, 0},
}

func TestLookupCombiningClass(t *testing.T) {
	for _, tt := range cccTests {
		if got := LookupCombiningClass(tt.args); got != tt.want {
			t.Errorf("LookupCombiningClass(%s) = %v, want %v", string(tt.args), got, tt.want)
		}
	}
}

var extendedPictoTests = []struct {
	r       rune
	isEmoji bool
}{
	{0, false},
	{1, false},
	{2, false},
	{3, false},
	{4, false},
	{'a', false},
	{'H', false},
	{0x00A9, true},
	{0x00AE, true},
	{0x203C, true},
	{0x2049, true},
	{0x2122, true},
	{0x2139, true},
	{0x2194, true},
	{0x21A9, true},
	{0x231A, true},
	{0x2328, true},
	{0x23CF, true},
	{0x23E9, true},
	{0x23ED, true},
	{0x23EF, true},
	{0x23F0, true},
	{0x23F1, true},
	{0x23F3, true},
	{0x23F8, true},
	{0x24C2, true},
	{0x25AA, true},
	{0x25B6, true},
	{0x25C0, true},
	{0x25FB, true},
	{0x2600, true},
	{0x2602, true},
	{0x2604, true},
	{0x260E, true},
	{0x2611, true},
	{0x2614, true},
	{0x2618, true},
	{0x261D, true},
	{0x2620, true},
	{0x2622, true},
	{0x2626, true},
	{0x262A, true},
	{0x262E, true},
	{0x262F, true},
	{0x2638, true},
	{0x263A, true},
	{0x2640, true},
	{0x2642, true},
	{0x2648, true},
	{0x265F, true},
	{0x2660, true},
	{0x2663, true},
	{0x2665, true},
	{0x2668, true},
	{0x267B, true},
	{0x267E, true},
	{0x267F, true},
	{0x2692, true},
	{0x2693, true},
	{0x2694, true},
	{0x2695, true},
	{0x2696, true},
	{0x2699, true},
	{0x269B, true},
	{0x26A0, true},
	{0x26A7, true},
	{0x26AA, true},
	{0x26B0, true},
	{0x26BD, true},
	{0x26C4, true},
	{0x26C8, true},
	{0x26CE, true},
	{0x26CF, true},
	{0x26D1, true},
	{0x26D3, true},
	{0x26D4, true},
	{0x26E9, true},
	{0x26EA, true},
	{0x26F0, true},
	{0x26F2, true},
	{0x26F4, true},
	{0x26F5, true},
	{0x26F7, true},
	{0x26FA, true},
	{0x26FD, true},
	{0x2702, true},
	{0x2705, true},
	{0x2708, true},
	{0x270D, true},
	{0x270F, true},
	{0x2712, true},
	{0x2714, true},
	{0x2716, true},
	{0x271D, true},
	{0x2721, true},
	{0x2728, true},
	{0x2733, true},
	{0x2744, true},
	{0x2747, true},
	{0x274C, true},
	{0x274E, true},
	{0x2753, true},
	{0x2757, true},
	{0x2763, true},
	{0x2764, true},
	{0x2795, true},
	{0x27A1, true},
	{0x27B0, true},
	{0x27BF, true},
	{0x2934, true},
	{0x2B05, true},
	{0x2B1B, true},
	{0x2B50, true},
	{0x2B55, true},
	{0x3030, true},
	{0x303D, true},
	{0x3297, true},
	{0x3299, true},
	{0x1F004, true},
	{0x1F02C, true},
	{0x1F094, true},
	{0x1F0AF, true},
	{0x1F0C0, true},
	{0x1F0CF, true},
	{0x1F0D0, true},
	{0x1F0F6, true},
	{0x1F170, true},
	{0x1F17E, true},
	{0x1F18E, true},
	{0x1F191, true},
	{0x1F1AE, true},
	{0x1F201, true},
	{0x1F203, true},
	{0x1F21A, true},
	{0x1F22F, true},
	{0x1F232, true},
	{0x1F23C, true},
	{0x1F249, true},
	{0x1F250, true},
	{0x1F252, true},
	{0x1F266, true},
	{0x1F300, true},
	{0x1F30D, true},
	{0x1F30F, true},
	{0x1F310, true},
	{0x1F311, true},
	{0x1F312, true},
	{0x1F313, true},
	{0x1F316, true},
	{0x1F319, true},
	{0x1F31A, true},
	{0x1F31B, true},
	{0x1F31C, true},
	{0x1F31D, true},
	{0x1F31F, true},
	{0x1F321, true},
	{0x1F324, true},
	{0x1F32D, true},
	{0x1F330, true},
	{0x1F332, true},
	{0x1F334, true},
	{0x1F336, true},
	{0x1F337, true},
	{0x1F34B, true},
	{0x1F34C, true},
	{0x1F350, true},
	{0x1F351, true},
	{0x1F37C, true},
	{0x1F37D, true},
	{0x1F37E, true},
	{0x1F380, true},
	{0x1F396, true},
	{0x1F399, true},
	{0x1F39E, true},
	{0x1F3A0, true},
	{0x1F3C5, true},
	{0x1F3C6, true},
	{0x1F3C7, true},
	{0x1F3C8, true},
	{0x1F3C9, true},
	{0x1F3CA, true},
	{0x1F3CB, true},
	{0x1F3CF, true},
	{0x1F3D4, true},
	{0x1F3E0, true},
	{0x1F3E4, true},
	{0x1F3E5, true},
	{0x1F3F3, true},
	{0x1F3F4, true},
	{0x1F3F5, true},
	{0x1F3F7, true},
	{0x1F3F8, true},
	{0x1F400, true},
	{0x1F408, true},
	{0x1F409, true},
	{0x1F40C, true},
	{0x1F40F, true},
	{0x1F411, true},
	{0x1F413, true},
	{0x1F414, true},
	{0x1F415, true},
	{0x1F416, true},
	{0x1F417, true},
	{0x1F42A, true},
	{0x1F42B, true},
	{0x1F43F, true},
	{0x1F440, true},
	{0x1F441, true},
	{0x1F442, true},
	{0x1F465, true},
	{0x1F466, true},
	{0x1F46C, true},
	{0x1F46E, true},
	{0x1F4AD, true},
	{0x1F4AE, true},
	{0x1F4B6, true},
	{0x1F4B8, true},
	{0x1F4EC, true},
	{0x1F4EE, true},
	{0x1F4EF, true},
	{0x1F4F0, true},
	{0x1F4F5, true},
	{0x1F4F6, true},
	{0x1F4F8, true},
	{0x1F4F9, true},
	{0x1F4FD, true},
	{0x1F4FF, true},
	{0x1F503, true},
	{0x1F504, true},
	{0x1F508, true},
	{0x1F509, true},
	{0x1F50A, true},
	{0x1F515, true},
	{0x1F516, true},
	{0x1F52C, true},
	{0x1F52E, true},
	{0x1F549, true},
	{0x1F54B, true},
	{0x1F550, true},
	{0x1F55C, true},
	{0x1F56F, true},
	{0x1F573, true},
	{0x1F57A, true},
	{0x1F587, true},
	{0x1F58A, true},
	{0x1F590, true},
	{0x1F595, true},
	{0x1F5A4, true},
	{0x1F5A5, true},
	{0x1F5A8, true},
	{0x1F5B1, true},
	{0x1F5BC, true},
	{0x1F5C2, true},
	{0x1F5D1, true},
	{0x1F5DC, true},
	{0x1F5E1, true},
	{0x1F5E3, true},
	{0x1F5E8, true},
	{0x1F5EF, true},
	{0x1F5F3, true},
	{0x1F5FA, true},
	{0x1F5FB, true},
	{0x1F600, true},
	{0x1F601, true},
	{0x1F607, true},
	{0x1F609, true},
	{0x1F60E, true},
	{0x1F60F, true},
	{0x1F610, true},
	{0x1F611, true},
	{0x1F612, true},
	{0x1F615, true},
	{0x1F616, true},
	{0x1F617, true},
	{0x1F618, true},
	{0x1F619, true},
	{0x1F61A, true},
	{0x1F61B, true},
	{0x1F61C, true},
	{0x1F61F, true},
	{0x1F620, true},
	{0x1F626, true},
	{0x1F628, true},
	{0x1F62C, true},
	{0x1F62D, true},
	{0x1F62E, true},
	{0x1F630, true},
	{0x1F634, true},
	{0x1F635, true},
	{0x1F636, true},
	{0x1F637, true},
	{0x1F641, true},
	{0x1F645, true},
	{0x1F680, true},
	{0x1F681, true},
	{0x1F683, true},
	{0x1F686, true},
	{0x1F687, true},
	{0x1F688, true},
	{0x1F689, true},
	{0x1F68A, true},
	{0x1F68C, true},
	{0x1F68D, true},
	{0x1F68E, true},
	{0x1F68F, true},
	{0x1F690, true},
	{0x1F691, true},
	{0x1F694, true},
	{0x1F695, true},
	{0x1F696, true},
	{0x1F697, true},
	{0x1F698, true},
	{0x1F699, true},
	{0x1F69B, true},
	{0x1F6A2, true},
	{0x1F6A3, true},
	{0x1F6A4, true},
	{0x1F6A6, true},
	{0x1F6A7, true},
	{0x1F6AE, true},
	{0x1F6B2, true},
	{0x1F6B3, true},
	{0x1F6B6, true},
	{0x1F6B7, true},
	{0x1F6B9, true},
	{0x1F6BF, true},
	{0x1F6C0, true},
	{0x1F6C1, true},
	{0x1F6CB, true},
	{0x1F6CC, true},
	{0x1F6CD, true},
	{0x1F6D0, true},
	{0x1F6D1, true},
	{0x1F6D5, true},
	{0x1F6D6, true},
	{0x1F6D8, true},
	{0x1F6D9, true},
	{0x1F6DC, true},
	{0x1F6DD, true},
	{0x1F6E0, true},
	{0x1F6E9, true},
	{0x1F6EB, true},
	{0x1F6ED, true},
	{0x1F6F0, true},
	{0x1F6F3, true},
	{0x1F6F4, true},
	{0x1F6F7, true},
	{0x1F6F9, true},
	{0x1F6FA, true},
	{0x1F6FB, true},
	{0x1F6FD, true},
	{0x1F7DA, true},
	{0x1F7E0, true},
	{0x1F7EC, true},
	{0x1F7F0, true},
	{0x1F7F1, true},
	{0x1F80C, true},
	{0x1F848, true},
	{0x1F85A, true},
	{0x1F888, true},
	{0x1F8AE, true},
	{0x1F8BC, true},
	{0x1F8C2, true},
	{0x1F8D9, true},
	{0x1F90C, true},
	{0x1F90D, true},
	{0x1F910, true},
	{0x1F919, true},
	{0x1F91F, true},
	{0x1F920, true},
	{0x1F928, true},
	{0x1F930, true},
	{0x1F931, true},
	{0x1F933, true},
	{0x1F93C, true},
	{0x1F93F, true},
	{0x1F940, true},
	{0x1F947, true},
	{0x1F94C, true},
	{0x1F94D, true},
	{0x1F950, true},
	{0x1F95F, true},
	{0x1F96C, true},
	{0x1F971, true},
	{0x1F972, true},
	{0x1F973, true},
	{0x1F977, true},
	{0x1F979, true},
	{0x1F97A, true},
	{0x1F97B, true},
	{0x1F97C, true},
	{0x1F980, true},
	{0x1F985, true},
	{0x1F992, true},
	{0x1F998, true},
	{0x1F9A3, true},
	{0x1F9A5, true},
	{0x1F9AB, true},
	{0x1F9AE, true},
	{0x1F9B0, true},
	{0x1F9BA, true},
	{0x1F9C0, true},
	{0x1F9C1, true},
	{0x1F9C3, true},
	{0x1F9CB, true},
	{0x1F9CC, true},
	{0x1F9CD, true},
	{0x1F9D0, true},
	{0x1F9E7, true},
	{0x1FA58, true},
	{0x1FA6E, true},
	{0x1FA70, true},
	{0x1FA74, true},
	{0x1FA75, true},
	{0x1FA78, true},
	{0x1FA7B, true},
	{0x1FA7D, true},
	{0x1FA80, true},
	{0x1FA83, true},
	{0x1FA87, true},
	{0x1FA89, true},
	{0x1FA8A, true},
	{0x1FA8B, true},
	{0x1FA8E, true},
	{0x1FA8F, true},
	{0x1FA90, true},
	{0x1FA96, true},
	{0x1FAA9, true},
	{0x1FAAD, true},
	{0x1FAB0, true},
	{0x1FAB7, true},
	{0x1FABB, true},
	{0x1FABE, true},
	{0x1FABF, true},
	{0x1FAC0, true},
	{0x1FAC3, true},
	{0x1FAC6, true},
	{0x1FAC7, true},
	{0x1FAC8, true},
	{0x1FAC9, true},
	{0x1FACD, true},
	{0x1FACE, true},
	{0x1FAD0, true},
	{0x1FAD7, true},
	{0x1FADA, true},
	{0x1FADC, true},
	{0x1FADD, true},
	{0x1FADF, true},
	{0x1FAE0, true},
	{0x1FAE8, true},
	{0x1FAE9, true},
	{0x1FAEA, true},
	{0x1FAEB, true},
	{0x1FAEF, true},
	{0x1FAF0, true},
	{0x1FAF7, true},
	{0x1FAF9, true},
	{0x1FC00, true},
}

func TestIsExtendedPictographic(t *testing.T) {
	for _, test := range extendedPictoTests {
		tu.Assert(t, IsExtendedPictographic(test.r) == test.isEmoji)
	}
}

var eastAsianWidthTests = []struct {
	r  rune
	is bool
}{
	{0x32, false},
	{0x20a8, false},
	{0xfe53, false},
	{0x1100, true},
	{0x20a9, true},
	{0x231b, true},
	{0x232a, true},
	{0x23ea, true},
	{0x23f0, true},
	{0x25fd, true},
	{0x2614, true},
	{0x2630, true},
	{0x2648, true},
	{0x267f, true},
	{0x268b, true},
	{0x2693, true},
	{0x26aa, true},
	{0x26bd, true},
	{0x26c4, true},
	{0x26ce, true},
	{0x26ea, true},
	{0x26f3, true},
	{0x26fa, true},
	{0x2705, true},
	{0x270b, true},
	{0x274c, true},
	{0x2753, true},
	{0x2757, true},
	{0x2796, true},
	{0x27b0, true},
	{0x2b1b, true},
	{0x2b50, true},
	{0x2e80, true},
	{0x2e9b, true},
	{0x2f00, true},
	{0x2ff0, true},
	{0x3041, true},
	{0x3099, true},
	{0x3105, true},
	{0x3131, true},
	{0x3190, true},
	{0x31ef, true},
	{0x3220, true},
	{0x3250, true},
	{0xa490, true},
	{0xa960, true},
	{0xac00, true},
	{0xf900, true},
	{0xfe10, true},
	{0xfe30, true},
	{0xfe54, true},
	{0xfe68, true},
	{0xff01, true},
	{0xffc2, true},
	{0xffca, true},
	{0xffd2, true},
	{0xffda, true},
	{0xffe0, true},
	{0xffe8, true},
	{0x16fe0, true},
	{0x16ff0, true},
	{0x17000, true},
	{0x18cff, true},
	{0x18d80, true},
	{0x1aff0, true},
	{0x1aff5, true},
	{0x1affd, true},
	{0x1b000, true},
	{0x1b132, true},
	{0x1b151, true},
	{0x1b155, true},
	{0x1b165, true},
	{0x1b170, true},
	{0x1d300, true},
	{0x1d360, true},
	{0x1f004, true},
	{0x1f18e, true},
	{0x1f192, true},
	{0x1f200, true},
	{0x1f210, true},
	{0x1f240, true},
	{0x1f250, true},
	{0x1f260, true},
	{0x1f300, true},
	{0x1f32d, true},
	{0x1f337, true},
	{0x1f37e, true},
	{0x1f3a0, true},
	{0x1f3cf, true},
	{0x1f3e0, true},
	{0x1f3f4, true},
	{0x1f3f9, true},
	{0x1f440, true},
	{0x1f443, true},
	{0x1f4ff, true},
	{0x1f54b, true},
	{0x1f550, true},
	{0x1f57a, true},
	{0x1f596, true},
	{0x1f5fb, true},
	{0x1f680, true},
	{0x1f6cc, true},
	{0x1f6d1, true},
	{0x1f6d5, true},
	{0x1f6dc, true},
	{0x1f6eb, true},
	{0x1f6f4, true},
	{0x1f7e0, true},
	{0x1f7f0, true},
	{0x1f90d, true},
	{0x1f93c, true},
	{0x1f947, true},
	{0x1fa70, true},
	{0x1fa80, true},
	{0x1fa8e, true},
	{0x1fac8, true},
	{0x1face, true},
	{0x1fadf, true},
	{0x1faef, true},
	{0x20000, true},
	{0x30000, true},
}

func TestIsLargeAsianWidth(t *testing.T) {
	for _, test := range eastAsianWidthTests {
		tu.Assert(t, IsLargeEastAsian(test.r) == test.is)
	}
}

var indicConjunctBreakTests = []struct {
	r     rune
	class IndicConjunctBreak
}{
	{-1, 0},
	{1, 0},
	{'a', 0},
	{'9', 0},
	{0x100, 0},
	{0x1a60, ICBLinker},
	{0x1bab, ICBLinker},
	{0xaaf6, ICBLinker},
	{0x10a3f, ICBLinker},
	{0xaae1, ICBConsonant},
	{0xabc0, ICBConsonant},
	{0x10a00, ICBConsonant},
	{0x10a11, ICBConsonant},
	{0x10a15, ICBConsonant},
	{0x10a19, ICBConsonant},
	{0xfe01, ICBExtend},
	{0xfe20, ICBExtend},
	{0xff9e, ICBExtend},
	{0x101fd, ICBExtend},
	{0x10376, ICBExtend},
	{0x10a01, ICBExtend},
	{0x10a05, ICBExtend},
	{0x10a0c, ICBExtend},
	{0x10a38, ICBExtend},

	{0x1a50, 0b10},
	{0x1a51, 0b10},
	{0x1a52, 0b10},
	{0x1a53, 0b10},
	{0x1a54, 0b10},
	{0x1a55, 0b0},
	{0x1a56, 0b100},
	{0x1a57, 0b0},
	{0x1a58, 0b100},
	{0x1a59, 0b100},
	{0x1a5a, 0b100},
	{0x1a5b, 0b100},
	{0x1a5c, 0b100},
	{0x1a5d, 0b100},
	{0x1a5e, 0b100},
	{0x1a5f, 0b0},
	{0x1a60, 0b1},
	{0x1a61, 0b0},
	{0x1a62, 0b100},
	{0x1a63, 0b0},
	{0x1a64, 0b0},
	{0x1a65, 0b100},
	{0x1a66, 0b100},
	{0x1a67, 0b100},
	{0x1a68, 0b100},
	{0x1a69, 0b100},
	{0x1a6a, 0b100},
	{0x1a6b, 0b100},
	{0x1a6c, 0b100},
	{0x1a6d, 0b0},
	{0x1a6e, 0b0},
	{0x1a6f, 0b0},
	{0x1a70, 0b0},
	{0x1a71, 0b0},
	{0x1a72, 0b0},
	{0x1a73, 0b100},
	{0x1a74, 0b100},
	{0x1a75, 0b100},
	{0x1a76, 0b100},
	{0x1a77, 0b100},
	{0x1a78, 0b100},
	{0x1a79, 0b100},
	{0x1a7a, 0b100},
	{0x1a7b, 0b100},
	{0x1a7c, 0b100},
	{0x1a7d, 0b0},
	{0x1a7e, 0b0},
	{0x1a7f, 0b100},
	{0x1a80, 0b0},
	{0x1a81, 0b0},
	{0x1a82, 0b0},
	{0x1a83, 0b0},
	{0x1a84, 0b0},
	{0x1a85, 0b0},
	{0x1a86, 0b0},
	{0x1a87, 0b0},
	{0x1a88, 0b0},
	{0x1a89, 0b0},
	{0x1a8a, 0b0},
	{0x1a8b, 0b0},
	{0x1a8c, 0b0},
	{0x1a8d, 0b0},
	{0x1a8e, 0b0},
	{0x1a8f, 0b0},
	{0x1a90, 0b0},
	{0x1a91, 0b0},
	{0x1a92, 0b0},
	{0x1a93, 0b0},
	{0x1a94, 0b0},
	{0x1a95, 0b0},
	{0x1a96, 0b0},
	{0x1a97, 0b0},
	{0x1a98, 0b0},
	{0x1a99, 0b0},
	{0x1a9a, 0b0},
	{0x1a9b, 0b0},
	{0x1a9c, 0b0},
	{0x1a9d, 0b0},
	{0x1a9e, 0b0},
	{0x1a9f, 0b0},
	{0x1aa0, 0b0},
	{0x1aa1, 0b0},
	{0x1aa2, 0b0},
	{0x1aa3, 0b0},
	{0x1aa4, 0b0},
	{0x1aa5, 0b0},
	{0x1aa6, 0b0},
	{0x1aa7, 0b0},
	{0x1aa8, 0b0},
	{0x1aa9, 0b0},
	{0x1aaa, 0b0},
	{0x1aab, 0b0},
	{0x1aac, 0b0},
	{0x1aad, 0b0},
	{0x1aae, 0b0},
	{0x1aaf, 0b0},
	{0x1ab0, 0b100},
	{0x1ab1, 0b100},
	{0x1ab2, 0b100},
	{0x1ab3, 0b100},
	{0x1ab4, 0b100},
	{0x1ab5, 0b100},
	{0x1ab6, 0b100},
	{0x1ab7, 0b100},
	{0x1ab8, 0b100},
	{0x1ab9, 0b100},
	{0x1aba, 0b100},
	{0x1abb, 0b100},
	{0x1abc, 0b100},
	{0x1abd, 0b100},
	{0x1abe, 0b100},
	{0x1abf, 0b100},
	{0x1ac0, 0b100},
	{0x1ac1, 0b100},
	{0x1ac2, 0b100},
	{0x1ac3, 0b100},
	{0x1ac4, 0b100},
	{0x1ac5, 0b100},
	{0x1ac6, 0b100},
	{0x1ac7, 0b100},
	{0x1ac8, 0b100},
	{0x1ac9, 0b100},
	{0x1aca, 0b100},
	{0x1acb, 0b100},
	{0x1acc, 0b100},
	{0x1acd, 0b100},
	{0x1ace, 0b100},
	{0x1acf, 0b100},
	{0x1ad0, 0b100},
	{0x1ad1, 0b100},
	{0x1ad2, 0b100},
	{0x1ad3, 0b100},
	{0x1ad4, 0b100},
	{0x1ad5, 0b100},
	{0x1ad6, 0b100},
	{0x1ad7, 0b100},
	{0x1ad8, 0b100},
	{0x1ad9, 0b100},
	{0x1ada, 0b100},
	{0x1adb, 0b100},
	{0x1adc, 0b100},
	{0x1add, 0b100},
	{0x1ade, 0b0},
	{0x1adf, 0b0},
	{0x1ae0, 0b100},
	{0x1ae1, 0b100},
	{0x1ae2, 0b100},
	{0x1ae3, 0b100},
	{0x1ae4, 0b100},
	{0x1ae5, 0b100},
	{0x1ae6, 0b100},
	{0x1ae7, 0b100},
	{0x1ae8, 0b100},
	{0x1ae9, 0b100},
	{0x1aea, 0b100},
	{0x1aeb, 0b100},
	{0x1aec, 0b0},
	{0x1aed, 0b0},
	{0x1aee, 0b0},
	{0x1aef, 0b0},
}

func TestIndicConjunctBreak(t *testing.T) {
	for _, test := range indicConjunctBreakTests {
		tu.Assert(t, LookupIndicConjunctBreak(test.r) == test.class)
	}
}

// See https://www.unicode.org/reports/tr14/#DescriptionOfProperties
var lineBreakTests = []struct {
	args rune
	want LineBreak
}{
	{-1, 0},

	// AI: Ambiguous (Alphabetic or Ideograph)
	{'\u24EA', LB_AI},
	{'\u2780', LB_AI},
	// AL: Ordinary Alphabetic and Symbol Characters (XP)
	{'\u0600', LB_NU},     //   ARABIC NUMBER SIGN
	{'\u06DD', LB_NU},     //  	ARABIC END OF AYAH
	{'\u070F', LB_AL},     //  	SYRIAC ABBREVIATION MARK
	{'\u2061', LB_AL},     //   	FUNCTION APPLICATION
	{'\U000110BD', LB_NU}, //  	KAITHI NUMBER SIGN
	// BA: LB_ After (A)
	{'\u1680', LB_BA},     // OGHAM SPACE MARK
	{'\u2000', LB_BA},     // EN QUAD
	{'\u2001', LB_BA},     // EM QUAD
	{'\u2002', LB_BA},     // EN SPACE
	{'\u2003', LB_BA},     // EM SPACE
	{'\u2004', LB_BA},     // THREE-PER-EM SPACE
	{'\u2005', LB_BA},     // FOUR-PER-EM SPACE
	{'\u2006', LB_BA},     // SIX-PER-EM SPACE
	{'\u2008', LB_BA},     // PUNCTUATION SPACE
	{'\u2009', LB_BA},     // THIN SPACE
	{'\u200A', LB_BA},     // HAIR SPACE
	{'\u205F', LB_BA},     // MEDIUM MATHEMATICAL SPACE
	{'\u3000', LB_BA},     // IDEOGRAPHIC SPACE
	{'\u0009', LB_BA},     //	TAB
	{'\u00AD', LB_BA},     //	SOFT HYPHEN (SHY)
	{'\u058A', LB_HH},     //	ARMENIAN HYPHEN
	{'\u2010', LB_HH},     //	HYPHEN
	{'\u2012', LB_HH},     //	FIGURE DASH
	{'\u2013', LB_HH},     //	EN DASH
	{'\u05BE', LB_HH},     //	HEBREW PUNCTUATION MAQAF
	{'\u0F0B', LB_BA},     //	TIBETAN MARK INTERSYLLABIC TSHEG
	{'\u1361', LB_BA},     //	ETHIOPIC WORDSPACE
	{'\u17D8', LB_BA},     //	KHMER SIGN BEYYAL
	{'\u17DA', LB_BA},     //	KHMER SIGN KOOMUUT
	{'\u2027', LB_BA},     //	HYPHENATION POINT
	{'\u007C', LB_BA},     //	VERTICAL LINE
	{'\u16EB', LB_BA},     //	RUNIC SINGLE PUNCTUATION
	{'\u16EC', LB_BA},     //	RUNIC MULTIPLE PUNCTUATION
	{'\u16ED', LB_BA},     //	RUNIC CROSS PUNCTUATION
	{'\u2056', LB_BA},     //	THREE DOT PUNCTUATION
	{'\u2058', LB_BA},     //	FOUR DOT PUNCTUATION
	{'\u2059', LB_BA},     //	FIVE DOT PUNCTUATION
	{'\u205A', LB_BA},     //	TWO DOT PUNCTUATION
	{'\u205B', LB_BA},     //	FOUR DOT MARK
	{'\u205D', LB_BA},     //	TRICOLON
	{'\u205E', LB_BA},     //	VERTICAL FOUR DOTS
	{'\u2E19', LB_BA},     //	PALM BRANCH
	{'\u2E2A', LB_BA},     //	TWO DOTS OVER ONE DOT PUNCTUATION
	{'\u2E2B', LB_BA},     //	ONE DOT OVER TWO DOTS PUNCTUATION
	{'\u2E2C', LB_BA},     //	SQUARED FOUR DOT PUNCTUATION
	{'\u2E2D', LB_BA},     //	FIVE DOT MARK
	{'\u2E30', LB_BA},     //	RING POINT
	{'\U00010100', LB_BA}, //	AEGEAN WORD SEPARATOR LINE
	{'\U00010101', LB_BA}, //	AEGEAN WORD SEPARATOR DOT
	{'\U00010102', LB_BA}, //	AEGEAN CHECK MARK
	{'\U0001039F', LB_BA}, //	UGARITIC WORD DIVIDER
	{'\U000103D0', LB_BA}, //	OLD PERSIAN WORD DIVIDER
	{'\U0001091F', LB_BA}, //	PHOENICIAN WORD SEPARATOR
	{'\U00012470', LB_BA}, //	CUNEIFORM PUNCTUATION SIGN OLD ASSYRIAN WORD DIVIDER
	{'\u0964', LB_BA},     //	DEVANAGARI DANDA
	{'\u0965', LB_BA},     //	DEVANAGARI DOUBLE DANDA
	{'\u0E5A', LB_BA},     //	THAI CHARACTER ANGKHANKHU
	{'\u0E5B', LB_BA},     //	THAI CHARACTER KHOMUT
	{'\u104A', LB_BA},     //	MYANMAR SIGN LITTLE SECTION
	{'\u104B', LB_BA},     //	MYANMAR SIGN SECTION
	{'\u1735', LB_BA},     //	PHILIPPINE SINGLE PUNCTUATION
	{'\u1736', LB_BA},     //	PHILIPPINE DOUBLE PUNCTUATION
	{'\u17D4', LB_BA},     //	KHMER SIGN KHAN
	{'\u17D5', LB_BA},     //	KHMER SIGN BARIYOOSAN
	{'\u1B5E', LB_BA},     //	BALINESE CARIK SIKI
	{'\u1B5F', LB_BA},     //	BALINESE CARIK PAREREN
	{'\uA8CE', LB_BA},     //	SAURASHTRA DANDA
	{'\uA8CF', LB_BA},     //	SAURASHTRA DOUBLE DANDA
	{'\uAA5D', LB_BA},     //	CHAM PUNCTUATION DANDA
	{'\uAA5E', LB_BA},     //	CHAM PUNCTUATION DOUBLE DANDA
	{'\uAA5F', LB_BA},     //	CHAM PUNCTUATION TRIPLE DANDA
	{'\U00010A56', LB_BA}, //	KHAROSHTHI PUNCTUATION DANDA
	{'\U00010A57', LB_BA}, //	KHAROSHTHI PUNCTUATION DOUBLE DANDA
	{'\u0F34', LB_BA},     //	TIBETAN MARK BSDUS RTAGS
	{'\u0F7F', LB_BA},     //	TIBETAN SIGN RNAM BCAD
	{'\u0F85', LB_BA},     //	TIBETAN MARK PALUTA
	{'\u0FBE', LB_BA},     //	TIBETAN KU RU KHA
	{'\u0FBF', LB_BA},     //	TIBETAN KU RU KHA BZHI MIG CAN
	{'\u0FD2', LB_BA},     //	TIBETAN MARK NYIS TSHEG
	{'\u1804', LB_BA},     //	MONGOLIAN COLON
	{'\u1805', LB_BA},     //	MONGOLIAN FOUR DOTS
	{'\u1B5A', LB_BA},     //	BALINESE PANTI
	{'\u1B5B', LB_BA},     //	BALINESE PAMADA
	{'\u1B5D', LB_BA},     //	BALINESE CARIK PAMUNGKAH
	{'\u1B60', LB_BA},     //	BALINESE PAMENENG
	{'\u1C3B', LB_BA},     //	LEPCHA PUNCTUATION TA-ROL
	{'\u1C3C', LB_BA},     //	LEPCHA PUNCTUATION NYET THYOOM TA-ROL
	{'\u1C3D', LB_BA},     //	LEPCHA PUNCTUATION CER-WA
	{'\u1C3E', LB_BA},     //	LEPCHA PUNCTUATION TSHOOK CER-WA
	{'\u1C3F', LB_BA},     //	LEPCHA PUNCTUATION TSHOOK
	{'\u1C7E', LB_BA},     //	OL CHIKI PUNCTUATION MUCAAD
	{'\u1C7F', LB_BA},     //	OL CHIKI PUNCTUATION DOUBLE MUCAAD
	{'\u2CFA', LB_BA},     //	COPTIC OLD NUBIAN DIRECT QUESTION MARK
	{'\u2CFB', LB_BA},     //	COPTIC OLD NUBIAN INDIRECT QUESTION MARK
	{'\u2CFC', LB_BA},     //	COPTIC OLD NUBIAN VERSE DIVIDER
	{'\u2CFF', LB_BA},     //	COPTIC MORPHOLOGICAL DIVIDER
	{'\u2E0E', LB_BA},     // EDITORIAL CORONIS
	{'\u2E17', LB_HH},     //	DOUBLE OBLIQUE HYPHEN
	{'\uA60D', LB_BA},     //	VAI COMMA
	{'\uA60F', LB_BA},     //	VAI QUESTION MARK
	{'\uA92E', LB_BA},     //	KAYAH LI SIGN CWI
	{'\uA92F', LB_BA},     //	KAYAH LI SIGN SHYA
	{'\U00010A50', LB_BA}, //	KHAROSHTHI PUNCTUATION DOT
	{'\U00010A51', LB_BA}, //	KHAROSHTHI PUNCTUATION SMALL CIRCLE
	{'\U00010A52', LB_BA}, //	KHAROSHTHI PUNCTUATION CIRCLE
	{'\U00010A53', LB_BA}, //	KHAROSHTHI PUNCTUATION CRESCENT BAR
	{'\U00010A54', LB_BA}, //	KHAROSHTHI PUNCTUATION MANGALAM
	{'\U00010A55', LB_BA}, //	KHAROSHTHI PUNCTUATION LOTUS
	// BB: LB_ Before (B)
	{'\u00B4', LB_BB}, //	ACUTE ACCENT
	{'\u1FFD', LB_BB}, //	GREEK OXIA
	{'\u02DF', LB_BB}, //	MODIFIER LETTER CROSS ACCENT
	{'\u02C8', LB_BB}, //	MODIFIER LETTER VERTICAL LINE
	{'\u02CC', LB_BB}, //	MODIFIER LETTER LOW VERTICAL LINE
	{'\u0F01', LB_BB}, //	TIBETAN MARK GTER YIG MGO TRUNCATED A
	{'\u0F02', LB_BB}, //	TIBETAN MARK GTER YIG MGO -UM RNAM BCAD MA
	{'\u0F03', LB_BB}, //	TIBETAN MARK GTER YIG MGO -UM GTER TSHEG MA
	{'\u0F04', LB_BB}, //	TIBETAN MARK INITIAL YIG MGO MDUN MA
	{'\u0F06', LB_BB}, //	TIBETAN MARK CARET YIG MGO PHUR SHAD MA
	{'\u0F07', LB_BB}, //	TIBETAN MARK YIG MGO TSHEG SHAD MA
	{'\u0F09', LB_BB}, //	TIBETAN MARK BSKUR YIG MGO
	{'\u0F0A', LB_BB}, //	TIBETAN MARK BKA- SHOG YIG MGO
	{'\u0FD0', LB_BB}, //	TIBETAN MARK BSKA- SHOG GI MGO RGYAN
	{'\u0FD1', LB_BB}, //	TIBETAN MARK MNYAM YIG GI MGO RGYAN
	{'\u0FD3', LB_BB}, //	TIBETAN MARK INITIAL BRDA RNYING YIG MGO MDUN MA
	{'\uA874', LB_BB}, //	PHAGS-PA SINGLE HEAD MARK
	{'\uA875', LB_BB}, //	PHAGS-PA DOUBLE HEAD MARK
	{'\u1806', LB_BB}, //	MONGOLIAN TODO SOFT HYPHEN

	//	B2: LB_ Opportunity Before and After (B/A/XP)
	{'\u2014', LB_B2}, //	EM DASH

	// BK: Mandatory LB_ (A) (Non-tailorable)

	{'\u000C', LB_BK}, //	FORM FEED (FF)
	{'\u000B', LB_BK}, //	LINE TABULATION (VT)
	{'\u2028', LB_BK}, //	LINE SEPARATOR
	{'\u2029', LB_BK}, //	PARAGRAPH SEPARATOR

	// CB: Contingent LB_ Opportunity (B/A)

	{'\uFFFC', LB_CB}, //	OBJECT REPLACEMENT CHARACTER

	// CJ: Conditional Japanese Starter

	{'\u3041', LB_CJ}, // 	Small hiragana
	{'\u30A1', LB_CJ}, // 	Small katakana
	{'\u30FC', LB_CJ}, //	KATAKANA-HIRAGANA PROLONGED SOUND MARK
	{'\uFF67', LB_CJ}, // 	Halfwidth variants

	// CL: Close Punctuation (XB)

	{'\u3001', LB_CL}, // 	IDEOGRAPHIC COMMA..IDEOGRAPHIC FULL STOP
	{'\uFE11', LB_CL}, //	PRESENTATION FORM FOR VERTICAL IDEOGRAPHIC COMMA
	{'\uFE12', LB_CL}, //	PRESENTATION FORM FOR VERTICAL IDEOGRAPHIC FULL STOP
	{'\uFE50', LB_CL}, //	SMALL COMMA
	{'\uFE52', LB_CL}, //	SMALL FULL STOP
	{'\uFF0C', LB_CL}, //	FULLWIDTH COMMA
	{'\uFF0E', LB_CL}, //	FULLWIDTH FULL STOP
	{'\uFF61', LB_CL}, //	HALFWIDTH IDEOGRAPHIC FULL STOP
	{'\uFF64', LB_CL}, //	HALFWIDTH IDEOGRAPHIC COMMA

	// CP: Closing Parenthesis (XB)

	{'\u0029', LB_CP}, //	RIGHT PARENTHESIS
	{'\u005D', LB_CP}, //	RIGHT SQUARE BRACKET

	// CR: Carriage Return (A) (Non-tailorable)
	{'\u000D', LB_CR}, //	CARRIAGE RETURN (CR)

	// EB: Emoji Base (B/A)
	{'\U0001F466', LB_EB}, //	BOY
	{'\U0001F478', LB_EB}, //	PRINCESS
	{'\U0001F6B4', LB_EB}, //	BICYCLIST

	// EM: Emoji Modifier (A)

	{'\U0001F3FB', LB_EM}, // EMOJI MODIFIER FITZPATRICK TYPE-1-2
	{'\U0001F3FF', LB_EM}, // EMOJI MODIFIER FITZPATRICK TYPE-6

	// EX: Exclamation/Interrogation (XB)

	{'\u0021', LB_EX}, //	EXCLAMATION MARK
	{'\u003F', LB_EX}, //	QUESTION MARK
	{'\u05C6', LB_EX}, //	HEBREW PUNCTUATION NUN HAFUKHA
	{'\u061B', LB_EX}, //	ARABIC SEMICOLON
	{'\u061E', LB_EX}, //	ARABIC TRIPLE DOT PUNCTUATION MARK
	{'\u061F', LB_EX}, //	ARABIC QUESTION MARK
	{'\u06D4', LB_EX}, //	ARABIC FULL STOP
	{'\u07F9', LB_EX}, //	NKO EXCLAMATION MARK
	{'\u0F0D', LB_EX}, //	TIBETAN MARK SHAD
	{'\uFF01', LB_EX}, //	FULLWIDTH EXCLAMATION MARK
	{'\uFF1F', LB_EX}, //	FULLWIDTH QUESTION MARK

	// GL: Non-breaking (“Glue”) (XB/XA) (Non-tailorable)

	{'\u00A0', LB_GL}, //	NO-BREAK SPACE (NBSP)
	{'\u202F', LB_GL}, //	NARROW NO-BREAK SPACE (NNBSP)
	{'\u180E', LB_GL}, //	MONGOLIAN VOWEL SEPARATOR (MVS)
	{'\u034F', LB_CM}, //	COMBINING GRAPHEME JOINER
	{'\u2007', LB_GL}, //	FIGURE SPACE
	{'\u2011', LB_GL}, //	NON-BREAKING HYPHEN
	{'\u0F08', LB_GL}, //	TIBETAN MARK SBRUL SHAD
	{'\u0F0C', LB_GL}, //	TIBETAN MARK DELIMITER TSHEG BSTAR
	{'\u0F12', LB_GL}, //	TIBETAN MARK RGYA GRAM SHAD
	{'\u035C', LB_GL}, // COMBINING DOUBLE BREVE BELOW
	{'\u0362', LB_GL}, //	 COMBINING DOUBLE RIGHTWARDS ARROW BELOW

	// HY: Hyphen (XA)
	{'\u002D', LB_HY}, //	HYPHEN-MINUS

	// ID: Ideographic (B/A)

	{'\u2E80', LB_ID},     // 	CJK, Kangxi Radicals, Ideographic Description Symbols
	{'\u30A2', LB_ID},     // 	Katakana (except small characters)
	{'\u3400', LB_ID},     // 	CJK Unified Ideographs Extension A
	{'\u4E00', LB_ID},     // 	CJK Unified Ideographs
	{'\uF900', LB_ID},     // 	CJK Compatibility Ideographs
	{'\u3400', LB_ID},     // 	CJK Unified Ideographs Extension A
	{'\u4E00', LB_ID},     // 	CJK Unified Ideographs
	{'\uF900', LB_ID},     // 	CJK Compatibility Ideographs
	{'\U00020000', LB_ID}, // 	Plane 2
	{'\U00030000', LB_ID}, // 	Plane 3
	{'\U0001F000', LB_ID}, // 	Plane 1 range

	// IN: Inseparable Characters (XP)

	{'\u2024', LB_IN}, //	ONE DOT LEADER
	{'\u2025', LB_IN}, //	TWO DOT LEADER
	{'\u2026', LB_IN}, //	HORIZONTAL ELLIPSIS
	{'\uFE19', LB_IN}, //	PRESENTATION FORM FOR VERTICAL HORIZONTAL ELLIPSIS

	// IS: Infix Numeric Separator (XB)

	{'\u002C', LB_IS}, //	COMMA
	{'\u002E', LB_IS}, //	FULL STOP
	{'\u003A', LB_IS}, //	COLON
	{'\u003B', LB_IS}, //	SEMICOLON
	{'\u037E', LB_IS}, //	GREEK QUESTION MARK (canonically equivalent to 003B)
	{'\u0589', LB_IS}, //	ARMENIAN FULL STOP
	{'\u060C', LB_IS}, //	ARABIC COMMA
	{'\u060D', LB_IS}, //	ARABIC DATE SEPARATOR
	{'\u07F8', LB_IS}, //	NKO COMMA
	{'\u2044', LB_IS}, //	FRACTION SLASH
	{'\uFE10', LB_CL}, //	PRESENTATION FORM FOR VERTICAL COMMA
	{'\uFE13', LB_NS}, //	PRESENTATION FORM FOR VERTICAL COLON
	{'\uFE14', LB_NS}, //	PRESENTATION FORM FOR VERTICAL SEMICOLON

	// LF: Line Feed (A) (Non-tailorable)
	{'\u000A', LB_LF}, //	LINE FEED (LF)

	// NL: Next Line (A) (Non-tailorable)
	{'\u0085', LB_NL}, //	NEXT LINE (NEL)

	// NS: Nonstarters (XB)

	{'\u17D6', LB_NS}, //	KHMER SIGN CAMNUC PII KUUH
	{'\u203C', LB_NS}, //	DOUBLE EXCLAMATION MARK
	{'\u203D', LB_NS}, //	INTERROBANG
	{'\u2047', LB_NS}, //	DOUBLE QUESTION MARK
	{'\u2048', LB_NS}, //	QUESTION EXCLAMATION MARK
	{'\u2049', LB_NS}, //	EXCLAMATION QUESTION MARK
	{'\u3005', LB_NS}, //	IDEOGRAPHIC ITERATION MARK
	{'\u301C', LB_NS}, //	WAVE DASH
	{'\u303C', LB_NS}, //	MASU MARK
	{'\u303B', LB_NS}, //	VERTICAL IDEOGRAPHIC ITERATION MARK
	{'\u309B', LB_NS}, // 	KATAKANA-HIRAGANA VOICED SOUND MARK..HIRAGANA VOICED ITERATION MARK
	{'\u30A0', LB_NS}, //	KATAKANA-HIRAGANA DOUBLE HYPHEN
	{'\u30FB', LB_NS}, //	KATAKANA MIDDLE DOT
	{'\u30FD', LB_NS}, // 	KATAKANA ITERATION MARK..KATAKANA VOICED ITERATION MARK
	{'\uFE54', LB_NS}, // 	SMALL SEMICOLON..SMALL COLON
	{'\uFF1A', LB_NS}, // 	FULLWIDTH COLON.. FULLWIDTH SEMICOLON
	{'\uFF65', LB_NS}, //	HALFWIDTH KATAKANA MIDDLE DOT
	{'\uFF9E', LB_NS}, // 	HALFWIDTH KATAKANA VOICED SOUND MARK..HALFWIDTH KATAKANA SEMI-VOICED SOUND MARK

	// NU: Numeric (XP)

	{'\u066B', LB_NU}, //	ARABIC DECIMAL SEPARATOR
	{'\u066C', LB_NU}, //	ARABIC THOUSANDS SEPARATOR

	// OP: Open Punctuation (XA)

	{'\u00A1', LB_OP}, //	INVERTED EXCLAMATION MARK
	{'\u00BF', LB_OP}, //	INVERTED QUESTION MARK
	{'\u2E18', LB_OP}, //	INVERTED INTERROBANG

	// PO: Postfix Numeric (XB)

	{'\u0025', LB_PO}, //	PERCENT SIGN
	{'\u00A2', LB_PO}, //	CENT SIGN
	{'\u00B0', LB_PO}, //	DEGREE SIGN
	{'\u060B', LB_PO}, //	AFGHANI SIGN
	{'\u066A', LB_PO}, //	ARABIC PERCENT SIGN
	{'\u2030', LB_PO}, //	PER MILLE SIGN
	{'\u2031', LB_PO}, //	PER TEN THOUSAND SIGN
	{'\u2032', LB_PO}, // 	PRIME
	{'\u20A7', LB_PO}, //	PESETA SIGN
	{'\u2103', LB_PO}, //	DEGREE CELSIUS
	{'\u2109', LB_PO}, //	DEGREE FAHRENHEIT
	{'\uFDFC', LB_PO}, //	RIAL SIGN
	{'\uFE6A', LB_PO}, //	SMALL PERCENT SIGN
	{'\uFF05', LB_PO}, //	FULLWIDTH PERCENT SIGN
	{'\uFFE0', LB_PO}, //	FULLWIDTH CENT SIGN

	// PR: Prefix Numeric (XA)

	{'\u002B', LB_PR}, //	PLUS SIGN
	{'\u005C', LB_PR}, //	REVERSE SOLIDUS
	{'\u00B1', LB_PR}, //	PLUS-MINUS SIGN
	{'\u2116', LB_PR}, //	NUMERO SIGN
	{'\u2212', LB_PR}, //	MINUS SIGN
	{'\u2213', LB_PR}, //	MINUS-OR-PLUS SIGN

	// QU: Quotation (XB/XA)

	{'\u0022', LB_QU}, //	QUOTATION MARK
	{'\u0027', LB_QU}, //	APOSTROPHE
	{'\u275B', LB_QU}, //	HEAVY SINGLE TURNED COMMA QUOTATION MARK ORNAMENT
	{'\u275C', LB_QU}, //	HEAVY SINGLE COMMA QUOTATION MARK ORNAMENT
	{'\u275D', LB_QU}, //	HEAVY DOUBLE TURNED COMMA QUOTATION MARK ORNAMENT
	{'\u275E', LB_QU}, //	HEAVY DOUBLE COMMA QUOTATION MARK ORNAMENT
	{'\u2E00', LB_QU}, //	RIGHT ANGLE SUBSTITUTION MARKER
	{'\u2E06', LB_QU}, //	RAISED INTERPOLATION MARKER
	{'\u2E0B', LB_QU}, //	RAISED SQUARE
	// RI: Regional Indicator (B/A/XP)

	{'\U0001F1E6', LB_RI}, // REGIONAL INDICATOR SYMBOL LETTER A
	{'\U0001F1FF', LB_RI}, // REGIONAL INDICATOR SYMBOL LETTER Z

	// SA: Complex-Context Dependent (South East Asian) (P)

	{'\u1000', LB_SA},     //	Myanmar
	{'\u1780', LB_SA},     //	Khmer
	{'\u1950', LB_SA},     //	Tai Le
	{'\u1980', LB_SA},     //	New Tai Lue
	{'\u1A20', LB_SA},     //	Tai Tham
	{'\uA9E0', LB_SA},     //	Myanmar Extended-B
	{'\uAA60', LB_SA},     //	Myanmar Extended-A
	{'\uAA80', LB_SA},     //	Tai Viet
	{'\U00011700', LB_SA}, // 	Ahom

	// SP: Space (A) (Non-tailorable)

	{'\u0020', LB_SP}, //	SPACE (SP)

	// SY: Symbols Allowing LB_ After (A)

	{'\u002F', LB_SY}, //	SOLIDUS

	// WJ: Word Joiner (XB/XA) (Non-tailorable)

	{'\u2060', LB_WJ}, //	WORD JOINER (WJ)
	{'\uFEFF', LB_WJ}, //	ZERO WIDTH NO-BREAK SPACE (ZWNBSP)

	// ZW: Zero Width Space (A) (Non-tailorable)
	{'\u200B', LB_ZW}, //	ZERO WIDTH SPACE (ZWSP)

	// ZWJ: Zero Width Joiner (XA/XB) (Non-tailorable)
	{'\u200D', LB_ZWJ}, //	ZERO WIDTH JOINER (ZWJ)
}

func TestLookupLineBreak(t *testing.T) {
	for _, tt := range lineBreakTests {
		tu.Assert(t, LookupLineBreak(tt.args) == tt.want)
	}
}

// See https://www.unicode.org/reports/tr29/#Grapheme_Cluster_Break_Property_Values
var graphemeBreakTests = []struct {
	args rune
	want GraphemeBreak
}{
	{-1, 0},
	{'a', 0},
	// CR
	{'\u000D', GB_CR}, // CARRIAGE RETURN (CR)
	// Control
	{0xe01f0, GB_Control},
	// Extend
	{'\u200C', GB_Extend}, // ZERO WIDTH NON-JOINER
	// LF
	{'\u000A', GB_LF}, // LINE FEED (LF)
	// ZWJ
	{'\u200D', GB_ZWJ}, // ZERO WIDTH JOINER
	// RI
	{'\U0001F1E6', GB_Regional_Indicator}, // REGIONAL INDICATOR SYMBOL LETTER A
	// SpacingMark
	{'\u0E33', GB_SpacingMark}, // ( ำ ) THAI CHARACTER SARA AM
	{'\u0EB3', GB_SpacingMark}, // ( ຳ ) LAO VOWEL SIGN AM

	// L
	{'\u1100', GB_L}, // ( ᄀ ) HANGUL CHOSEONG KIYEOK
	{'\u115F', GB_L}, // ( ᅟ ) HANGUL CHOSEONG FILLER
	{'\uA960', GB_L}, // ( ꥠ ) HANGUL CHOSEONG TIKEUT-MIEUM
	{'\uA97C', GB_L}, // ( ꥼ ) HANGUL CHOSEONG SSANGYEORINHIEUH
	// V
	{'\u1160', GB_V}, // ( ᅠ ) HANGUL JUNGSEONG FILLER
	{'\u11A2', GB_V}, // ( ᆢ ) HANGUL JUNGSEONG SSANGARAEA
	{'\uD7B0', GB_V}, // ( ힰ ) HANGUL JUNGSEONG O-YEO
	{'\uD7C6', GB_V}, // ( ퟆ ) HANGUL JUNGSEONG ARAEA-E
	// T
	{'\u11A8', GB_T}, // ( ᆨ ) HANGUL JONGSEONG KIYEOK
	{'\u11F9', GB_T}, // ( ᇹ ) HANGUL JONGSEONG YEORINHIEUH
	{'\uD7CB', GB_T}, // ( ퟋ ) HANGUL JONGSEONG NIEUN-RIEUL
	{'\uD7FB', GB_T}, // ( ퟻ ) HANGUL JONGSEONG PHIEUPH-THIEUTH
	// LV
	{'\uAC00', GB_LV}, // ( 가 ) HANGUL SYLLABLE GA
	{'\uAC1C', GB_LV}, // ( 개 ) HANGUL SYLLABLE GAE
	{'\uAC38', GB_LV}, // ( 갸 ) HANGUL SYLLABLE GYA
	// LVT
	{'\uAC01', GB_LVT}, // ( 각 ) HANGUL SYLLABLE GAG
	{'\uAC02', GB_LVT}, // ( 갂 ) HANGUL SYLLABLE GAGG
	{'\uAC03', GB_LVT}, // ( 갃 ) HANGUL SYLLABLE GAGS
	{'\uAC04', GB_LVT}, // ( 간 ) HANGUL SYLLABLE GAN
}

func TestLookupGraphemeBreak(t *testing.T) {
	for _, tt := range graphemeBreakTests {
		tu.Assert(t, LookupGraphemeBreak(tt.args) == tt.want)
	}
}

var wordBreakTests = []struct {
	args rune
	want WordBreak
}{
	// these runes have changed from Unicode v15.0 to 15.1
	{0x6DD, WB_Numeric},
	{0x661, WB_Numeric},

	{'\u000D', WB_NewlineCRLF},            // CARRIAGE RETURN (CR)
	{'\u000A', WB_NewlineCRLF},            // LINE FEED (LF)
	{'\u000B', WB_NewlineCRLF},            // LINE TABULATION
	{'\u000C', WB_NewlineCRLF},            // FORM FEED (FF)
	{'\u0085', WB_NewlineCRLF},            // NEXT LINE (NEL)
	{'\u2028', WB_NewlineCRLF},            // LINE SEPARATOR
	{'\u2029', WB_NewlineCRLF},            // PARAGRAPH SEPARATOR
	{'\U0001F1E6', WB_Regional_Indicator}, // REGIONAL INDICATOR SYMBOL LETTER A
	{'\U0001F1FF', WB_Regional_Indicator}, // REGIONAL INDICATOR SYMBOL LETTER Z
	{'\u3031', WB_Katakana},               // ( 〱 ) VERTICAL KANA REPEAT MARK
	{'\u3032', WB_Katakana},               // ( 〲 ) VERTICAL KANA REPEAT WITH VOICED SOUND MARK
	{'\u3033', WB_Katakana},               // ( 〳 ) VERTICAL KANA REPEAT MARK UPPER HALF
	{'\u3034', WB_Katakana},               // ( 〴 ) VERTICAL KANA REPEAT WITH VOICED SOUND MARK UPPER HALF
	{'\u3035', WB_Katakana},               // ( 〵 ) VERTICAL KANA REPEAT MARK LOWER HALF
	{'\u309B', WB_Katakana},               // ( ゛ ) KATAKANA-HIRAGANA VOICED SOUND MARK
	{'\u309C', WB_Katakana},               // ( ゜ ) KATAKANA-HIRAGANA SEMI-VOICED SOUND MARK
	{'\u30A0', WB_Katakana},               // ( ゠ ) KATAKANA-HIRAGANA DOUBLE HYPHEN
	{'\u30FC', WB_Katakana},               // ( ー ) KATAKANA-HIRAGANA PROLONGED SOUND MARK
	{'\uFF70', WB_Katakana},               // ( ｰ ) HALFWIDTH KATAKANA-HIRAGANA PROLONGED SOUND MARK
	{'\u00B8', WB_ALetter},                // ( ¸ ) CEDILLA
	{'\u02C2', WB_ALetter},                // ( ˂ ) MODIFIER LETTER LEFT ARROWHEAD
	{'\u02C5', WB_ALetter},                // ( ˅ ) MODIFIER LETTER DOWN ARROWHEAD
	{'\u02D2', WB_ALetter},                // ( ˒ ) MODIFIER LETTER CENTRED RIGHT HALF RING
	{'\u02D7', WB_ALetter},                // ( ˗ ) MODIFIER LETTER MINUS SIGN
	{'\u02DE', WB_ALetter},                // ( ˞ ) MODIFIER LETTER RHOTIC HOOK
	{'\u02DF', WB_ALetter},                // ( ˟ ) MODIFIER LETTER CROSS ACCENT
	{'\u02E5', WB_ALetter},                // ( ˥ ) MODIFIER LETTER EXTRA-HIGH TONE BAR
	{'\u02EB', WB_ALetter},                // ( ˫ ) MODIFIER LETTER YANG DEPARTING TONE MARK
	{'\u02ED', WB_ALetter},                // ( ˭ ) MODIFIER LETTER UNASPIRATED
	{'\u02EF', WB_ALetter},                // ( ˯ ) MODIFIER LETTER LOW DOWN ARROWHEAD
	{'\u02FF', WB_ALetter},                // ( ˿ ) MODIFIER LETTER LOW LEFT ARROW
	{'\u055A', WB_ALetter},                // ( ՚ ) ARMENIAN APOSTROPHE
	{'\u055B', WB_ALetter},                // ( ՛ ) ARMENIAN EMPHASIS MARK
	{'\u055C', WB_ALetter},                // ( ՜ ) ARMENIAN EXCLAMATION MARK
	{'\u055E', WB_ALetter},                // ( ՞ ) ARMENIAN QUESTION MARK
	{'\u058A', WB_ALetter},                // ( ֊ ) ARMENIAN HYPHEN
	{'\u05F3', WB_ALetter},                // ( ׳ ) HEBREW PUNCTUATION GERESH
	{'\u070F', WB_ALetter},                // ( ܏ ) SYRIAC ABBREVIATION MARK
	{'\uA708', WB_ALetter},                // ( ꜈ ) MODIFIER LETTER EXTRA-HIGH DOTTED TONE BAR
	{'\uA716', WB_ALetter},                // ( ꜖ ) MODIFIER LETTER EXTRA-LOW LEFT-STEM TONE BAR
	{'\uA720', WB_ALetter},                // (꜠ ) MODIFIER LETTER STRESS AND HIGH TONE
	{'\uA721', WB_ALetter},                // (꜡ ) MODIFIER LETTER STRESS AND LOW TONE
	{'\uA789', WB_ALetter},                // (꞉ ) MODIFIER LETTER COLON
	{'\uA78A', WB_ALetter},                // ( ꞊ ) MODIFIER LETTER SHORT EQUALS SIGN
	{'\uAB5B', WB_ALetter},                // ( ꭛ ) MODIFIER BREVE WITH INVERTED BREVE
	{'\u0027', WB_Single_Quote},           // ( ' ) APOSTROPHE
	{'\u0022', WB_Double_Quote},           // ( " ) QUOTATION MARK
	{'\u002E', WB_MidNumLet},              // ( . ) FULL STOP
	{'\u2018', WB_MidNumLet},              // ( ‘ ) LEFT SINGLE QUOTATION MARK
	{'\u2019', WB_MidNumLet},              // ( ’ ) RIGHT SINGLE QUOTATION MARK
	{'\u2024', WB_MidNumLet},              // ( ․ ) ONE DOT LEADER
	{'\uFE52', WB_MidNumLet},              // ( ﹒ ) SMALL FULL STOP
	{'\uFF07', WB_MidNumLet},              // ( ＇ ) FULLWIDTH APOSTROPHE
	{'\uFF0E', WB_MidNumLet},              // ( ． ) FULLWIDTH FULL STOP
	{'\u003A', WB_MidLetter},              //  ( : ) COLON (used in Swedish)
	{'\u00B7', WB_MidLetter},              //  ( · ) MIDDLE DOT
	{'\u0387', WB_MidLetter},              //  ( · ) GREEK ANO TELEIA
	{'\u055F', WB_MidLetter},              //  ( ՟ ) ARMENIAN ABBREVIATION MARK
	{'\u05F4', WB_MidLetter},              //  ( ״ ) HEBREW PUNCTUATION GERSHAYIM
	{'\u2027', WB_MidLetter},              //  ( ‧ ) HYPHENATION POINT
	{'\uFE13', WB_MidLetter},              //  ( ︓ ) PRESENTATION FORM FOR VERTICAL COLON
	{'\uFE55', WB_MidLetter},              //  ( ﹕ ) SMALL COLON
	{'\uFF1A', WB_MidLetter},              //  ( ： ) FULLWIDTH COLON
	{'\u066C', WB_MidNum},                 // ( ٬ ) ARABIC THOUSANDS SEPARATOR
	{'\uFE50', WB_MidNum},                 // ( ﹐ ) SMALL COMMA
	{'\uFE54', WB_MidNum},                 // ( ﹔ ) SMALL SEMICOLON
	{'\uFF0C', WB_MidNum},                 // ( ， ) FULLWIDTH COMMA
	{'\uFF1B', WB_MidNum},                 // ( ； ) FULLWIDTH SEMICOLON
	{'\u202F', WB_ExtendNumLet},           // NARROW NO-BREAK SPACE (NNBSP)
}

func TestLookupWordBreak(t *testing.T) {
	for _, tt := range wordBreakTests {
		tu.Assert(t, LookupWordBreak(tt.args) == tt.want)
	}
}

var isWordTests = []struct {
	args rune
	want bool
}{
	{',', false},
	{'.', false},
	{0xffd2, true},
	{0xffd7, true},
	{0xffd8, false},
	{0xffda, true},
	{0xffdc, true},
	{0xffdd, false},
	{0x10000, true},
	{0x1000b, true},
	{0x1000c, false},
	{0x1000d, true},
	{0x10026, true},
}

func TestIsWord(t *testing.T) {
	for _, tt := range isWordTests {
		tu.Assert(t, IsWord(tt.args) == tt.want)
	}
}

var mirTests = []struct {
	args rune
	want rune
}{
	{'a', 'a'},
	{'\u2231', '\u2231'},
	{'\u0028', '\u0029'},
	{'\u0029', '\u0028'},
	{'\u22E0', '\u22E1'},
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
	/* Unicode-6.1 character additions */
	{0x27CB, 0x27CD},
	/* Unicode-11.0 character additions */
	{0x2BFE, 0x221F},
	{0x111111, 0x111111},
}

func TestLookupMirrorChar(t *testing.T) {
	for _, tt := range mirTests {
		if got := LookupMirrorChar(tt.args); got != tt.want {
			t.Errorf("LookupMirrorChar() got = %v, want %v", got, tt.want)
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

func BenchmarkLookups(b *testing.B) {
	// b.Run("GeneralCategory unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range generalCategoryTests {
	// 			_ = lookupType(test.args)
	// 		}
	// 	}
	// })
	// b.Run("GeneralCategory packtable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range generalCategoryTests {
	// 			_ = LookupType(test.args)
	// 		}
	// 	}
	// })

	// b.Run("Combining class unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range generalCategoryTests {
	// 			_ = lookupCombiningClass(test.args)
	// 		}
	// 	}
	// })
	// b.Run("Combining class packtable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range generalCategoryTests {
	// 			_ = LookupCombiningClass(test.args)
	// 		}
	// 	}
	// })

	// b.Run("Mirroring unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range generalCategoryTests {
	// 			_ = lookupMirrorChar(test.args)
	// 		}
	// 	}
	// })
	// b.Run("Mirroring packtable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range generalCategoryTests {
	// 			_ = LookupMirrorChar(test.args)
	// 		}
	// 	}
	// })

	// b.Run("Decompose map", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range decomposeTests {
	// 			_, _, _ = decompose(test.ab)
	// 		}
	// 	}
	// })
	// b.Run("Decompose packtable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range decomposeTests {
	// 			_, _, _ = Decompose(test.ab)
	// 		}
	// 	}
	// })
	// b.Run("Compose map", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range composeTests {
	// 			_, _ = compose_(test.a, test.b)
	// 		}
	// 	}
	// })
	// b.Run("Compose packtable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range composeTests {
	// 			_, _ = Compose(test.a, test.b)
	// 		}
	// 	}
	// })

	// b.Run("IsExtendedPictographic unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range extendedPictoTests {
	// 			_ = unicode.Is(Extended_Pictographic, test.r)
	// 		}
	// 	}
	// })
	// b.Run("IsExtendedPictographic packtab", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range extendedPictoTests {
	// 			_ = IsExtendedPictographic(test.r)
	// 		}
	// 	}
	// })

	// b.Run("IsLargeEastAsian unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range eastAsianWidthTests {
	// 			_ = unicode.Is(LargeEastAsian, test.r)
	// 		}
	// 	}
	// })
	// b.Run("IsLargeEastAsian packtab", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range eastAsianWidthTests {
	// 			_ = IsLargeEastAsian(test.r)
	// 		}
	// 	}
	// })

	// b.Run("IndicConjunctBreak unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range indicConjunctBreakTests {
	// 			_ = lookupIndicConjunctBreak(test.r)
	// 		}
	// 	}
	// })
	// b.Run("IndicConjunctBreak packtab", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range indicConjunctBreakTests {
	// 			_ = LookupIndicConjunctBreak(test.r)
	// 		}
	// 	}
	// })

	// b.Run("GraphemeBreak unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range indicConjunctBreakTests {
	// 			_ = lookupGraphemeBreak(test.r)
	// 		}
	// 	}
	// })
	// b.Run("GraphemeBreak packtab", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range indicConjunctBreakTests {
	// 			_ = LookupIndicConjunctBreak(test.r)
	// 		}
	// 	}
	// })

	// b.Run("WordBreak unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range wordBreakTests {
	// 			_ = lookupWordBreak(test.args)
	// 		}
	// 	}
	// })
	// b.Run("WordBreak packtab", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range wordBreakTests {
	// 			_ = LookupIndicConjunctBreak(test.args)
	// 		}
	// 	}
	// })

	// b.Run("IsWord unicode.RangeTable", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range isWordTests {
	// 			_ = unicode.Is(Word, test.args)
	// 		}
	// 	}
	// })
	// b.Run("IsWord packtab", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		for _, test := range isWordTests {
	// 			_ = LookupIndicConjunctBreak(test.args)
	// 		}
	// 	}
	// })

	b.Run("LineBreak unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range lineBreakTests {
				_ = lookupLineBreak(test.args)
			}
		}
	})
	b.Run("LineBreak packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range lineBreakTests {
				_ = LookupLineBreak(test.args)
			}
		}
	})
}

var allCategories = [...]*unicode.RangeTable{
	unicode.Cc,
	unicode.Cf,
	unicode.Co,
	unicode.Cs,
	unicode.Ll,
	unicode.Lm,
	unicode.Lo,
	unicode.Lt,
	unicode.Lu,
	unicode.Mc,
	unicode.Me,
	unicode.Mn,
	unicode.Nd,
	unicode.Nl,
	unicode.No,
	unicode.Pc,
	unicode.Pd,
	unicode.Pe,
	unicode.Pf,
	unicode.Pi,
	unicode.Po,
	unicode.Ps,
	unicode.Sc,
	unicode.Sk,
	unicode.Sm,
	unicode.So,
	unicode.Zl,
	unicode.Zp,
	unicode.Zs,
}

// simple implementation using standard library,
// here for benchmark reference
func lookupType(r rune) *unicode.RangeTable {
	for _, table := range allCategories {
		if unicode.Is(table, r) {
			return table
		}
	}
	return nil
}

// used as reference in benchmark
func lookupCombiningClass(ch rune) uint8 {
	for i, t := range combiningClasses {
		if t == nil {
			continue
		}
		if unicode.Is(t, ch) {
			return uint8(i)
		}
	}
	return 0
}

// used as reference in benchmark
func lookupMirrorChar(ch rune) rune {
	m, ok := mirroring[ch]
	if !ok {
		m = ch
	}
	return m
}

// used as reference in benchmark
func decompose(ab rune) (a, b rune, ok bool) {
	if a, b, ok = decomposeHangul(ab); ok {
		return a, b, true
	}

	// Check if it's a single-character decomposition.
	if m1, ok := decompose1[ab]; ok {
		return m1, 0, true
	}
	if m2, ok := decompose2[ab]; ok {
		return m2[0], m2[1], true
	}
	return ab, 0, false
}

// used as reference in benchmark
func compose_(a, b rune) (rune, bool) {
	// Hangul is handled algorithmically.
	if ab, ok := composeHangul(a, b); ok {
		return ab, true
	}
	u := compose[[2]rune{a, b}]
	return u, u != 0
}

// used as reference in benchmark
func lookupIndicConjunctBreak(r rune) IndicConjunctBreak {
	if unicode.Is(indicCBLinker, r) {
		return ICBLinker
	} else if unicode.Is(indicCBConsonant, r) {
		return ICBConsonant
	} else if unicode.Is(indicCBExtend, r) {
		return ICBExtend
	} else {
		return 0
	}
}

// used as reference in benchmark
func lookupGraphemeBreak(ch rune) *unicode.RangeTable {
	// a lot of runes do not have a grapheme break property :
	// avoid testing all the graphemeBreaks classes for them
	if !unicode.Is(graphemeBreakAll, ch) {
		return nil
	}
	for _, class := range graphemeBreaks {
		if unicode.Is(class, ch) {
			return class
		}
	}
	return nil
}

// used as reference in benchmark
func lookupWordBreak(ch rune) *unicode.RangeTable {
	// a lot of runes do not have a word break property :
	// avoid testing all the wordBreaks classes for them
	if !unicode.Is(wordBreakAll, ch) {
		return nil
	}
	for _, class := range wordBreaks {
		if unicode.Is(class, ch) {
			return class
		}
	}
	return nil
}

// used as reference in benchmark
func lookupLineBreak(ch rune) *unicode.RangeTable {
	for _, class := range lineBreaks {
		if unicode.Is(class, ch) {
			return class
		}
	}
	return nil
}
