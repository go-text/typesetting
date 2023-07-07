package fontscan

import (
	"encoding/binary"
	"errors"
	"math/bits"
	"sort"

	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/opentype/api"
)

// Rune coverage implementation, inspired by the fontconfig FcCharset type.
//
// The internal representation is a slice of `pageSet` pages, where each page is a boolean
// set of size 256, encoding the last byte of a rune.
// Each rune is then mapped to a page index (`pageNumber`), defined by it second and third bytes.

// pageSet is the base storage for a compact rune set.
// A rune is first reduced to its lower byte 'b'. Then the index
// of 'b' in the page is given by the 3 high bits (from 0 to 7)
// and the position in the resulting uint32 is given by the 5 lower bits (from 0 to 31)
type pageSet [8]uint32

// pageRef stores the second and third bytes of a rune (uint16(r >> 8)),
// shared by all the runes in a page.
type pageRef = uint16

type runePage struct {
	ref pageRef
	set pageSet
}

// runeSet is an efficient implementation of a rune set (that is a map[rune]bool),
// used to store the Unicode points supported by a font, and optimized to deal with consecutive
// runes.
type runeSet []runePage

// newRuneSetFromCmap iterates through the given `cmap`
// to build the corresponding rune set.
// buffer may be provided to reduce allocations, and is returned
func newRuneSetFromCmap(cmap api.Cmap, buffer [][2]rune) (runeSet, [][2]rune) {
	if ranger, ok := cmap.(api.CmapRuneRanger); ok { // use the fast range implementation
		return newRuneSetFromCmapRange(ranger, buffer)
	}

	var rs runeSet
	iter := cmap.Iter()
	for iter.Next() {
		r, _ := iter.Char()
		rs.Add(r)
	}
	return rs, buffer
}

// assume a <= b
func addRangeToPage(page *pageSet, start, end byte) {
	// indexes in [0; 8[
	uintIndexStart := start >> 5
	uintIndexEnd := end >> 5

	// bit index, in [0; 32[
	bitIndexStart := (start & 0x1f)
	bitIndexEnd := (end & 0x1f)

	// handle the start uint
	bitEnd := byte(31)
	if uintIndexEnd == uintIndexStart {
		bitEnd = bitIndexEnd
	}
	b := &page[uintIndexStart]
	alt := (uint32(1)<<(bitEnd-bitIndexStart+1) - 1) << bitIndexStart // mask for bits from a to b (included)
	*b |= alt

	// handle the end uint, when required
	if uintIndexEnd != uintIndexStart {
		// fill uint between with ones
		for index := uintIndexStart + 1; index < uintIndexEnd; index++ {
			page[index] = 0xFFFFFFFF
		}

		// handle the last
		b := &page[uintIndexEnd]
		alt := (uint32(1)<<(bitIndexEnd+1) - 1) // mask for bits from a to b (included)
		*b |= alt
	}
}

// newRuneSetFromCmapRange iterates through the given `cmap`
// to build the corresponding rune set.
func newRuneSetFromCmapRange(cmap api.CmapRuneRanger, buffer [][2]rune) (runeSet, [][2]rune) {
	buffer = cmap.RuneRanges(buffer)
	var rs runeSet
	lastPage := &runePage{ref: 0xFFFF} // start with an invalid sentinel value
	for _, ra := range buffer {
		start, end := ra[0], ra[1]

		pageStart, pageEnd := uint16(start>>8), uint16(end>>8)

		// handle the starting page
		startByte, endByte := byte(start&0xff), byte(end&0xff)
		endByteClamped := byte(0xFF)
		if pageEnd == pageStart {
			endByteClamped = endByte
		}

		// check if we can reuse the last page
		var leaf *pageSet
		if pageStart == lastPage.ref { // use the same page
			leaf = &lastPage.set
		} else {
			rs = append(rs, runePage{ref: pageStart})
			leaf = &rs[len(rs)-1].set
		}
		addRangeToPage(leaf, startByte, endByteClamped)

		// handle the next
		if pageEnd != pageStart { // this means pageStart < pageEnd
			// fill the strictly intermediate pages with ones
			for pageIndex := pageStart + 1; pageIndex < pageEnd; pageIndex++ {
				rs = append(rs, runePage{
					ref: pageIndex,
					set: pageSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
				})
			}

			// hande the last
			rs = append(rs, runePage{ref: pageEnd})
			leaf = &rs[len(rs)-1].set
			addRangeToPage(leaf, 0, endByte)
		}

		lastPage = &rs[len(rs)-1]
	}
	return rs, buffer
}

