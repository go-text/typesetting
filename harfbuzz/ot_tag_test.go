package harfbuzz

import (
	"testing"

	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	"github.com/go-text/typesetting/language"
)

// ported from harfbuzz/test/api/test-ot-tag.c Copyright © 2011  Google, Inc. Behdad Esfahbod

func assertEqualTag(t *testing.T, t1, t2 tables.Tag) {
	t.Helper()

	if t1 != t2 {
		t.Fatalf("unexpected %s != %s", t1, t2)
	}
}

/* https://docs.microsoft.com/en-us/typography/opentype/spec/scripttags */

func testSimpleTags(t *testing.T, s string, script language.Script) {
	tag := ot.MustNewTag(s)

	tags, _ := newOTTagsFromScriptAndLanguage(script, "")

	if len(tags) != 0 {
		assertEqualTag(t, tags[0], tag)
	} else {
		assertEqualTag(t, ot.MustNewTag("DFLT"), tag)
	}
}

func testScriptTagsFromLanguage(t *testing.T, s, langS string, script language.Script) {
	var tag tables.Tag
	if s != "" {
		tag = ot.MustNewTag(s)
	}

	tags, _ := newOTTagsFromScriptAndLanguage(script, language.NewLanguage(langS))
	if len(tags) != 0 {
		assertEqualInt(t, len(tags), 1)
		assertEqualTag(t, tags[0], tag)
	}
}

func testIndicTags(t *testing.T, s1, s2, s3 string, script language.Script) {
	tag1 := ot.MustNewTag(s1)
	tag2 := ot.MustNewTag(s2)
	tag3 := ot.MustNewTag(s3)

	tags, _ := newOTTagsFromScriptAndLanguage(script, "")

	assertEqualInt(t, len(tags), 3)
	assertEqualTag(t, tags[0], tag1)
	assertEqualTag(t, tags[1], tag2)
	assertEqualTag(t, tags[2], tag3)
}

func TestOtTagScriptDegenerate(t *testing.T) {
	assertEqualTag(t, ot.MustNewTag("DFLT"), tagDefaultScript)

	/* HIRAGANA and KATAKANA both map to 'kana' */
	testSimpleTags(t, "kana", language.Katakana)

	tags, _ := newOTTagsFromScriptAndLanguage(language.Hiragana, "")

	assertEqualInt(t, len(tags), 1)
	assertEqualTag(t, tags[0], ot.MustNewTag("kana"))

	testSimpleTags(t, "DFLT", 0)

	/* Spaces are replaced */
	// assertEqualInt(t, hb_ot_tag_to_script(ot.MustNewTag("be  ")), hb_script_from_string("Beee", -1))
}

func TestOtTagScriptSimple(t *testing.T) {
	/* Arbitrary non-existent script */
	// testSimpleTags(t, "wwyz", hb_script_from_string("wWyZ", -1))

	/* These we don't really care about */
	testSimpleTags(t, "zyyy", language.Common)
	testSimpleTags(t, "zinh", language.Inherited)
	testSimpleTags(t, "zzzz", language.Unknown)

	testSimpleTags(t, "arab", language.Arabic)
	testSimpleTags(t, "copt", language.Coptic)
	testSimpleTags(t, "kana", language.Katakana)
	testSimpleTags(t, "latn", language.Latin)

	testSimpleTags(t, "math", language.Mathematical_notation)

	/* These are trickier since their OT script tags have space. */
	testSimpleTags(t, "lao ", language.Lao)
	testSimpleTags(t, "yi  ", language.Yi)
	/* Unicode-5.0 additions */
	testSimpleTags(t, "nko ", language.Nko)
	/* Unicode-5.1 additions */
	testSimpleTags(t, "vai ", language.Vai)

	/* https://docs.microsoft.com/en-us/typography/opentype/spec/scripttags */

	/* Unicode-5.2 additions */
	testSimpleTags(t, "mtei", language.Meetei_Mayek)
	/* Unicode-6.0 additions */
	testSimpleTags(t, "mand", language.Mandaic)
}

