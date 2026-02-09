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
		{'\u058A', BreakHH},     //	ARMENIAN HYPHEN
		{'\u2010', BreakHH},     //	HYPHEN
		{'\u2012', BreakHH},     //	FIGURE DASH
		{'\u2013', BreakHH},     //	EN DASH
		{'\u05BE', BreakHH},     //	HEBREW PUNCTUATION MAQAF
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
		{'\u2E17', BreakHH},     //	DOUBLE OBLIQUE HYPHEN
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
		{'\u034F', BreakCM}, //	COMBINING GRAPHEME JOINER
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
		{'\uFE10', BreakCL}, //	PRESENTATION FORM FOR VERTICAL COMMA
		{'\uFE13', BreakNS}, //	PRESENTATION FORM FOR VERTICAL COLON
		{'\uFE14', BreakNS}, //	PRESENTATION FORM FOR VERTICAL SEMICOLON

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
		if got := LookupLineBreakClass(tt.args); got != tt.want {
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
	b.Run("GeneralCategory unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = lookupType(test.args)
			}
		}
	})
	b.Run("GeneralCategory packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = LookupType(test.args)
			}
		}
	})

	b.Run("Combining class unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = lookupCombiningClass(test.args)
			}
		}
	})
	b.Run("Combining class packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = LookupCombiningClass(test.args)
			}
		}
	})

	b.Run("Mirroring unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = lookupMirrorChar(test.args)
			}
		}
	})
	b.Run("Mirroring packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range generalCategoryTests {
				_ = LookupMirrorChar(test.args)
			}
		}
	})

	b.Run("Decompose map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range decomposeTests {
				_, _, _ = decompose(test.ab)
			}
		}
	})
	b.Run("Decompose packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range decomposeTests {
				_, _, _ = Decompose(test.ab)
			}
		}
	})
	b.Run("Compose map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range composeTests {
				_, _ = compose_(test.a, test.b)
			}
		}
	})
	b.Run("Compose packtable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range composeTests {
				_, _ = Compose(test.a, test.b)
			}
		}
	})

	b.Run("IsExtendedPictographic unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range extendedPictoTests {
				_ = unicode.Is(Extended_Pictographic, test.r)
			}
		}
	})
	b.Run("IsExtendedPictographic packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range extendedPictoTests {
				_ = IsExtendedPictographic(test.r)
			}
		}
	})

	b.Run("IsLargeEastAsian unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range eastAsianWidthTests {
				_ = unicode.Is(LargeEastAsian, test.r)
			}
		}
	})
	b.Run("IsLargeEastAsian packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range eastAsianWidthTests {
				_ = IsLargeEastAsian(test.r)
			}
		}
	})

	b.Run("IndicConjunctBreak unicode.RangeTable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range indicConjunctBreakTests {
				_ = lookupIndicConjunctBreak(test.r)
			}
		}
	})
	b.Run("IndicConjunctBreak packtab", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range indicConjunctBreakTests {
				_ = LookupIndicConjunctBreak(test.r)
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
		return InCBLinker
	} else if unicode.Is(indicCBConsonant, r) {
		return InCBConsonant
	} else if unicode.Is(indicCBExtend, r) {
		return InCBExtend
	} else {
		return 0
	}
}
