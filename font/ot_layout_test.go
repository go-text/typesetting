// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"reflect"
	"sort"
	"testing"

	"github.com/go-text/typesetting/font/opentype/tables"
	tu "github.com/go-text/typesetting/testutils"
)

func TestGetProps(t *testing.T) {
	file := readFontFile(t, "common/Raleway-v4020-Regular.otf")

	gpos, _, err := tables.ParseLayout(readTable(t, file, "GPOS"))
	tu.AssertNoErr(t, err)
	gsub, _, err := tables.ParseLayout(readTable(t, file, "GSUB"))
	tu.AssertNoErr(t, err)

	for _, table := range []Layout{newLayout(gpos), newLayout(gsub)} {
		var tags []int
		for _, s := range table.Scripts {
			tags = append(tags, int(s.Tag))
		}
		tu.Assert(t, sort.IntsAreSorted(tags))

		for i, s := range table.Scripts {
			ptr := table.FindScript(s.Tag)
			tu.Assert(t, ptr == i)
		}

		s := table.FindScript(Tag(0)) // invalid
		tu.Assert(t, s == -1)

		for _, feat := range table.Features {
			_, ok := table.FindFeatureIndex(feat.Tag)
			tu.Assert(t, ok)
		}
		_, ok := table.FindFeatureIndex(Tag(0)) // invalid
		tu.Assert(t, !ok)

		// now check the languages

		for _, script := range table.Scripts {
			var tags []int
			for _, s := range script.LangSysRecords {
				tags = append(tags, int(s.Tag))
			}
			tu.Assert(t, sort.IntsAreSorted(tags))

			for i, l := range script.LangSysRecords {
				ptr := script.FindLanguage(l.Tag)
				tu.Assert(t, ptr == i)
			}

			s := script.FindLanguage(Tag(0)) // invalid
			tu.Assert(t, s == -1)

			tu.Assert(t, script.DefaultLangSys != nil)
			tu.Assert(t, reflect.DeepEqual(script.GetLangSys(0xFFFF), *script.DefaultLangSys))
		}
	}
}

func TestOTFeatureVariation(t *testing.T) {
	ft := readFontFile(t, "common/Commissioner-VF.ttf")

	gsubT, _, err := tables.ParseLayout(readTable(t, ft, "GSUB"))
	tu.AssertNoErr(t, err)

	gsub := newLayout(gsubT)
	tu.Assert(t, gsub.FindVariationIndex([]VarCoord{tables.NewCoord(0.8)}) == 0)
	tu.Assert(t, gsub.FindVariationIndex([]VarCoord{tables.NewCoord(0.4)}) == -1)
}
