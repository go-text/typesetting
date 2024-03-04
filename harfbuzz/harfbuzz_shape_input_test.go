package harfbuzz

import (
	"flag"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"testing"

	td "github.com/go-text/typesetting-utils/harfbuzz"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	tu "github.com/go-text/typesetting/testutils"
)

// collectTests walk through the tests directories, parsing .tests files
func collectTests(t testing.TB) []testData {
	disabledTests := map[string]struct{}{
		// requires proprietary fonts from the system (see the file)
		// enabling this tests requires to use a local replace of typesetting-utils
		"harfbuzz_reference/in-house/tests/macos.tests": {},

		// already handled in emojis_test.go
		"harfbuzz_reference/in-house/tests/emoji-clusters.tests": {},

		// disabled by harfbuzz (see harfbuzz/test/shaping/data/text-rendering-tests/DISABLED)
		"harfbuzz_reference/text-rendering-tests/tests/CMAP-3.tests":    {},
		"harfbuzz_reference/text-rendering-tests/tests/SHARAN-1.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHBALI-1.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHBALI-2.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHKNDA-2.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHKNDA-3.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-1.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-10.tests": {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-2.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-3.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-4.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-5.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-6.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-7.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-8.tests":  {},
		"harfbuzz_reference/text-rendering-tests/tests/SHLANA-9.tests":  {},
	}

	var testFiles []string
	testFiles = append(testFiles, tu.FilenamesFS(t, &td.Files, "harfbuzz_reference/aots/tests")...)
	testFiles = append(testFiles, tu.FilenamesFS(t, &td.Files, "harfbuzz_reference/in-house/tests")...)
	testFiles = append(testFiles, tu.FilenamesFS(t, &td.Files, "harfbuzz_reference/text-rendering-tests/tests")...)

	var allTests []testData
	for _, file := range testFiles {
		if _, isDisabled := disabledTests[file]; isDisabled {
			continue
		}

		allTests = append(allTests, readTestFile(t, file)...)
	}
	return allTests
}

// opens and parses a "xxx.tests" test file
func readTestFile(t testing.TB, filename string) (out []testData) {
	f, err := td.Files.ReadFile(filename)
	tu.AssertNoErr(t, err)

	// We can't use filepath.Dir here because embed.FS always uses unix-style paths, even on windows.
	dir := path.Dir(filename) // xxx/tests/file.tests

	for _, line := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" { // skip comments
			continue
		}

		// special case
		if strings.Contains(line, "--shaper=fallback") {
			// we do not support fallback shaper
			continue
		}

		out = append(out, newTestData(t, dir, line))
	}

	return out
}

// testData represents one line of a .tests file
type testData struct {
	// the test line yielding this test
	originDir  string
	originLine string

	input    testInput
	expected string
}

func newTestData(t testing.TB, dir string, line string) testData {
	chunks := strings.Split(line, ";")
	tu.Assert(t, len(chunks) == 4)

	fontFileHash, options, unicodes, expected := chunks[0], chunks[1], chunks[2], chunks[3]

	splitHash := strings.Split(fontFileHash, "@")
	// We should not use filepath.Join here because embed.FS expects unix-style paths always.
	fontFile := path.Join(dir, splitHash[0])
	// skip costy hash check, trusting upstream repo

	input := newTestInput(t, options)
	input.fontOpts.fontRef.File = fontFile

	input.text = parseUnicodes(t, unicodes)

	return testData{originDir: dir, originLine: line, input: input, expected: expected}
}

// Step 1 - parse the input test string
type testInput struct {
	text                  []rune
	textBefore, textAfter []rune

	shaper   shapeOpts
	fontOpts fontOpts
	format   formatOpts
}