func TestOtTagScriptFromLanguage(t *testing.T) {
	testScriptTagsFromLanguage(t, "", "", 0)
	testScriptTagsFromLanguage(t, "", "en", 0)
	testScriptTagsFromLanguage(t, "copt", "en", language.Coptic)
	testScriptTagsFromLanguage(t, "", "x-hbsc", 0)
	testScriptTagsFromLanguage(t, "copt", "x-hbsc", language.Coptic)
	testScriptTagsFromLanguage(t, "", "x-hbsc-", 0)
	testScriptTagsFromLanguage(t, "", "x-hbsc-1", 0)
	testScriptTagsFromLanguage(t, "", "x-hbsc-1a", 0)
	testScriptTagsFromLanguage(t, "", "x-hbsc-1a2b3c4x", 0)
	testScriptTagsFromLanguage(t, "2lon", "x-hbsc-326c6f6e67", 0)
	testScriptTagsFromLanguage(t, "abc ", "x-hbscabc", 0)
	testScriptTagsFromLanguage(t, "deva", "x-hbscdeva", 0)
	testScriptTagsFromLanguage(t, "dev2", "x-hbscdev2", 0)
	testScriptTagsFromLanguage(t, "dev3", "x-hbscdev3", 0)
	testScriptTagsFromLanguage(t, "dev3", "x-hbsc-64657633", 0)
	testScriptTagsFromLanguage(t, "copt", "x-hbotpap0-hbsccopt", 0)
	testScriptTagsFromLanguage(t, "", "en-x-hbsc", 0)
	testScriptTagsFromLanguage(t, "copt", "en-x-hbsc", language.Coptic)
	testScriptTagsFromLanguage(t, "abc ", "en-x-hbscabc", 0)
	testScriptTagsFromLanguage(t, "deva", "en-x-hbscdeva", 0)
	testScriptTagsFromLanguage(t, "dev2", "en-x-hbscdev2", 0)
	testScriptTagsFromLanguage(t, "dev3", "en-x-hbscdev3", 0)
	testScriptTagsFromLanguage(t, "dev3", "en-x-hbsc-64657633", 0)
	testScriptTagsFromLanguage(t, "copt", "en-x-hbotpap0-hbsccopt", 0)
	testScriptTagsFromLanguage(t, "", "UTF-8", 0)

	// corner cases should not panic
	testScriptTagsFromLanguage(t, "", "x", 0)
}

func TestOtTagScriptIndic(t *testing.T) {
	testIndicTags(t, "bng3", "bng2", "beng", language.Bengali)
	testIndicTags(t, "dev3", "dev2", "deva", language.Devanagari)
	testIndicTags(t, "gjr3", "gjr2", "gujr", language.Gujarati)
	testIndicTags(t, "gur3", "gur2", "guru", language.Gurmukhi)
	testIndicTags(t, "knd3", "knd2", "knda", language.Kannada)
	testIndicTags(t, "mlm3", "mlm2", "mlym", language.Malayalam)
	testIndicTags(t, "ory3", "ory2", "orya", language.Oriya)
	testIndicTags(t, "tml3", "tml2", "taml", language.Tamil)
	testIndicTags(t, "tel3", "tel2", "telu", language.Telugu)
}

/* https://docs.microsoft.com/en-us/typography/opentype/spec/languagetags */

func testLanguageTwoWay(t *testing.T, tagS, langS string) {
	lang := language.NewLanguage(langS)
	tag := ot.MustNewTag(tagS)

	_, tags := newOTTagsFromScriptAndLanguage(0, lang)

	if len(tags) != 0 {
		assertEqualTag(t, tag, tags[0])
	} else {
		assertEqualTag(t, tag, ot.MustNewTag("dflt"))
	}
}

func testTagFromLanguage(t *testing.T, tagS, langS string) {
	t.Helper()

	lang := language.NewLanguage(langS)
	tag := ot.MustNewTag(tagS)

	_, tags := newOTTagsFromScriptAndLanguage(0, lang)

	if len(tags) != 0 {
		assertEqualTag(t, tag, tags[0])
	} else {
		assertEqualTag(t, tag, ot.MustNewTag("dflt"))
	}
}

