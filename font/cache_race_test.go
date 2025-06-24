package font

import (
	"bytes"
	"sync"
	"testing"

	td "github.com/go-text/typesetting-utils/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
)

// TestGlyphExtentsCacheRace tests concurrent access to the glyph extents cache
func TestGlyphExtentsCacheRace(t *testing.T) {
	// Load a test font
	face := loadTestFace(t)
	if face == nil {
		t.Skip("Test font not available")
	}

	// Test concurrent reads and writes to the cache
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Each goroutine accesses different and overlapping glyphs
			for j := 0; j < numOperations; j++ {
				// Access glyphs in a pattern that causes overlap
				gid := GID((goroutineID + j) % 100)

				// This should trigger cache reads and writes
				_, _ = face.GlyphExtents(gid)
			}
		}(i)
	}

	wg.Wait()
}

// TestGlyphExtentsCacheReset tests concurrent cache resets with accesses
func TestGlyphExtentsCacheReset(t *testing.T) {
	face := loadTestFace(t)
	if face == nil {
		t.Skip("Test font not available")
	}

	var wg sync.WaitGroup

	// Goroutine 1: Continuously access glyphs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			gid := GID(i % 50)
			_, _ = face.GlyphExtents(gid)
		}
	}()

	// Goroutine 2: Periodically reset cache via SetPpem
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			face.SetPpem(uint16(12+i%4), uint16(12+i%4))
		}
	}()

	// Goroutine 3: Periodically reset cache via SetCoords
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Empty coords, just to trigger reset
			face.SetCoords(nil)
		}
	}()

	wg.Wait()
}

// TestGlyphExtentsCacheConcurrentSameGlyph tests multiple goroutines accessing the same glyph
func TestGlyphExtentsCacheConcurrentSameGlyph(t *testing.T) {
	face := loadTestFace(t)
	if face == nil {
		t.Skip("Test font not available")
	}

	var wg sync.WaitGroup
	numGoroutines := 20
	targetGlyph := GID(42) // All goroutines will access this glyph

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each goroutine tries to access the same glyph many times
			for j := 0; j < 100; j++ {
				extents, ok := face.GlyphExtents(targetGlyph)
				if !ok {
					t.Errorf("Failed to get extents for glyph %d", targetGlyph)
				}
				// Verify we get consistent results
				if ok && extents.Width < 0 {
					t.Errorf("Got invalid width: %f", extents.Width)
				}
			}
		}()
	}

	wg.Wait()
}

// TestGlyphExtentsCacheBoundary tests boundary conditions
func TestGlyphExtentsCacheBoundary(t *testing.T) {
	face := loadTestFace(t)
	if face == nil {
		t.Skip("Test font not available")
	}

	var wg sync.WaitGroup

	// Test accessing glyphs at cache boundaries
	wg.Add(3)

	// Access very high GIDs
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			gid := GID(65000 + i) // Likely out of bounds
			_, _ = face.GlyphExtents(gid)
		}
	}()

	// Access GID 0
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_, _ = face.GlyphExtents(0)
		}
	}()

	// Access normal range
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			gid := GID(i % 256)
			_, _ = face.GlyphExtents(gid)
		}
	}()

	wg.Wait()
}

// TestPpemFieldRace tests concurrent access to xPpem/yPpem fields
func TestPpemFieldRace(t *testing.T) {
	face := loadTestFace(t)
	if face == nil {
		t.Skip("Test font not available")
	}

	var wg sync.WaitGroup

	// Goroutine 1: Continuously read Ppem
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			x, y := face.Ppem()
			// Verify values are reasonable
			if x > 1000 || y > 1000 {
				t.Errorf("Unexpected ppem values: %d, %d", x, y)
			}
		}
	}()

	// Goroutine 2: Continuously set Ppem
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ppem := uint16(12 + i%20)
			face.SetPpem(ppem, ppem)
		}
	}()

	// Goroutine 3: Use glyphExtentsRaw which reads xPpem/yPpem
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			// This internally uses xPpem/yPpem
			_, _ = face.GlyphExtents(GID(i % 100))
		}
	}()

	wg.Wait()
}

// TestCoordsFieldRace tests concurrent access to coords field
func TestCoordsFieldRace(t *testing.T) {
	face := loadTestFace(t)
	if face == nil {
		t.Skip("Test font not available")
	}

	var wg sync.WaitGroup

	// Goroutine 1: Continuously read Coords
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			coords := face.Coords()
			// Just access the slice to ensure it's valid
			_ = len(coords)
		}
	}()

	// Goroutine 2: Continuously set Coords
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if i%2 == 0 {
				face.SetCoords([]tables.Coord{0, 0})
			} else {
				face.SetCoords(nil)
			}
		}
	}()

	// Goroutine 3: Call methods that use coords internally
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			// These methods check isVar() which reads coords
			_, _ = face.FontHExtents()
			_, _ = face.FontVExtents()
		}
	}()

	wg.Wait()
}

// TestMixedOperationsRace tests all operations together
func TestMixedOperationsRace(t *testing.T) {
	face := loadTestFace(t)
	if face == nil {
		t.Skip("Test font not available")
	}

	var wg sync.WaitGroup

	// Goroutine 1: GlyphExtents calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			_, _ = face.GlyphExtents(GID(i % 200))
		}
	}()

	// Goroutine 2: SetPpem calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			face.SetPpem(uint16(10+i%10), uint16(10+i%10))
		}
	}()

	// Goroutine 3: SetCoords calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if i%3 == 0 {
				face.SetCoords([]tables.Coord{0})
			} else {
				face.SetCoords(nil)
			}
		}
	}()

	// Goroutine 4: Read operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			switch i % 4 {
			case 0:
				face.Ppem()
			case 1:
				face.Coords()
			case 2:
				face.FontHExtents()
			case 3:
				face.HorizontalAdvance(GID(i % 50))
			}
		}
	}()

	wg.Wait()
}

// loadTestFace loads a face for testing
func loadTestFace(t *testing.T) *Face {
	t.Helper()

	// Try multiple font paths
	fontPaths := []string{
		"common/DejaVuSans.ttf",        // A common font available in the test data
		"common/Roboto-BoldItalic.ttf", // Alternative Roboto variant
		"common/mplus-1p-regular.ttf",  // Another alternative
	}

	var fontData []byte
	var err error
	var foundPath string

	for _, path := range fontPaths {
		fontData, err = td.Files.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to load any test font. Tried paths: %v. Last error: %v", fontPaths, err)
	}

	face, err := ParseTTF(bytes.NewReader(fontData))
	if err != nil {
		t.Fatalf("Failed to parse font from %s: %v", foundPath, err)
	}

	return face
}
