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
	const N = len(languagesRunes)
	// binary search
	i, j := 0, N
	for i < j {
		h := i + (j-i)/2
		entry := languagesRunes[h]
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
		entry := languagesRunes[i]
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

// langset is a bit set for 512 languages
// the page of a LanguageID l is given by its 3 high bits : 8-6
// and the bit position by its 6 lower bits : 5-0
type langset [8]uint64

// newLangsetFromCoverage compile the languages supported by the given
// rune coverage
func newLangsetFromCoverage(rs runeSet) (out langset) {
	for id, item := range languagesRunes {
		if rs.includes(item.runes) {
			out.add(LangID(id))
		}
	}
	return out
}

func (ls langset) String() string {
	var chunks []string
	for pageN, page := range ls {
		for bit := 0; bit < 64; bit++ {
			if page&(1<<bit) != 0 {
				id := pageN<<6 | bit
				chunks = append(chunks, string(languagesRunes[id].lang))
			}
		}
	}
	return "{" + strings.Join(chunks, "|") + "}"
}

func (ls *langset) add(l LangID) {
	page := (l & 0b111111111 >> 6)
	bit := l & 0b111111
	ls[page] |= 1 << bit
}

func (ls langset) contains(l LangID) bool {
	page := (l & 0b111111111 >> 6)
	bit := l & 0b111111
	return ls[page]&(1<<bit) != 0
}

const langsetSize = 8 * 8

func (ls langset) serialize() []byte {
	var buffer [langsetSize]byte
	for i, v := range ls {
		binary.BigEndian.PutUint64(buffer[i*8:], v)
	}
	return buffer[:]
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func (ls *langset) deserializeFrom(data []byte) (int, error) {
	if len(data) < langsetSize {
		return 0, errors.New("invalid lang set (EOF)")
	}
	for i := range ls {
		ls[i] = binary.BigEndian.Uint64(data[i*8:])
	}
	return langsetSize, nil
}