func TestOtTagLanguage(t *testing.T) {
	assertEqualInt(t, int(ot.MustNewTag("dflt")), int(tagDefaultLanguage))
	testLanguageTwoWay(t, "dflt", "")

	testLanguageTwoWay(t, "ALT ", "alt")

	testLanguageTwoWay(t, "ARA ", "ar")

	testLanguageTwoWay(t, "AZE ", "az")
	testTagFromLanguage(t, "AZE ", "az-ir")
	testTagFromLanguage(t, "AZE ", "az-az")

	testLanguageTwoWay(t, "ENG ", "en")
	testTagFromLanguage(t, "ENG ", "en_US")

	testLanguageTwoWay(t, "CJA ", "cja-x-hbot-434a4120") /* Western Cham */
	testLanguageTwoWay(t, "CJM ", "cjm-x-hbot-434a4d20") /* Eastern Cham */
	testTagFromLanguage(t, "CJM ", "cjm")
	testLanguageTwoWay(t, "EVN ", "eve")

	testLanguageTwoWay(t, "HAL ", "cfm")  /* BCP47 and current ISO639-3 code for Halam/Falam Chin */
	testTagFromLanguage(t, "HAL ", "flm") /* Retired ISO639-3 code for Halam/Falam Chin */

	testLanguageTwoWay(t, "HYE0", "hy")
	testLanguageTwoWay(t, "HYE ", "hyw")

	testTagFromLanguage(t, "QIN ", "bgr") /* Bawm Chin */
	testTagFromLanguage(t, "QIN ", "cbl") /* Bualkhaw Chin */
	testTagFromLanguage(t, "QIN ", "cka") /* Khumi Awa Chin */
	testTagFromLanguage(t, "QIN ", "cmr") /* Mro-Khimi Chin */
	testTagFromLanguage(t, "QIN ", "cnb") /* Chinbon Chin */
	testTagFromLanguage(t, "QIN ", "cnh") /* Hakha Chin */
	testTagFromLanguage(t, "QIN ", "cnk") /* Khumi Chin */
	testTagFromLanguage(t, "QIN ", "cnw") /* Ngawn Chin */
	testTagFromLanguage(t, "QIN ", "csh") /* Asho Chin */
	testTagFromLanguage(t, "QIN ", "csy") /* Siyin Chin */
	testTagFromLanguage(t, "QIN ", "ctd") /* Tedim Chin */
	testTagFromLanguage(t, "QIN ", "czt") /* Zotung Chin */
	testTagFromLanguage(t, "QIN ", "dao") /* Daai Chin */
	testTagFromLanguage(t, "QIN ", "hlt") /* Matu Chin */
	testTagFromLanguage(t, "QIN ", "mrh") /* Mara Chin */
	testTagFromLanguage(t, "QIN ", "pck") /* Paite Chin */
	testTagFromLanguage(t, "QIN ", "sez") /* Senthang Chin */
	testTagFromLanguage(t, "QIN ", "tcp") /* Tawr Chin */
	testTagFromLanguage(t, "QIN ", "tcz") /* Thado Chin */
	testTagFromLanguage(t, "QIN ", "yos") /* Yos, deprecated by IANA in favor of Zou [zom] */
	testTagFromLanguage(t, "QIN ", "zom") /* Zou */
	// test_tag_to_language("QIN", "bgr")    /* no single BCP47 tag for Chin; picking Bawm Chin */

	testLanguageTwoWay(t, "FAR ", "fa")
	testTagFromLanguage(t, "FAR ", "fa_IR")

	testLanguageTwoWay(t, "MNK ", "man") /* Mandingo [macrolanguage] */

	testLanguageTwoWay(t, "SWA ", "aii") /* Swadaya Aramaic */

	testLanguageTwoWay(t, "SYR ", "syr")  /* Syriac [macrolanguage] */
	testTagFromLanguage(t, "SYR ", "amw") /* Western Neo-Aramaic */
	testTagFromLanguage(t, "SYR ", "cld") /* Chaldean Neo-Aramaic */
	testTagFromLanguage(t, "SYR ", "syc") /* Classical Syriac */

	testLanguageTwoWay(t, "TUA ", "tru") /* Turoyo Aramaic */

	testTagFromLanguage(t, "ZHS ", "zh")         /* Chinese */
	testTagFromLanguage(t, "ZHS ", "zh-cn")      /* Chinese (China) */
	testTagFromLanguage(t, "ZHS ", "zh-sg")      /* Chinese (Singapore) */
	testTagFromLanguage(t, "ZHTM", "zh-mo")      /* Chinese (Macao) */
	testTagFromLanguage(t, "ZHTM", "zh-hant-mo") /* Chinese (Macao) */
	testTagFromLanguage(t, "ZHS ", "zh-hans-mo") /* Chinese (Simplified, Macao) */
	testLanguageTwoWay(t, "ZHH ", "zh-HK")       /* Chinese (Hong Kong) */
	testTagFromLanguage(t, "ZHH ", "zH-HanT-hK") /* Chinese (Hong Kong) */
	testTagFromLanguage(t, "ZHS ", "zH-HanS-hK") /* Chinese (Simplified, Hong Kong) */
	testTagFromLanguage(t, "ZHT ", "zh-tw")      /* Chinese (Taiwan) */
	testLanguageTwoWay(t, "ZHS ", "zh-Hans")     /* Chinese (Simplified) */
	testLanguageTwoWay(t, "ZHT ", "zh-Hant")     /* Chinese (Traditional) */
	testTagFromLanguage(t, "ZHS ", "zh-xx")      /* Chinese (Other) */

	testTagFromLanguage(t, "ZHS ", "zh-Hans-TW")

	testTagFromLanguage(t, "ZHH ", "yue")
	testTagFromLanguage(t, "ZHH ", "yue-Hant")
	testTagFromLanguage(t, "ZHS ", "yue-Hans")

	testLanguageTwoWay(t, "ABC ", "abc-x-hbot-41424320")
	testLanguageTwoWay(t, "ABCD", "x-hbot-41424344")
	testTagFromLanguage(t, "ABC ", "asdf-asdf-wer-x-hbotabc-zxc")
	testTagFromLanguage(t, "ABC ", "asdf-asdf-wer-x-hbotabc")
	testTagFromLanguage(t, "ABCD", "asdf-asdf-wer-x-hbotabcd")
	testTagFromLanguage(t, "ABC ", "asdf-asdf-wer-x-hbot-41424320-zxc")
	testTagFromLanguage(t, "ABC ", "asdf-asdf-wer-x-hbot-41424320")
	testTagFromLanguage(t, "ABCD", "asdf-asdf-wer-x-hbot-41424344")

	testTagFromLanguage(t, "dflt", "asdf-asdf-wer-x-hbot")
	testTagFromLanguage(t, "dflt", "asdf-asdf-wer-x-hbot-zxc")
	testTagFromLanguage(t, "dflt", "asdf-asdf-wer-x-hbot-zxc-414243")
	testTagFromLanguage(t, "dflt", "asdf-asdf-wer-x-hbot-414243")
	testTagFromLanguage(t, "dflt", "asdf-asdf-wer-x-hbot-4142432")

	testTagFromLanguage(t, "dflt", "xy")
	testTagFromLanguage(t, "XYZ ", "xyz")    /* Unknown ISO 639-3 */
	testTagFromLanguage(t, "XYZ ", "xyz-qw") /* Unknown ISO 639-3 */

	/*
	* Invalid input. The precise answer does not matter, as long as it
	* does not crash or get into an infinite loop.
	 */
	testTagFromLanguage(t, "IPPH", "-fonipa")

	/*
	* Tags that contain "-fonipa" as a substring but which do not contain
	* the subtag "fonipa".
	 */
	testTagFromLanguage(t, "ENG ", "en-fonipax")
	testTagFromLanguage(t, "ENG ", "en-x-fonipa")
	testTagFromLanguage(t, "ENG ", "en-a-fonipa")
	testTagFromLanguage(t, "ENG ", "en-a-qwe-b-fonipa")

	/* International Phonetic Alphabet */
	testTagFromLanguage(t, "IPPH", "en-fonipa")
	testTagFromLanguage(t, "IPPH", "en-fonipax-fonipa")
	testTagFromLanguage(t, "IPPH", "rm-CH-fonipa-sursilv-x-foobar")
	testLanguageTwoWay(t, "IPPH", "und-fonipa")
	testTagFromLanguage(t, "IPPH", "zh-fonipa")

	/* North American Phonetic Alphabet (Americanist Phonetic Notation) */
	testTagFromLanguage(t, "APPH", "en-fonnapa")
	testTagFromLanguage(t, "APPH", "chr-fonnapa")
	testLanguageTwoWay(t, "APPH", "und-fonnapa")

	/* Khutsuri Georgian */
	testTagFromLanguage(t, "KGE ", "ka-Geok")
	testLanguageTwoWay(t, "KGE ", "und-Geok")

	/* Irish Traditional */
	testLanguageTwoWay(t, "IRT ", "ga-Latg")

	/* Moldavian */
	testLanguageTwoWay(t, "MOL ", "ro-MD")

	/* Polytonic Greek */
	testLanguageTwoWay(t, "PGR ", "el-polyton")
	testTagFromLanguage(t, "PGR ", "el-CY-polyton")

	/* Estrangela Syriac */
	testTagFromLanguage(t, "SYRE", "aii-Syre")
	testTagFromLanguage(t, "SYRE", "de-Syre")
	testTagFromLanguage(t, "SYRE", "syr-Syre")
	testLanguageTwoWay(t, "SYRE", "und-Syre")

	/* Western Syriac */
	testTagFromLanguage(t, "SYRJ", "aii-Syrj")
	testTagFromLanguage(t, "SYRJ", "de-Syrj")
	testTagFromLanguage(t, "SYRJ", "syr-Syrj")
	testLanguageTwoWay(t, "SYRJ", "und-Syrj")

	/* Eastern Syriac */
	testTagFromLanguage(t, "SYRN", "aii-Syrn")
	testTagFromLanguage(t, "SYRN", "de-Syrn")
	testTagFromLanguage(t, "SYRN", "syr-Syrn")
	testLanguageTwoWay(t, "SYRN", "und-Syrn")

	/* Test that x-hbot overrides the base language */
	testTagFromLanguage(t, "ABC ", "fa-x-hbotabc-hbot-41686121-zxc")
	testTagFromLanguage(t, "ABC ", "fa-ir-x-hbotabc-hbot-41686121-zxc")
	testTagFromLanguage(t, "ABC ", "zh-x-hbotabc-hbot-41686121-zxc")
	testTagFromLanguage(t, "ABC ", "zh-cn-x-hbotabc-hbot-41686121-zxc")
	testTagFromLanguage(t, "ABC ", "zh-xy-x-hbotabc-hbot-41686121-zxc")
	testTagFromLanguage(t, "ABC ", "xyz-xy-x-hbotabc-hbot-41686121-zxc")

	testTagFromLanguage(t, "Aha!", "fa-x-hbot-41686121-hbotabc-zxc")
	testTagFromLanguage(t, "Aha!", "fa-ir-x-hbot-41686121-hbotabc-zxc")
	testTagFromLanguage(t, "Aha!", "zh-x-hbot-41686121-hbotabc-zxc")
	testTagFromLanguage(t, "Aha!", "zh-cn-x-hbot-41686121-hbotabc-zxc")
	testTagFromLanguage(t, "Aha!", "zh-xy-x-hbot-41686121-hbotabc-zxc")
	testTagFromLanguage(t, "Aha!", "xyz-xy-x-hbot-41686121-hbotabc-zxc")

	/* Invalid x-hbot */
	testTagFromLanguage(t, "dflt", "x-hbot")
	testTagFromLanguage(t, "dflt", "x-hbot-")
	testTagFromLanguage(t, "dflt", "x-hbot-1")
	testTagFromLanguage(t, "dflt", "x-hbot-1a")
	testTagFromLanguage(t, "dflt", "x-hbot-1a2b3c4x")
	testTagFromLanguage(t, "2lon", "x-hbot-326c6f6e67")

	/* Unnormalized BCP 47 tags */
	testTagFromLanguage(t, "ARA ", "ar-aao")
	testTagFromLanguage(t, "JBO ", "art-lojban")
	testTagFromLanguage(t, "KOK ", "kok-gom")
	testTagFromLanguage(t, "LTZ ", "i-lux")
	testTagFromLanguage(t, "MNG ", "drh")
	testTagFromLanguage(t, "MOR ", "ar-ary")
	testTagFromLanguage(t, "MOR ", "ar-ary-DZ")
	testTagFromLanguage(t, "NOR ", "no-bok")
	testTagFromLanguage(t, "NYN ", "no-nyn")
	testTagFromLanguage(t, "ZHS ", "i-hak")
	testTagFromLanguage(t, "ZHS ", "zh-guoyu")
	testTagFromLanguage(t, "ZHS ", "zh-min")
	testTagFromLanguage(t, "ZHS ", "zh-min-nan")
	testTagFromLanguage(t, "ZHS ", "zh-xiang")

	/* BCP 47 tags that look similar to unrelated language system tags */
	testTagFromLanguage(t, "SQI ", "als")
	testTagFromLanguage(t, "dflt", "far")

	/* A UN M.49 region code, not an extended language subtag */
	testTagFromLanguage(t, "ARA ", "ar-001")

	/* An invalid tag */
	testTagFromLanguage(t, "TRK ", "tr@foo=bar")
}

