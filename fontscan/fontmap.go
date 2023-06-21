package fontscan

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/go-text/typesetting/font"
	meta "github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
)

type cacheEntry struct {
	font.Face
	meta.Description
}

// The family substitution algorithm is copied from fontconfig
// and the match algorithm is inspired from Rust font-kit library

// FontMap provides a mechanism to select a [font.Face] from a font description.
// It supports system and user-provided fonts, and implements the CSS font substitutions
// rules.
//
// A typical usage would be as following :
//
//	fontMap := NewFontMap()
//
//	// at least one of the following calls
//	fontMap.UseSystemFonts() // error handling omitted
//	fontMap.AddFont(font1, "font1") // error handling omitted
//	fontMap.AddFont(font2, "font2") // error handling omitted
//
//	// set the font description
//	fontMap.SetQuery(Query{Families: []string{"Arial", "serif"}}) // regular Aspect
//
//	// `fontMap` is now ready for text shaping, using the `ResolveFace` method
//
// Note that [FontMap] is NOT safe for concurrent use, but several font maps may coexist
// in an application.
//
// [FontMap] is designed to work with an index built by scanning the system fonts,
// which is a costly operation (see [UseSystemFonts] for more details).
// A lightweight alternative is provided by the [FindFont] function, which only uses
// file paths to select a font.
type FontMap struct {
	logger *log.Logger
	// cache of already loaded faces
	faces map[Location]cacheEntry

	// the database to query, either loaded from an index
	// or populated with the UseSystemFonts and AddFont method
	database fontSet

	// the candidates for the current query, which influences ResolveFace output
	candidates candidates

	// internal buffers used in SetQuery
	footprintsBuffer scoredFootprints
	cribleBuffer     familyCrible

	query Query // current query

	// cached value of the last footprint index
	// selected by ResolveFace
	lastFootprintIndex int
}

// NewFontMap return a new font map, which should be filled with the `UseSystemFonts`
// or `AddFont` methods. The provided logger will be used to record non-fatal errors
// encountered during font loading. If logger is nil, log.Default() is used.
func NewFontMap(logger *log.Logger) *FontMap {
	if logger == nil {
		logger = log.New(log.Writer(), "fontscan", log.Flags())
	}
	return &FontMap{
		logger:             logger,
		faces:              make(map[Location]cacheEntry),
		cribleBuffer:       make(familyCrible),
		lastFootprintIndex: -1,
	}
}

// UseSystemFonts loads the system fonts and adds them to the font map.
// This method is safe for concurrent use, but should only be called once
// per font map.
// The first call of this method trigger a rather long scan.
// A per-application on-disk cache is used to speed up subsequent initialisations.
// Callers can provide an appropriate directory path within which this cache may be
// stored.
//
// NOTE: On Android, callers *must* provide a writable path manually, as it cannot
// be inferred without access to the Java runtime environment of the application.
func (fm *FontMap) UseSystemFonts(cacheDir string) error {
	// safe for concurrent use; subsequent calls are no-ops
	err := initSystemFonts(fm.logger, cacheDir)
	if err != nil {
		return err
	}

	// systemFonts is read-only, so may be used concurrently
	fm.database = append(fm.database, systemFonts.flatten()...)

	fm.buildCandidates()

	return nil
}

// systemFonts is a global index of the system fonts.
// initSystemFontsOnce protects the initial assignment,
// and `systemFonts` use is then read-only
var (
	systemFonts         systemFontsIndex
	initSystemFontsOnce sync.Once
)

func cacheDir(userProvided string) (string, error) {
	if userProvided != "" {
		return userProvided, nil
	}
	// load an existing index
	if runtime.GOOS == "android" {
		// There is no stable way to infer the proper place to store the cache
		// with access to the Java runtime for the application. Rather than
		// clutter our API with that, require the caller to provide a path.
		return "", fmt.Errorf("user must provide cache directory on android")
	}
	configDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolving index cache path: %s", err)
	}
	return configDir, nil
}

// initSystemFonts scan the system fonts and update `SystemFonts`.
// If the returned error is nil, `SystemFonts` is guaranteed to contain
// at least one valid font.Face.
// It is protected by sync.Once, and is then safe to use by multiple goroutines.
func initSystemFonts(logger *log.Logger, userCacheDir string) error {
	var err error

	initSystemFontsOnce.Do(func() {
		const cacheFile = "font_index.cache"

		// load an existing index
		var dir string
		dir, err = cacheDir(userCacheDir)
		if err != nil {
			return
		}

		cachePath := filepath.Join(dir, cacheFile)

		systemFonts, err = refreshSystemFontsIndex(logger, cachePath)
	})

	return err
}

