package font

import "sync"

type glyphExtents struct {
	valid   bool
	extents GlyphExtents
}

type extentsCache struct {
	mu    sync.RWMutex
	elems []glyphExtents
}

func newExtentsCache(n int) extentsCache {
	return extentsCache{elems: make([]glyphExtents, n)}
}

func (ec *extentsCache) get(gid GID) (GlyphExtents, bool) {
	ec.mu.RLock()
	if int(gid) >= len(ec.elems) {
		ec.mu.RUnlock()
		return GlyphExtents{}, false
	}
	ge := ec.elems[gid]
	ec.mu.RUnlock()
	return ge.extents, ge.valid
}

func (ec *extentsCache) set(gid GID, extents GlyphExtents) {
	ec.mu.Lock()
	if int(gid) < len(ec.elems) {
		ec.elems[gid].valid = true
		ec.elems[gid].extents = extents
	}
	ec.mu.Unlock()
}

func (ec *extentsCache) reset() {
	ec.mu.Lock()
	for i := range ec.elems {
		ec.elems[i] = glyphExtents{}
	}
	ec.mu.Unlock()
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
