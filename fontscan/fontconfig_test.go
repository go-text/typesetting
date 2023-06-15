package fontscan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tu "github.com/go-text/typesetting/opentype/testutils"
)

func TestParseFontconfig(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/xdg")
	cwd, err := os.Getwd()
	tu.AssertNoErr(t, err)

	dirs, includes, err := parseFcFile("fontconfig_test/fonts.conf", cwd)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(dirs) == 4)
	tu.Assert(t, len(includes) == 1)

	dirs, includes, err = parseFcDir("fontconfig_test/conf.d", cwd, map[string]bool{})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(dirs) == 1)
	tu.Assert(t, len(includes) == 2)

	dirs, err = parseFcConfig("fontconfig_test/fonts.conf")
	tu.AssertNoErr(t, err)
	tu.Assert(t, reflect.DeepEqual(dirs, []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
		"/xdg/fonts",
		"~/.fonts",
		"my_Custom_Font_Dir",
		"fontconfig_test/conf.d/relative_font_dir",
		filepath.Join(cwd, "cwd_font_dir"),
	}))
}

func TestParseFontconfigErrors(t *testing.T) {
	_, _, err := parseFcFile("fontconfig_test/invalid.conf", "")
	tu.Assert(t, err != nil)
}
