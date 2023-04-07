package metadata

import "github.com/go-text/typesetting/opentype/tables"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func approximatelyEqual(x, y int) bool { return abs(x-y)*33 <= max(abs(x), abs(y)) }

func (fd *fontDescriptor) isMonospace() bool {
	// code adapted from fontconfig

	if fd.cmap == nil || fd.metrics.IsEmpty() {
		// we can't be sure, so be conservative
		return false
	}

	var firstAdvance int
	iter := fd.cmap.Iter()
	for iter.Next() {
		_, glyph := iter.Char()
		advance := int(fd.metrics.Advance(tables.GlyphID(glyph)))
		if advance == 0 { // do not count zero as a proper width
			continue
		}

		if firstAdvance == 0 {
			firstAdvance = advance
			continue
		}

		if approximatelyEqual(advance, firstAdvance) {
			continue
		}

		// two distinct advances : the font is not monospace
		return false
	}

	return true
}
