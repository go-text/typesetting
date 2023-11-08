package fontscan

// LanguageID is a compact representation of a language
// this package has orthographic knowledge of.
type LanguageID uint16

// langset is a bit set for 512 languages
// the page of a LanguageID l is given by its 3 high bits : 8-6
// and the bit position by its 6 lower bits : 5-0
type langset [8]uint64

func (ls *langset) add(l LanguageID) {
	page := (l & 0b111111111 >> 6)
	bit := l & 0b111111
	ls[page] |= 1 << bit
}

func (ls langset) contains(l LanguageID) bool {
	page := (l & 0b111111111 >> 6)
	bit := l & 0b111111
	return ls[page]&(1<<bit) != 0
}
