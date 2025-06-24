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
	f.cacheMu.RLock()
	if e, ok := f.extentsCache.get(glyph); ok {
		f.cacheMu.RUnlock()
		return e, ok
	}
	f.cacheMu.RUnlock()

	e, ok := f.glyphExtentsRaw(glyph)
	if ok {
		f.cacheMu.Lock()
		f.extentsCache.set(glyph, e)
		f.cacheMu.Unlock()
	}
	return e, ok
}
