package metadata

import (
	"github.com/go-text/typesetting/opentype/api"
	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

const (
	nameFontFamily         tables.NameID = 1
	nameFontSubfamily      tables.NameID = 2
	namePreferredFamily    tables.NameID = 16 // or Typographic Family
	namePreferredSubfamily tables.NameID = 17 // or Typographic Subfamily
	nameWWSFamily          tables.NameID = 21 //
	nameWWSSubfamily       tables.NameID = 22 //
)

type fontDescriptor struct {
	// these tables are required both in Family
	// and Aspect
	os2   *tables.Os2 // optional
	names tables.Name
	head  tables.Head

	cmap    api.Cmap // optional
	metrics tables.Hmtx
	post    tables.Post
}

func newFontDescriptor(ld *loader.Loader) *fontDescriptor {
	var out fontDescriptor

	// load tables, all considered optional
	raw, _ := ld.RawTable(loader.MustNewTag("OS/2"))
	fp := tables.FPNone
	if os2, _, err := tables.ParseOs2(raw); err != nil {
		out.os2 = &os2
		fp = os2.FontPage()
	}

	raw, _ = ld.RawTable(loader.MustNewTag("name"))
	out.names, _, _ = tables.ParseName(raw)

	out.head, _ = font.LoadHeadTable(ld)

	raw, _ = ld.RawTable(loader.MustNewTag("cmap"))
	tb, _, _ := tables.ParseCmap(raw)
	out.cmap, _, _ = api.ProcessCmap(tb, fp)

	raw, _ = ld.RawTable(loader.MustNewTag("name"))
	out.names, _, _ = tables.ParseName(raw)

	raw, _ = ld.RawTable(loader.MustNewTag("post"))
	out.post, _, _ = tables.ParsePost(raw)

	raw, _ = ld.RawTable(loader.MustNewTag("maxp"))
	maxp, _, _ := tables.ParseMaxp(raw)
	_, out.metrics, _ = font.LoadHmtx(ld, int(maxp.NumGlyphs))

	return &out
}

func (fd *fontDescriptor) family() string {
	var family string
	if fd.os2 != nil && fd.os2.FsSelection&256 != 0 {
		family = fd.names.Name(namePreferredFamily)
		if family == "" {
			family = fd.names.Name(nameFontFamily)
		}
	} else {
		family = fd.names.Name(nameWWSFamily)
		if family == "" {
			family = fd.names.Name(namePreferredFamily)
		}
		if family == "" {
			family = fd.names.Name(nameFontFamily)
		}
	}
	return family
}

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

	// try the fast shortcuts
	if fd.post.IsFixedPitch != 0 {
		return true
	}

	if fd.cmap == nil || fd.metrics.IsEmpty() {
		// we can't be sure, so be conservative
		return false
	}

	if len(fd.metrics.Metrics) == 1 {
		return true
	}

	// directly read the advances in the 'hmtx' table
	var firstAdvance int
	for gid, metric := range fd.metrics.Metrics {
		if gid == 0 { // ignore the 'unset' glyph, which may be different
			continue
		}
		advance := int(metric.AdvanceWidth)
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

// Description provides font metadata.
type Description struct {
	Family      string
	Aspect      Aspect
	IsMonospace bool
}

// Metadata queries the family and the aspect properties of the
// font loaded under [font]
func Metadata(font *loader.Loader) Description {
	descriptor := newFontDescriptor(font)

	var out Description
	out.Aspect = descriptor.aspect()
	out.Family = descriptor.family()
	out.IsMonospace = descriptor.isMonospace()

	return out
}
