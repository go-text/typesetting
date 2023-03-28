package metadata

import (
	"github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/opentype/loader"
	"github.com/go-text/typesetting/opentype/tables"
)

const (
	_ tables.NameID = iota
	nameFontFamily
	nameFontSubfamily
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	namePreferredFamily    // or Typographic Family
	namePreferredSubfamily // or Typographic Subfamily
	_
	_
	_
	nameWWSFamily
	nameWWSSubfamily
)

type fontDescriptor struct {
	// these tables are required both in Family
	// and Aspect
	os2   *tables.Os2 // optional
	names tables.Name
	head  tables.Head
}

func newFontDescriptor(ld *loader.Loader) *fontDescriptor {
	var out fontDescriptor

	// load tables, all considered optional
	raw, _ := ld.RawTable(loader.MustNewTag("OS/2"))
	if os2, _, err := tables.ParseOs2(raw); err != nil {
		out.os2 = &os2
	}

	raw, _ = ld.RawTable(loader.MustNewTag("name"))
	out.names, _, _ = tables.ParseName(raw)

	out.head, _ = font.LoadHeadTable(ld)

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

// Metadata queries the family and the aspect properties of the
// font loaded under [font]
func Metadata(font *loader.Loader) (aspect Aspect, family string) {
	descriptor := newFontDescriptor(font)

	aspect = descriptor.aspect()
	family = descriptor.family()

	return
}