func refreshSystemFontsIndex(logger *log.Logger, cachePath string) (systemFontsIndex, error) {
	fontDirectories, err := DefaultFontDirectories(logger)
	if err != nil {
		return nil, fmt.Errorf("searching font directories: %s", err)
	}
	logger.Printf("using system font dirs %q", fontDirectories)

	currentIndex, _ := deserializeIndexFile(cachePath)
	// if an error occured (the cache file does not exists or is invalid), we start from scratch

	updatedIndex, err := scanFontFootprints(logger, currentIndex, fontDirectories...)
	if err != nil {
		return nil, fmt.Errorf("scanning system fonts: %s", err)
	}

	// since ResolveFace must always return a valid face, we make sure
	// at least one font exists and is valid.
	// Otherwise, the font map is useless; this is an extreme case anyway.
	err = updatedIndex.assertValid()
	if err != nil {
		return nil, fmt.Errorf("loading system fonts: %s", err)
	}

	// write back the index in the cache file
	err = updatedIndex.serializeToFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("updating cache: %s", err)
	}

	return updatedIndex, nil
}

// [AddFont] loads the faces contained in [fontFile] and add them to
// the font map.
// [fileID] is used as the [Location.File] entry returned by [FontLocation].
//
// If `familyName` is not empty, it is used as the family name for `fontFile`
// instead of the one found in the font file.
//
// An error is returned if the font resource is not supported.
func (fm *FontMap) AddFont(fontFile font.Resource, fileID, familyName string) error {
	loaders, err := loader.NewLoaders(fontFile)
	if err != nil {
		return fmt.Errorf("unsupported font resource: %s", err)
	}

	// eagerly load the faces
	faces, err := font.ParseTTC(fontFile)
	if err != nil {
		return fmt.Errorf("unsupported font resource: %s", err)
	}

	// by construction of fonts.Loader and fonts.FontDescriptor,
	// fontDescriptors and face have the same length
	if len(faces) != len(loaders) {
		panic("internal error: inconsistent font descriptors and loader")
	}

	var addedFonts []footprint
	for i, fontDesc := range loaders {
		fp, err := newFootprintFromLoader(fontDesc)
		// the font won't be usable, just ignore it
		if err != nil {
			continue
		}

		fp.Location.File = fileID
		fp.Location.Index = uint16(i)
		// TODO: for now, we do not handle variable fonts

		if familyName != "" {
			// give priority to the user provided family
			fp.Family = meta.NormalizeFamily(familyName)
		}

		addedFonts = append(addedFonts, fp)
		fm.faces[fp.Location] = cacheEntry{Face: faces[i], Description: fp.metadata()}
	}

	if len(addedFonts) == 0 {
		return fmt.Errorf("empty font resource %s", fileID)
	}

	fm.database = append(fm.database, addedFonts...)

	fm.buildCandidates()

	return nil
}

// FontLocation look for the given font among the loaded font map fonts
// to find its origin.
// FontLocation should only be called for faces returned by `ResolveFace`,
// otherwise the returned Location will be empty.
func (fm *FontMap) FontLocation(ft font.Font) Location {
	for location, cachedFace := range fm.faces {
		if cachedFace.Font == ft {
			return location
		}
	}
	return Location{}
}

// SetQuery set the families and aspect required, influencing subsequent
// `ResolveFace` calls.
func (fm *FontMap) SetQuery(query Query) {
	fm.query = query

	// since many runes will be looked for the same query,
	// we eagerly revolve the candidates for the given query
	fm.buildCandidates()
}

func (cd *candidates) resetWithSize(L int) {
	if cap(cd.withFallback) < L { // reallocate
		cd.withFallback = make([][]int, L)
		cd.withoutFallback = make([]int, L)
	}
	// only reslice
	cd.withFallback = cd.withFallback[0:L]
	cd.withoutFallback = cd.withoutFallback[0:L]

	// reset to "zero" values
	for i := range cd.withoutFallback {
		cd.withFallback[i] = nil
		cd.withoutFallback[i] = -1
	}
}

