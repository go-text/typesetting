package font

type glyphExtents struct {
	valid   bool
	extents GlyphExtents
}

type extentsCache []glyphExtents

func (ec extentsCache) get(gid GID) (GlyphExtents, bool) {
	if int(gid) >= len(ec) {
		return GlyphExtents{}, false
	}
	ge := ec[gid]
	return ge.extents, ge.valid
}

func (ec extentsCache) set(gid GID, extents GlyphExtents) {
	if int(gid) >= len(ec) {
		return
	}
	ec[gid].valid = true
	ec[gid].extents = extents
}

func (ec extentsCache) reset() {
	for i := range ec {
		ec[i] = glyphExtents{}
	}
}

func (f *Face) GlyphExtents(glyph GID) (GlyphExtents, bool) {
	if e, ok := f.extentsCache.get(glyph); ok {
		return e, ok
	}
	e, ok := f.glyphExtentsRaw(glyph)
	if ok {
		f.extentsCache.set(glyph, e)
	}
	return e, ok
}

// advanceCache stores one direction of advances of variable fonts
// without HVAR/VVAR tables. Computing one resolves the full glyph
// outline. The horizontal and vertical caches are separate slices,
// since most fonts are never asked for vertical advances.
// Advances are never negative, so -1 marks an empty entry.
type advanceCache []float32

func newAdvanceCache(nGlyphs int) advanceCache {
	ac := make(advanceCache, nGlyphs)
	ac.reset()
	return ac
}

func (ac advanceCache) get(gid gID) (float32, bool) {
	if int(gid) >= len(ac) {
		return 0, false
	}
	adv := ac[gid]
	return adv, adv >= 0
}

func (ac advanceCache) set(gid gID, advance float32) {
	if int(gid) >= len(ac) {
		return
	}
	ac[gid] = advance
}

func (ac advanceCache) reset() {
	for i := range ac {
		ac[i] = -1
	}
}
