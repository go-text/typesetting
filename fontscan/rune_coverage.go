package fontscan

import (
	"encoding/binary"
	"errors"
	"math/bits"

	"github.com/benoitkugler/textlayout/fonts"
)

// Rune rserage implementation, inspired by the fontconfig FcCharset type.
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

// newRuneSet builds a set containing the given runes.
func newRuneSet(runes ...rune) runeSet {
	var rs runeSet
	for _, r := range runes {
		rs.Add(r)
	}
	return rs
}

// newRuneSetFromCmap iterates through the given `cmap`
// to build the corresponding rune set.
func newRuneSetFromCmap(cmap fonts.Cmap) runeSet {
	var rs runeSet
	iter := cmap.Iter()
	for iter.Next() {
		r, _ := iter.Char()
		rs.Add(r)
	}
	return rs
}

// runes returns a copy of the runes in the set.
func (rs runeSet) runes() (out []rune) {
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

// return true if and only if `a` is a subset of `b`
func (a runeSet) isSubset(b runeSet) bool {
	ai, bi := 0, 0
	for ai < len(a) && bi < len(b) {
		an := a[ai].ref
		bn := b[bi].ref
		// Check matching pages
		if an == bn {
			am := a[ai].set
			bm := b[bi].set

			if am != bm {
				//  Does am have any bits not in bm?
				for j, av := range am {
					if av & ^bm[j] != 0 {
						return false
					}
				}
			}
			ai++
			bi++
		} else if an < bn { // Does a have any pages not in b?
			return false
		} else {
			bi = b.findPageFrom(bi+1, an)
			if bi < 0 {
				bi = -bi - 1
			}
		}
	}
	// did we look at every page?
	return ai >= len(a)
}

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

// serializeTo serialize the Coverage in binary format
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
