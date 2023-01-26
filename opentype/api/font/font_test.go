package font

import (
	"testing"

	"github.com/go-text/typesetting/opentype/api"
	tu "github.com/go-text/typesetting/opentype/testutils"
)

func TestCrashes(t *testing.T) {
	for _, filepath := range append(tu.Filenames(t, "common"), "toys/chromacheck-svg.ttf") {
		loadFont(t, filepath)
	}
}

func TestGlyphName(t *testing.T) {
	ft := loadFont(t, "toys/NamesCFF.ttf")
	tu.Assert(t, ft.post.names == nil)

	expected := [20]string{
		".notdef", "space", "uni0622", "uni0623", "uni0624", "uni0625", "uni0626", "uni0628",
		"uni06C0", "uni06C2", "uni06D3", "uni0625.fina", "uni0623.fina", "uni0622.fina", "uni0628.fina",
		"uni0626.init", "uni0628.init", "uni0626.medi", "uni0628.medi", "uni06C1.fina",
	}
	for i, exp := range expected {
		tu.Assert(t, ft.GlyphName(api.GID(i)) == exp)
	}
}

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, filepath := range tu.Filenames(b, "common") {
			loadFont(b, filepath)
		}
	}
}