// findPageFrom is the same as findPagePos, but
// start the binary search with the given `low` index
func (rs runeSet) findPageFrom(low int, ref pageRef) int {
	high := len(rs) - 1
	for low <= high {
		mid := (low + high) >> 1
		page := rs[mid].ref
		if page == ref {
			return mid // found the page
		}
		if page < ref {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if high < 0 || (high < len(rs) && rs[high].ref < ref) {
		high++
	}
	return -(high + 1) // the page is not in the set, but should be inserted at high
}

// findPagePos searches for the leaf containing the specified number.
// It returns its index if it exists, otherwise it returns the negative of
// the (`position` + 1) where `position` is the index where it should be inserted
func (rs runeSet) findPagePos(page pageRef) int { return rs.findPageFrom(0, page) }

// findPage returns the page containing the specified char, or nil
// if it doesn't exists
func (rs runeSet) findPage(ref pageRef) *pageSet {
	pos := rs.findPagePos(ref)
	if pos >= 0 {
		return &rs[pos].set
	}
	return nil
}

// findOrCreatePage locates the page containing the specified char, creating it if needed,
// and returns a pointer to it
func (rs *runeSet) findOrCreatePage(ref pageRef) *pageSet {
	pos := rs.findPagePos(ref)
	if pos < 0 { // the page doest not exists, create it
		pos = -pos - 1
		rs.insertPage(runePage{ref: ref}, pos)
	}

	return &(*rs)[pos].set
}

// insertPage inserts the given `page` at `pos`, meaning the resulting page can be accessed via &rs[pos]
func (rs *runeSet) insertPage(page runePage, pos int) {
	// insert in slice
	*rs = append(*rs, runePage{})
	copy((*rs)[pos+1:], (*rs)[pos:])
	(*rs)[pos] = page
}

// Add adds `r` to the rune set.
func (rs *runeSet) Add(r rune) {
	leaf := rs.findOrCreatePage(uint16(r >> 8))
	b := &leaf[(r&0xff)>>5] // (r&0xff)>>5 is the index in the page
	*b |= (1 << (r & 0x1f)) // r & 0x1f is the bit in the uint32
}

// Delete removes the rune from the rune set.
func (rs runeSet) Delete(r rune) {
	leaf := rs.findPage(uint16(r >> 8))
	if leaf == nil {
		return
	}
	b := &leaf[(r&0xff)>>5]  // (r&0xff)>>5 is the index in the page
	*b &= ^(1 << (r & 0x1f)) // r & 0x1f is the bit in the uint32
	// we don't bother removing the leaf if it's empty
}

// Contains returns `true` if `r` is in the set.
func (rs *runeSet) Contains(r rune) bool {
	leaf := rs.findPage(uint16(r >> 8))
	if leaf == nil {
		return false
	}
	return leaf[(r&0xff)>>5]&(1<<(r&0x1f)) != 0
}

// Len returns the number of runes in the set.
func (a runeSet) Len() int {
	count := 0
	for _, page := range a {
		for _, am := range page.set {
			count += bits.OnesCount32(am)
		}
	}
	return count
}

const runePageSize = 2 + 8*4 // uint16 + 8 * uint32

// serialize serialize the Coverage in binary format
func (rs runeSet) serialize() []byte {
	buffer := make([]byte, 2+runePageSize*len(rs))
	binary.BigEndian.PutUint16(buffer, uint16(len(rs)))
	for i, page := range rs {
		binary.BigEndian.PutUint16(buffer[2+runePageSize*i:], page.ref)
		slice := buffer[2+runePageSize*i+2:]
		for j, k := range page.set {
			binary.BigEndian.PutUint32(slice[4*j:], k)
		}
	}
	return buffer
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func (rs *runeSet) deserializeFrom(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("invalid Coverage (EOF)")
	}
	L := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+runePageSize*L {
		return 0, errors.New("invalid Coverage size (EOF)")
	}
	v := make(runeSet, L)
	for i := range v {
		v[i].ref = binary.BigEndian.Uint16(data[2+runePageSize*i:])
		slice := data[2+runePageSize*i+2:]
		for j := range v[i].set {
			v[i].set[j] = binary.BigEndian.Uint32(slice[4*j:])
		}
	}

	*rs = v

	return 2 + runePageSize*L, nil
}

type scriptSet []language.Script

// insert adds the given script to the set if it is not already present.
func (s *scriptSet) insert(newScript language.Script) {
	scriptIdx := sort.Search(len([]language.Script(*s)), func(i int) bool {
		return (*s)[i] >= newScript
	})
	if scriptIdx != len(*s) && (*s)[scriptIdx] == newScript {
		return
	}
	// Grow the slice if necessary.
	startLen := len(*s)
	*s = append(*s, language.Script(0))[:startLen]
	// Shift all elements from scriptIdx onward to the right one position.
	*s = append((*s)[:scriptIdx+1], (*s)[scriptIdx:]...)
	// Insert newScript at the correct position.
	(*s)[scriptIdx] = newScript
}

const scriptSize = 4

// serialize serialize the script set in binary format
func (ss scriptSet) serialize() []byte {
	buffer := make([]byte, 1+scriptSize*len(ss))
	buffer[0] = byte(len(ss)) // there are about 190 scripts, a byte is enough
	for i, script := range ss {
		binary.BigEndian.PutUint32(buffer[1+scriptSize*i:], uint32(script))
	}
	return buffer
}

// deserializeFrom reads the binary format produced by serialize
// it returns the number of bytes read from `data`
func (ss *scriptSet) deserializeFrom(data []byte) (int, error) {
	if len(data) < 1 {
		return 0, errors.New("invalid Script set (EOF)")
	}
	L := int(data[0])
	if len(data) < 1+scriptSize*L {
		return 0, errors.New("invalid Script set size (EOF)")
	}
	v := make(scriptSet, L)
	for i := range v {
		v[i] = language.Script(binary.BigEndian.Uint32(data[1+scriptSize*i:]))
	}

	*ss = v

	return 1 + scriptSize*L, nil
}

// Scripts returns an approximation of the scripts that the runeSet has coverage for.
// It works by sampling the coverage set for the first covered rune in every page and
// mapping that to a supported script. This means that it can miss some supported
// scripts.
func (rs runeSet) Scripts() []language.Script {
	scripts := make(scriptSet, 0, 1)
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

// scriptsFromRanges returns the set of scripts used in [ranges],
// which must be sorted (in ascending order).
// The ranges have inclusive bounds.
func scriptsFromRanges(ranges [][2]rune) scriptSet {
	out := make(scriptSet, 0, 2)

	// we leverage the fact that both ranges and scriptRanges are sorted
	// to loop through both slices at the same time
	indexS := 0 // indices in ranges and scriptRanges
	for _, ra := range ranges {
		start, end := ra[0], ra[1]
		// find the scriptItem for start
		for indexS < len(language.ScriptRanges) && language.ScriptRanges[indexS].End < start {
			indexS++
		}
		// we now have start <= ScriptRange.End
		// we can add the script for every ScriptRange such that ScriptRangeStart <= end
		for indexS < len(language.ScriptRanges) && language.ScriptRanges[indexS].Start <= end {
			// we have a covered script
			out.insert(language.ScriptRanges[indexS].Script)
			indexS++
		}
	}

	return out
}