// newTestInput parses the options, written in command line format
func newTestInput(t testing.TB, options string) testInput {
	flags := flag.NewFlagSet("options", flag.ContinueOnError)

	var fmtOpts formatOpts
	flags.BoolVar(&fmtOpts.hideClusters, "no-clusters", false, "Do not output cluster indices")
	flags.BoolVar(&fmtOpts.hideGlyphNames, "no-glyph-names", false, "Output glyph indices instead of names")
	flags.BoolVar(&fmtOpts.hidePositions, "no-positions", false, "Do not output glyph positions")
	flags.BoolVar(&fmtOpts.hideAdvances, "no-advances", false, "Do not output glyph advances")
	flags.BoolVar(&fmtOpts.showExtents, "show-extents", false, "Output glyph extents")
	flags.BoolVar(&fmtOpts.showFlags, "show-flags", false, "Output glyph flags")

	ned := flags.Bool("ned", false, "No Extra Data; Do not output clusters or advances")

	var so shapeOpts
	flags.StringVar(&so.features, "features", "", featuresUsage)

	flags.String("list-shapers", "", "(ignored)")
	flags.StringVar(&so.shaper, "shaper", "", "Force a shaper")
	flags.String("shapers", "", "(ignored)")
	flags.Func("direction", "Set text direction (default: auto)", so.parseDirection)
	flags.Func("language", "Set text language (default: $LANG)", func(s string) error {
		so.props.Language = language.NewLanguage(s)
		return nil
	})
	flags.Func("script", "Set text script, as an ISO-15924 tag (default: auto)", func(s string) error {
		var err error
		so.props.Script, err = language.ParseScript(s)
		return err
	})
	flags.BoolVar(&so.bot, "bot", false, "Treat text as beginning-of-paragraph")
	flags.BoolVar(&so.eot, "eot", false, "Treat text as end-of-paragraph")
	flags.BoolVar(&so.removeDefaultIgnorables, "remove-default-ignorables", false, "Remove Default-Ignorable characters")
	flags.BoolVar(&so.preserveDefaultIgnorables, "preserve-default-ignorables", false, "Preserve Default-Ignorable characters")
	flags.Func("cluster-level", "Cluster merging level (0/1/2, default: 0)", func(s string) error {
		l, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid cluster-level option: %s", err)
		}
		if l < 0 || l > 2 {
			return fmt.Errorf("invalid cluster-level option : %d", l)
		}
		so.clusterLevel = ClusterLevel(l)
		return nil
	})
	flags.BoolVar(&so.unsafeToConcat, "unsafe-to-concat", false, "Produce unsafe-to-concat glyph flag")
	flags.BoolVar(&so.safeToInsertTatweel, "safe-to-insert-tatweel", false, "Produce safe-to-insert-tatweel glyph flag")

	fo := newFontOptions()

	fontRefIndex := flags.Int("face-index", 0, "Set face index (default: 0)")
	flags.Func("font-size", "Font size", fo.parseFontSize)
	flags.Func("font-ppem", "Set x,y pixels per EM (default: 0; disabled)", fo.parseFontPpem)
	flags.Float64Var(&fo.ptem, "font-ptem", 0, "Set font point-size (default: 0; disabled)")
	flags.Func("variations", variationsUsage, fo.parseVariations)
	flags.String("font-funcs", "", "(ignored)")
	flags.String("ft-load-flags", "", "(ignored)")

	ub := flags.String("unicodes-before", "", "Set Unicode codepoints context before each line")
	ua := flags.String("unicodes-after", "", "Set Unicode codepoints context after each line")

	err := flags.Parse(strings.Split(options, " "))
	tu.AssertNoErr(t, err)

	if *ned {
		fmtOpts.hideClusters = true
		fmtOpts.hideAdvances = true
	}
	fo.fontRef.Index = uint16(*fontRefIndex)
	out := testInput{
		fontOpts: fo,
		format:   fmtOpts,
		shaper:   so,
	}

	if *ub != "" {
		out.textBefore = parseUnicodes(t, *ub)
	}
	if *ua != "" {
		out.textAfter = parseUnicodes(t, *ua)
	}

	return out
}

// how to print the output, which is what is compared
type formatOpts struct {
	hideGlyphNames bool
	hidePositions  bool
	hideAdvances   bool
	hideClusters   bool
	showExtents    bool
	showFlags      bool
}

type shapeOpts struct {
	shaper                    string
	features                  string
	props                     SegmentProperties
	invisibleGlyph            GID
	clusterLevel              ClusterLevel
	bot                       bool
	eot                       bool
	preserveDefaultIgnorables bool
	removeDefaultIgnorables   bool
	unsafeToConcat            bool
	safeToInsertTatweel       bool
}

func (opts *shapeOpts) parseDirection(s string) error {
	switch toLower(s[0]) {
	case 'l':
		opts.props.Direction = LeftToRight
	case 'r':
		opts.props.Direction = RightToLeft
	case 't':
		opts.props.Direction = TopToBottom
	case 'b':
		opts.props.Direction = BottomToTop
	default:
		return fmt.Errorf("invalid direction %s", s)
	}
	return nil
}