func (fm *FontMap) buildCandidates() {
	fm.lastFootprintIndex = -1
	fm.candidates.resetWithSize(len(fm.query.Families))

	selectFootprints := func(systemFallback bool) {
		for familyIndex, family := range fm.query.Families {
			candidates := fm.database.selectByFamily(family, systemFallback, &fm.footprintsBuffer, fm.cribleBuffer)
			if len(candidates) == 0 {
				continue
			}

			// select the correct aspects
			candidates = fm.database.retainsBestMatches(candidates, fm.query.Aspect)

			if systemFallback {
				fm.candidates.withFallback[familyIndex] = candidates
			} else {
				// when no systemFallback is required, the CSS spec says
				// that only one font among the candidates must be tried
				fm.candidates.withoutFallback[familyIndex] = candidates[0]
			}
		}
	}

	selectFootprints(false)
	selectFootprints(true)
}

// candidates is a cache storing the indices into FontMap.database of footprints matching a Query
// the two slices has the same length: the number of family in the query
type candidates struct {
	withFallback    [][]int // for each queried family
	withoutFallback []int   // for each queried family, only one footprint is selected
}

// returns nil if not candidates supports the rune `r`
func (fm *FontMap) resolveForRune(candidates []int, r rune) (font.Face, meta.Description) {
	// we first look up for an exact family match, without substitutions
	for _, footprintIndex := range candidates {
		// check the coverage
		if fp := fm.database[footprintIndex]; fp.Runes.Contains(r) {
			// try to use the font
			face, err := fm.loadFont(fp)
			if err != nil { // very unlikely; try an other family
				fm.logger.Println(err)
				continue
			}

			// register the face used
			fm.lastFootprintIndex = footprintIndex

			return face, fp.metadata()
		}
	}

	return nil, meta.Description{}
}

// ResolveFace select a font based on the current query (see `SetQuery`),
// and supporting the given rune, applying CSS font selection rules.
// The function will return nil if the underlying font database is empty,
// or if the file system is broken; otherwise the returned [font.Font] is always valid.
func (fm *FontMap) ResolveFace(r rune) font.Face {
	face, _ := fm.ResolveFaceAndMetadata(r)
	return face
}

// ResolveFaceAndMetadata select a font based on the current query (see `SetQuery`),
// and supporting the given rune, applying CSS font selection rules.
// The function will return nil if the underlying font database is empty,
// or if the file system is broken; otherwise the returned [font.Font] is always valid.
// The returned metadata describes the selected face.
func (fm *FontMap) ResolveFaceAndMetadata(r rune) (font.Face, meta.Description) {
	// in many case, the same font will support a lot of runes
	// thus, as an optimisation, we register the last used footprint and start
	// to check if it supports `r`
	if fm.lastFootprintIndex != -1 {
		// check the coverage
		if fp := fm.database[fm.lastFootprintIndex]; fp.Runes.Contains(r) {
			// try to use the font
			face, err := fm.loadFont(fp)
			if err == nil {
				return face, fp.metadata()
			}

			// very unlikely; warn and keep going
			fm.logger.Println(err)
		}
	}

	// we first look up for an exact family match, without substitutions
	for _, footprintIndex := range fm.candidates.withoutFallback {
		if footprintIndex == -1 {
			continue
		}
		if face, md := fm.resolveForRune([]int{footprintIndex}, r); face != nil {
			return face, md
		}
	}

	// if no family has matched so far, try again with system fallback
	for _, footprintIndexList := range fm.candidates.withFallback {
		if face, md := fm.resolveForRune(footprintIndexList, r); face != nil {
			return face, md
		}
	}

	// this is very very unlikely, since the substitution
	// always add a default generic family
	fm.logger.Printf("No font matched for %v -> returning arbitrary face", fm.query.Families)

	// return an arbitrary face
	for _, face := range fm.faces {
		return face.Face, face.Description
	}
	for _, fp := range fm.database {
		face, err := fm.loadFont(fp)
		if err != nil { // very unlikely
			continue
		}
		return face, fp.metadata()
	}

	// refreshSystemFontsIndex makes sure at least one face is valid
	// and AddFont also check for valid font files, meaning that
	// a valid FontMap should always contain a valid face,
	// and this should never happen in pratice
	return nil, meta.Description{}
}

func (fm *FontMap) loadFont(fp footprint) (font.Face, error) {
	if face, hasCached := fm.faces[fp.Location]; hasCached {
		return face.Face, nil
	}

	// since user provided fonts are added to `fonts`
	// we may now assume the font is stored on the file system
	face, err := fp.loadFromDisk()
	if err != nil {
		return nil, err
	}

	// add the face to the cache
	fm.faces[fp.Location] = cacheEntry{Face: face, Description: fp.metadata()}

	return face, nil
}
