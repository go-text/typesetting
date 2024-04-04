package fontscan

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	tu "github.com/go-text/typesetting/testutils"
)

func TestParseFontconfig(t *testing.T) {
	cwd, err := os.Getwd()
	tu.AssertNoErr(t, err)
	fc := fcVars{
		xdgDataHome:   filepath.Clean("/xdgData"),
		xdgConfigHome: filepath.Clean("/xdgConfig"),
		userHome:      filepath.Clean("/home/me"),
		configFile:    "fonts.conf",
		paths:         []string{filepath.Join(cwd, "fontconfig_test")},
		sysroot:       "",
	}
	logger := log.New(io.Discard, "", 0)

	dirs, includes, err := fc.parseFcFile(logger, "fontconfig_test/fonts.conf", cwd)
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(dirs) == 4)
	tu.Assert(t, len(includes) == 1)

	dirs, includes, err = fc.parseFcDir(logger, "fontconfig_test/conf.d", cwd, map[string]bool{})
	tu.AssertNoErr(t, err)
	tu.Assert(t, len(dirs) == 1)
	tu.Assert(t, len(includes) == 2)

	dirs, err = fc.parseFcConfig(logger)
	for i, s := range dirs {
		dirs[i] = filepath.ToSlash(s)
	}
	sort.Strings(dirs)
	tu.AssertNoErr(t, err)
	expected := []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
		"/xdgData/fonts",
		"~/.fonts",
		"my_Custom_Font_Dir",
		filepath.Join(cwd, "fontconfig_test/conf.d/relative_font_dir"),
		filepath.Join(cwd, "cwd_font_dir"),
	}
	for i, s := range expected {
		expected[i] = filepath.ToSlash(s)
	}
	sort.Strings(expected)
	if !reflect.DeepEqual(expected, dirs) {
		t.Errorf("expected %q\ngot %q", expected, dirs)
	}
}

func TestParseFontconfigErrors(t *testing.T) {
	fc := fcVars{
		xdgDataHome:   "/xdgData",
		xdgConfigHome: "/xdgConfig",
		userHome:      "",
		configFile:    "fonts.conf",
		paths:         []string{""},
		sysroot:       "",
	}

	logger := log.New(io.Discard, "", 0)
	_, _, err := fc.parseFcFile(logger, "fontconfig_test/invalid.conf", "")
	tu.Assert(t, err != nil)
}