func parseUnicodes(t testing.TB, s string) []rune {
	runes := strings.Split(s, ",")
	text := make([]rune, len(runes))
	for i, r := range runes {
		if _, err := fmt.Sscanf(r, "U+%x", &text[i]); err == nil {
			continue
		}
		if _, err := fmt.Sscanf(r, "0x%x", &text[i]); err == nil {
			continue
		}
		if _, err := fmt.Sscanf(r, "%x", &text[i]); err == nil {
			continue
		}

		t.Fatalf("invalid unicode rune : %s", r)
	}
	return text
}

const featuresUsage = `Comma-separated list of font features

    Features can be enabled or disabled, either globally or limited to
    specific character ranges.  The format for specifying feature settings
    follows.  All valid CSS font-feature-settings values other than 'normal'
    and the global values are also accepted, though not documented below.
    CSS string escapes are not supported.

    The range indices refer to the positions between Unicode characters,
    unless the --utf8-clusters is provided, in which case range indices
    refer to UTF-8 byte indices. The position before the first character
    is always 0.

    The format is Python-esque.  Here is how it all works:

      Syntax:       Value:    Start:    End:

    Setting value:
      "kern"        1         0         ∞         // Turn feature on
      "+kern"       1         0         ∞         // Turn feature on
      "-kern"       0         0         ∞         // Turn feature off
      "kern=0"      0         0         ∞         // Turn feature off
      "kern=1"      1         0         ∞         // Turn feature on
      "aalt=2"      2         0         ∞         // Choose 2nd alternate

    Setting index:
      "kern[]"      1         0         ∞         // Turn feature on
      "kern[:]"     1         0         ∞         // Turn feature on
      "kern[5:]"    1         5         ∞         // Turn feature on, partial
      "kern[:5]"    1         0         5         // Turn feature on, partial
      "kern[3:5]"   1         3         5         // Turn feature on, range
      "kern[3]"     1         3         3+1       // Turn feature on, single char

    Mixing it all:

      "aalt[3:5]=2" 2         3         5         // Turn 2nd alternate on for range
`

func (opts *shapeOpts) parseFeatures() ([]Feature, error) {
	if opts.features == "" {
		return nil, nil
	}
	// remove possible quote
	s := strings.Trim(opts.features, `"`)

	features := strings.Split(s, ",")
	out := make([]Feature, len(features))

	var err error
	for i, feature := range features {
		out[i], err = ParseFeature(feature)
		if err != nil {
			return nil, fmt.Errorf("parsing features %s: %s", opts.features, err)
		}
	}
	return out, nil
}

type fontOpts struct {
	fontRef    font.FontID
	variations []font.Variation

	subpixelBits         int
	fontSizeX, fontSizeY int
	ptem                 float64
	yPpem, xPpem         uint16
}

const fontSizeUpem = 0x7FFFFFFF

func newFontOptions() fontOpts {
	return fontOpts{
		subpixelBits: 0,
		fontSizeX:    fontSizeUpem,
		fontSizeY:    fontSizeUpem,
	}
}

func (opts *fontOpts) parseFontSize(arg string) error {
	if arg == "upem" {
		opts.fontSizeY = fontSizeUpem
		opts.fontSizeX = fontSizeUpem
		return nil
	}
	n, err := fmt.Sscanf(arg, "%d %d", &opts.fontSizeX, &opts.fontSizeY)
	if err != io.EOF {
		return fmt.Errorf("font-size argument should be one or two space-separated numbers")
	}
	if n == 1 {
		opts.fontSizeY = opts.fontSizeX
	}
	return nil
}

func (opts *fontOpts) parseFontPpem(arg string) error {
	n, err := fmt.Sscanf(arg, "%d %d", &opts.xPpem, &opts.yPpem)
	if err != io.EOF {
		return fmt.Errorf("font-ppem argument should be one or two space-separated integers")
	}
	if n == 1 {
		opts.yPpem = opts.xPpem
	}
	return nil
}

const variationsUsage = `Comma-separated list of font variations

    Variations are set globally. The format for specifying variation settings
    follows.  All valid CSS font-variation-settings values other than 'normal'
    and 'inherited' are also accepted, although not documented below.

    The format is a tag, optionally followed by an equals sign, followed by a
    number. For example:

      "wght=500"
      "slnt=-7.5";
`

// see variationsUsage
func (opts *fontOpts) parseVariations(s string) error {
	// remove possible quote
	s = strings.Trim(s, `"`)

	variations := strings.Split(s, ",")
	opts.variations = make([]font.Variation, len(variations))

	var err error
	for i, feature := range variations {
		opts.variations[i], err = ParseVariation(feature)
		if err != nil {
			return err
		}
	}
	return nil
}