func testTags(t *testing.T, script language.Script, langS string, expectedScriptCount, expectedLanguageCount int, expected ...string) {
	lang := language.NewLanguage(langS)

	scriptTags, languageTags := newOTTagsFromScriptAndLanguage(script, lang)

	assertEqualInt(t, len(scriptTags), expectedScriptCount)
	assertEqualInt(t, len(languageTags), expectedLanguageCount)

	for i, s := range expected {
		expectedTag := ot.MustNewTag(s)
		var actualTag tables.Tag
		if i < expectedScriptCount {
			actualTag = scriptTags[i]
		} else {
			actualTag = languageTags[i-expectedScriptCount]
		}
		assertEqualTag(t, actualTag, expectedTag)
	}
}

func TestOtTagFull(t *testing.T) {
	testTags(t, 0, "en", 0, 1, "ENG ")
	testTags(t, 0, "en-x-hbscdflt", 1, 1, "DFLT", "ENG ")
	testTags(t, language.Latin, "en", 1, 1, "latn", "ENG ")
	testTags(t, 0, "und-fonnapa", 0, 1, "APPH")
	testTags(t, 0, "en-fonnapa", 0, 1, "APPH")
	testTags(t, 0, "x-hbot1234-hbsc5678", 1, 1, "5678", "1234")
	testTags(t, 0, "x-hbsc5678-hbot1234", 1, 1, "5678", "1234")
	testTags(t, language.Malayalam, "ml", 3, 2, "mlm3", "mlm2", "mlym", "MAL ", "MLR ")
	testTags(t, language.Myanmar, "und", 2, 1, "mym2", "mymr", "UND ")
	testTags(t, 0, "xyz", 0, 1, "XYZ ")
}

func TestOtTagFromLanguage(t *testing.T) {
	scs, _ := newOTTagsFromScriptAndLanguage(language.Tai_Tham, "")
	if len(scs) != 1 && scs[0] != 1818324577 {
		t.Fatalf("exected [lana], got %v", scs)
	}
}
