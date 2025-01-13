package fontscan

import (
	"encoding/binary"
	"errors"
	"strings"

	"github.com/go-text/typesetting/language"
)

// LangID is a compact representation of a language
// this package has orthographic knowledge of.
type LangID uint16

// NewLangID returns the compact index of the given language,
// or false if it is not supported by this package.
//
// Derived languages not exactly supported are mapped to their primary part : for instance,
// 'fr-be' is mapped to 'fr'
func NewLangID(l language.Language) (LangID, bool) {
	const N = len(languagesInfo)
	// binary search
	i, j := 0, N
	for i < j {
		h := i + (j-i)/2
		entry := languagesInfo[h]
		if l < entry.lang {
			j = h
		} else if entry.lang < l {
			i = h + 1
		} else {
			// extact match
			return LangID(h), true
		}
	}
	// i is the index where l should be :
	// try to match the primary part
	root := l.Primary()
	for ; i >= 0; i-- {
		entry := languagesInfo[i]
		if entry.lang > root { // keep going
			continue
		} else if entry.lang < root {
			// no root match
			return 0, false
		} else { // found the root
			return LangID(i), true
		}

	}
	return 0, false
}

// LangSet is a bit set for 512 languages
//
// It works as a map[LangID]bool, with the limitation
// that only the 9 low bits of a LangID are used.
// More precisely, the page of a LangID l is given by its 3 "higher" bits : 8-6
// and the bit position by its 6 lower bits : 5-0
type LangSet [8]uint64

// newLangsetFromCoverage compile the languages supported by the given
// rune coverage
func newLangsetFromCoverage(rs RuneSet) (out LangSet) {
	for id, item := range languagesInfo {
		if rs.includes(item.runes) {
			out.Add(LangID(id))
		}
	}
	return out
}

func (ls LangSet) String() string {
	var chunks []string
	for pageN, page := range ls {
		for bit := 0; bit < 64; bit++ {
			if page&(1<<bit) != 0 {
				id := pageN<<6 | bit
				chunks = append(chunks, string(languagesInfo[id].lang))
			}
		}
	}
	return "{" + strings.Join(chunks, "|") + "}"
}

func (ls *LangSet) Add(l LangID) {
	page := (l & 0b111111111 >> 6)
	bit := l & 0b111111
	ls[page] |= 1 << bit
}

func (ls LangSet) Contains(l LangID) bool {
	page := (l & 0b111111111 >> 6)
	bit := l & 0b111111
	return ls[page]&(1<<bit) != 0
}

const langSetSize = 8 * 8

func (ls LangSet) serialize() []byte {
	var buffer [langSetSize]byte
	for i, v := range ls {
		binary.BigEndian.PutUint64(buffer[i*8:], v)
	}
	return buffer[:]
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func (ls *LangSet) deserializeFrom(data []byte) (int, error) {
	if len(data) < langSetSize {
		return 0, errors.New("invalid lang set (EOF)")
	}
	for i := range ls {
		ls[i] = binary.BigEndian.Uint64(data[i*8:])
	}
	return langSetSize, nil
}

// This map returns a language tag that is reasonably
// representative of the script. This will usually be the
// most widely spoken or used language written in that script:
// for instance, the sample language for `Cyrillic`
// is 'ru' (Russian), the sample language for `Arabic` is 'ar'.
//
// For some scripts, no sample language will be returned because there
// is no language that is sufficiently representative. The best
// example of this is `Han`, where various different
// variants of written Chinese, Japanese, and Korean all use
// significantly different sets of Han characters and forms
// of shared characters. No sample language can be provided
// for many historical scripts as well.
//
// inspired by pango/pango-language.c
var scriptToLang = map[language.Script]LangID{
	language.Arabic:   langAr,
	language.Armenian: langHy,
	language.Bengali:  langBn,
	// Used primarily in Taiwan, but not part of the standard
	// zh-tw orthography
	language.Bopomofo: 0,
	language.Cherokee: langChr,
	language.Coptic:   langCop,
	language.Cyrillic: langRu,
	// Deseret was used to write English
	language.Deseret:    0,
	language.Devanagari: langHi,
	language.Ethiopic:   langAm,
	language.Georgian:   langKa,
	language.Gothic:     0,
	language.Greek:      langEl,
	language.Gujarati:   langGu,
	language.Gurmukhi:   langPa,
	language.Han:        0,
	language.Hangul:     langKo,
	language.Hebrew:     langHe,
	language.Hiragana:   langJa,
	language.Kannada:    langKn,
	language.Katakana:   langJa,
	language.Khmer:      langKm,
	language.Lao:        langLo,
	language.Latin:      langEn,
	language.Malayalam:  langMl,
	language.Mongolian:  langMn,
	language.Myanmar:    langMy,
	// Ogham was used to write old Irish
	language.Ogham:               0,
	language.Old_Italic:          0,
	language.Oriya:               langOr,
	language.Runic:               0,
	language.Sinhala:             langSi,
	language.Syriac:              langSyr,
	language.Tamil:               langTa,
	language.Telugu:              langTe,
	language.Thaana:              langDv,
	language.Thai:                langTh,
	language.Tibetan:             langBo,
	language.Canadian_Aboriginal: langIu,
	language.Yi:                  0,
	language.Tagalog:             langTl,
	// Phillipino languages/scripts
	language.Hanunoo:  langHnn,
	language.Buhid:    langBku,
	language.Tagbanwa: langTbw,

	language.Braille: 0,
	language.Cypriot: 0,
	language.Limbu:   0,
	// Used for Somali (so) in the past
	language.Osmanya: 0,
	// The Shavian alphabet was designed for English
	language.Shavian:  0,
	language.Linear_B: 0,
	language.Tai_Le:   0,
	language.Ugaritic: langUga,

	language.New_Tai_Lue: 0,
	language.Buginese:    langBug,
	// The original script for Old Church Slavonic (chu), later
	// written with Cyrillic
	language.Glagolitic: 0,
	// Used for for Berber (ber), but Arabic script is more common
	language.Tifinagh:     0,
	language.Syloti_Nagri: langSyl,
	language.Old_Persian:  langPeo,

	language.Nko: langNqo,
}
