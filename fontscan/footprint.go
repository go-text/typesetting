package fontscan

import (
	"fmt"
	"os"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/opentype/api"
	meta "github.com/go-text/typesetting/opentype/api/metadata"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

// Location identifies where a font.Face is stored.
type Location = api.FontID

// footprint is a condensed summary of the main information
// about a font, serving as a lightweight surrogate
// for the original font file.
type footprint struct {
	// Location stores the adress of the font resource.
	Location Location

	// Family is the general nature of the font, like
	// "Arial"
	Family string

	// Runes is the set of runes supported by the font.
	Runes runeSet

	// Aspect precises the visual characteristics
	// of the font among a family, like "Bold Italic"
	Aspect meta.Aspect
}

func newFootprintFromLoader(ld *loader.Loader) (out footprint, err error) {
	raw, err := ld.RawTable(loader.MustNewTag("cmap"))
	if err != nil {
		return footprint{}, err
	}
	tb, _, err := tables.ParseCmap(raw)
	if err != nil {
		return footprint{}, err
	}
	cmap, _, err := api.ProcessCmap(tb)
	if err != nil {
		return footprint{}, err
	}
	out.Runes = newRuneSetFromCmap(cmap) // ... and build the corresponding rune set

	me := meta.NewFontDescriptor(ld)
	out.Family = meta.NormalizeFamily(me.Family())
	out.Aspect = me.Aspect()

	return out, nil
}

// loadFromDisk assume the footprint location refers to the file system
func (fp *footprint) loadFromDisk() (font.Face, error) {
	location := fp.Location

	file, err := os.Open(location.File)
	if err != nil {
		return nil, err
	}

	faces, err := font.ParseTTC(file)
	if err != nil {
		return nil, err
	}

	if index := int(location.Index); len(faces) <= index {
		// this should only happen if the font file as changed
		// since the last scan (very unlikely)
		return nil, fmt.Errorf("invalid font index in collection: %d >= %d", index, len(faces))
	}

	return faces[location.Index], nil
}
