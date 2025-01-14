package language

// Code generated by typesetting-utils/generators/langsamples. DO NOT EDIT

type languageInfo struct {
	lang    Language
	scripts [3]Script // scripts used by the language
}

// languagesInfos defines the LangID
// and stores the scripts commonly used to write a language
var languagesInfos = [...]languageInfo{
	{},
	{ /** index: 1 */
		"aa",
		[3]Script{Latin},
	},
	{ /** index: 2 */
		"ab",
		[3]Script{Cyrillic},
	},
	{ /** index: 3 */
		"af",
		[3]Script{Latin},
	},
	{ /** index: 4 */
		"agr",
		[3]Script{Latin},
	},
	{ /** index: 5 */
		"ak",
		[3]Script{Latin},
	},
	{ /** index: 6 */
		"am",
		[3]Script{Ethiopic},
	},
	{ /** index: 7 */
		"an",
		[3]Script{Latin},
	},
	{ /** index: 8 */
		"anp",
		[3]Script{Devanagari},
	},
	{ /** index: 9 */
		"ar",
		[3]Script{Arabic},
	},
	{ /** index: 10 */
		"as",
		[3]Script{Bengali},
	},
	{ /** index: 11 */
		"ast",
		[3]Script{Latin},
	},
	{ /** index: 12 */
		"av",
		[3]Script{Cyrillic},
	},
	{ /** index: 13 */
		"ay",
		[3]Script{Latin},
	},
	{ /** index: 14 */
		"ayc",
		[3]Script{Latin},
	},
	{ /** index: 15 */
		"az-az",
		[3]Script{Latin},
	},
	{ /** index: 16 */
		"az-ir",
		[3]Script{Arabic},
	},
	{ /** index: 17 */
		"ba",
		[3]Script{Cyrillic},
	},
	{ /** index: 18 */
		"be",
		[3]Script{Cyrillic},
	},
	{ /** index: 19 */
		"bem",
		[3]Script{Latin},
	},
	{ /** index: 20 */
		"ber-dz",
		[3]Script{Latin},
	},
	{ /** index: 21 */
		"ber-ma",
		[3]Script{Tifinagh},
	},
	{ /** index: 22 */
		"bg",
		[3]Script{Cyrillic},
	},
	{ /** index: 23 */
		"bh",
		[3]Script{Devanagari},
	},
	{ /** index: 24 */
		"bhb",
		[3]Script{Devanagari},
	},
	{ /** index: 25 */
		"bho",
		[3]Script{Devanagari},
	},
	{ /** index: 26 */
		"bi",
		[3]Script{Latin},
	},
	{ /** index: 27 */
		"bin",
		[3]Script{Latin},
	},
	{ /** index: 28 */
		"bm",
		[3]Script{Latin},
	},
	{ /** index: 29 */
		"bn",
		[3]Script{Bengali},
	},
	{ /** index: 30 */
		"bo",
		[3]Script{Tibetan},
	},
	{ /** index: 31 */
		"br",
		[3]Script{Latin},
	},
	{ /** index: 32 */
		"brx",
		[3]Script{Devanagari},
	},
	{ /** index: 33 */
		"bs",
		[3]Script{Latin},
	},
	{ /** index: 34 */
		"bua",
		[3]Script{Cyrillic},
	},
	{ /** index: 35 */
		"byn",
		[3]Script{Ethiopic},
	},
	{ /** index: 36 */
		"ca",
		[3]Script{Latin},
	},
	{ /** index: 37 */
		"ce",
		[3]Script{Cyrillic},
	},
	{ /** index: 38 */
		"ch",
		[3]Script{Latin},
	},
	{ /** index: 39 */
		"chm",
		[3]Script{Cyrillic},
	},
	{ /** index: 40 */
		"chr",
		[3]Script{Cherokee},
	},
	{ /** index: 41 */
		"ckb",
		[3]Script{Arabic},
	},
	{ /** index: 42 */
		"cmn",
		[3]Script{Han},
	},
	{ /** index: 43 */
		"co",
		[3]Script{Latin},
	},
	{ /** index: 44 */
		"cop",
		[3]Script{Coptic},
	},
	{ /** index: 45 */
		"crh",
		[3]Script{Latin},
	},
	{ /** index: 46 */
		"cs",
		[3]Script{Latin},
	},
	{ /** index: 47 */
		"csb",
		[3]Script{Latin},
	},
	{ /** index: 48 */
		"cu",
		[3]Script{Cyrillic},
	},
	{ /** index: 49 */
		"cv",
		[3]Script{Cyrillic, Latin},
	},
	{ /** index: 50 */
		"cy",
		[3]Script{Latin},
	},
	{ /** index: 51 */
		"da",
		[3]Script{Latin},
	},
	{ /** index: 52 */
		"de",
		[3]Script{Latin},
	},
	{ /** index: 53 */
		"doi",
		[3]Script{Devanagari},
	},
	{ /** index: 54 */
		"dsb",
		[3]Script{Latin},
	},
	{ /** index: 55 */
		"dv",
		[3]Script{Thaana},
	},
	{ /** index: 56 */
		"dz",
		[3]Script{Tibetan},
	},
	{ /** index: 57 */
		"ee",
		[3]Script{Latin},
	},
	{ /** index: 58 */
		"el",
		[3]Script{Greek},
	},
	{ /** index: 59 */
		"en",
		[3]Script{Latin},
	},
	{ /** index: 60 */
		"eo",
		[3]Script{Latin},
	},
	{ /** index: 61 */
		"es",
		[3]Script{Latin},
	},
	{ /** index: 62 */
		"et",
		[3]Script{Latin},
	},
	{ /** index: 63 */
		"eu",
		[3]Script{Latin},
	},
	{ /** index: 64 */
		"fa",
		[3]Script{Arabic},
	},
	{ /** index: 65 */
		"fat",
		[3]Script{Latin},
	},
	{ /** index: 66 */
		"ff",
		[3]Script{Latin},
	},
	{ /** index: 67 */
		"fi",
		[3]Script{Latin},
	},
	{ /** index: 68 */
		"fil",
		[3]Script{Latin},
	},
	{ /** index: 69 */
		"fj",
		[3]Script{Latin},
	},
	{ /** index: 70 */
		"fo",
		[3]Script{Latin},
	},
	{ /** index: 71 */
		"fr",
		[3]Script{Latin},
	},
	{ /** index: 72 */
		"fur",
		[3]Script{Latin},
	},
	{ /** index: 73 */
		"fy",
		[3]Script{Latin},
	},
	{ /** index: 74 */
		"ga",
		[3]Script{Latin},
	},
	{ /** index: 75 */
		"gd",
		[3]Script{Latin},
	},
	{ /** index: 76 */
		"gez",
		[3]Script{Ethiopic},
	},
	{ /** index: 77 */
		"gl",
		[3]Script{Latin},
	},
	{ /** index: 78 */
		"gn",
		[3]Script{Latin},
	},
	{ /** index: 79 */
		"got",
		[3]Script{Gothic},
	},
	{ /** index: 80 */
		"gu",
		[3]Script{Gujarati},
	},
	{ /** index: 81 */
		"gv",
		[3]Script{Latin},
	},
	{ /** index: 82 */
		"ha",
		[3]Script{Latin},
	},
	{ /** index: 83 */
		"hak",
		[3]Script{Han},
	},
	{ /** index: 84 */
		"haw",
		[3]Script{Latin},
	},
	{ /** index: 85 */
		"he",
		[3]Script{Hebrew},
	},
	{ /** index: 86 */
		"hi",
		[3]Script{Devanagari},
	},
	{ /** index: 87 */
		"hif",
		[3]Script{Devanagari},
	},
	{ /** index: 88 */
		"hne",
		[3]Script{Devanagari},
	},
	{ /** index: 89 */
		"ho",
		[3]Script{Latin},
	},
	{ /** index: 90 */
		"hr",
		[3]Script{Latin},
	},
	{ /** index: 91 */
		"hsb",
		[3]Script{Latin},
	},
	{ /** index: 92 */
		"ht",
		[3]Script{Latin},
	},
	{ /** index: 93 */
		"hu",
		[3]Script{Latin},
	},
	{ /** index: 94 */
		"hy",
		[3]Script{Armenian},
	},
	{ /** index: 95 */
		"hz",
		[3]Script{Latin},
	},
	{ /** index: 96 */
		"ia",
		[3]Script{Latin},
	},
	{ /** index: 97 */
		"id",
		[3]Script{Latin},
	},
	{ /** index: 98 */
		"ie",
		[3]Script{Latin},
	},
	{ /** index: 99 */
		"ig",
		[3]Script{Latin},
	},
	{ /** index: 100 */
		"ii",
		[3]Script{Yi},
	},
	{ /** index: 101 */
		"ik",
		[3]Script{Cyrillic},
	},
	{ /** index: 102 */
		"io",
		[3]Script{Latin},
	},
	{ /** index: 103 */
		"is",
		[3]Script{Latin},
	},
	{ /** index: 104 */
		"it",
		[3]Script{Latin},
	},
	{ /** index: 105 */
		"iu",
		[3]Script{Canadian_Aboriginal},
	},
	{ /** index: 106 */
		"ja",
		[3]Script{Han, Hiragana, Katakana},
	},
	{ /** index: 107 */
		"jv",
		[3]Script{Latin},
	},
	{ /** index: 108 */
		"ka",
		[3]Script{Georgian},
	},
	{ /** index: 109 */
		"kaa",
		[3]Script{Cyrillic},
	},
	{ /** index: 110 */
		"kab",
		[3]Script{Latin},
	},
	{ /** index: 111 */
		"ki",
		[3]Script{Latin},
	},
	{ /** index: 112 */
		"kj",
		[3]Script{Latin},
	},
	{ /** index: 113 */
		"kk",
		[3]Script{Cyrillic},
	},
	{ /** index: 114 */
		"kl",
		[3]Script{Latin},
	},
	{ /** index: 115 */
		"km",
		[3]Script{Khmer},
	},
	{ /** index: 116 */
		"kn",
		[3]Script{Kannada},
	},
	{ /** index: 117 */
		"ko",
		[3]Script{Hangul},
	},
	{ /** index: 118 */
		"kok",
		[3]Script{Devanagari},
	},
	{ /** index: 119 */
		"kr",
		[3]Script{Latin},
	},
	{ /** index: 120 */
		"ks",
		[3]Script{Arabic},
	},
	{ /** index: 121 */
		"ku-am",
		[3]Script{Cyrillic},
	},
	{ /** index: 122 */
		"ku-iq",
		[3]Script{Arabic},
	},
	{ /** index: 123 */
		"ku-ir",
		[3]Script{Arabic},
	},
	{ /** index: 124 */
		"ku-tr",
		[3]Script{Latin},
	},
	{ /** index: 125 */
		"kum",
		[3]Script{Cyrillic},
	},
	{ /** index: 126 */
		"kv",
		[3]Script{Cyrillic},
	},
	{ /** index: 127 */
		"kw",
		[3]Script{Latin},
	},
	{ /** index: 128 */
		"kwm",
		[3]Script{Latin},
	},
	{ /** index: 129 */
		"ky",
		[3]Script{Cyrillic},
	},
	{ /** index: 130 */
		"la",
		[3]Script{Latin},
	},
	{ /** index: 131 */
		"lah",
		[3]Script{Arabic},
	},
	{ /** index: 132 */
		"lb",
		[3]Script{Latin},
	},
	{ /** index: 133 */
		"lez",
		[3]Script{Cyrillic},
	},
	{ /** index: 134 */
		"lg",
		[3]Script{Latin},
	},
	{ /** index: 135 */
		"li",
		[3]Script{Latin},
	},
	{ /** index: 136 */
		"lij",
		[3]Script{Latin},
	},
	{ /** index: 137 */
		"ln",
		[3]Script{Latin},
	},
	{ /** index: 138 */
		"lo",
		[3]Script{Lao},
	},
	{ /** index: 139 */
		"lt",
		[3]Script{Latin},
	},
	{ /** index: 140 */
		"lv",
		[3]Script{Latin},
	},
	{ /** index: 141 */
		"lzh",
		[3]Script{Han},
	},
	{ /** index: 142 */
		"mag",
		[3]Script{Devanagari},
	},
	{ /** index: 143 */
		"mai",
		[3]Script{Devanagari},
	},
	{ /** index: 144 */
		"mfe",
		[3]Script{Latin},
	},
	{ /** index: 145 */
		"mg",
		[3]Script{Latin},
	},
	{ /** index: 146 */
		"mh",
		[3]Script{Latin},
	},
	{ /** index: 147 */
		"mhr",
		[3]Script{Cyrillic},
	},
	{ /** index: 148 */
		"mi",
		[3]Script{Latin},
	},
	{ /** index: 149 */
		"miq",
		[3]Script{Latin},
	},
	{ /** index: 150 */
		"mjw",
		[3]Script{Latin},
	},
	{ /** index: 151 */
		"mk",
		[3]Script{Cyrillic},
	},
	{ /** index: 152 */
		"ml",
		[3]Script{Malayalam},
	},
	{ /** index: 153 */
		"mn-cn",
		[3]Script{Mongolian},
	},
	{ /** index: 154 */
		"mn-mn",
		[3]Script{Cyrillic},
	},
	{ /** index: 155 */
		"mni",
		[3]Script{Bengali},
	},
	{ /** index: 156 */
		"mnw",
		[3]Script{Myanmar},
	},
	{ /** index: 157 */
		"mo",
		[3]Script{Cyrillic, Latin},
	},
	{ /** index: 158 */
		"mr",
		[3]Script{Devanagari},
	},
	{ /** index: 159 */
		"ms",
		[3]Script{Latin},
	},
	{ /** index: 160 */
		"mt",
		[3]Script{Latin},
	},
	{ /** index: 161 */
		"my",
		[3]Script{Myanmar},
	},
	{ /** index: 162 */
		"na",
		[3]Script{Latin},
	},
	{ /** index: 163 */
		"nan",
		[3]Script{Han, Latin},
	},
	{ /** index: 164 */
		"nb",
		[3]Script{Latin},
	},
	{ /** index: 165 */
		"nds",
		[3]Script{Latin},
	},
	{ /** index: 166 */
		"ne",
		[3]Script{Devanagari},
	},
	{ /** index: 167 */
		"ng",
		[3]Script{Latin},
	},
	{ /** index: 168 */
		"nhn",
		[3]Script{Latin},
	},
	{ /** index: 169 */
		"niu",
		[3]Script{Latin},
	},
	{ /** index: 170 */
		"nl",
		[3]Script{Latin},
	},
	{ /** index: 171 */
		"nn",
		[3]Script{Latin},
	},
	{ /** index: 172 */
		"no",
		[3]Script{Latin},
	},
	{ /** index: 173 */
		"nqo",
		[3]Script{Nko},
	},
	{ /** index: 174 */
		"nr",
		[3]Script{Latin},
	},
	{ /** index: 175 */
		"nso",
		[3]Script{Latin},
	},
	{ /** index: 176 */
		"nv",
		[3]Script{Latin},
	},
	{ /** index: 177 */
		"ny",
		[3]Script{Latin},
	},
	{ /** index: 178 */
		"oc",
		[3]Script{Latin},
	},
	{ /** index: 179 */
		"om",
		[3]Script{Latin},
	},
	{ /** index: 180 */
		"or",
		[3]Script{Oriya},
	},
	{ /** index: 181 */
		"os",
		[3]Script{Cyrillic},
	},
	{ /** index: 182 */
		"ota",
		[3]Script{Arabic},
	},
	{ /** index: 183 */
		"pa",
		[3]Script{Gurmukhi},
	},
	{ /** index: 184 */
		"pa-pk",
		[3]Script{Arabic},
	},
	{ /** index: 185 */
		"pap-an",
		[3]Script{Latin},
	},
	{ /** index: 186 */
		"pap-aw",
		[3]Script{Latin},
	},
	{ /** index: 187 */
		"pes",
		[3]Script{Arabic},
	},
	{ /** index: 188 */
		"pl",
		[3]Script{Latin},
	},
	{ /** index: 189 */
		"prs",
		[3]Script{Arabic},
	},
	{ /** index: 190 */
		"ps-af",
		[3]Script{Arabic},
	},
	{ /** index: 191 */
		"ps-pk",
		[3]Script{Arabic},
	},
	{ /** index: 192 */
		"pt",
		[3]Script{Latin},
	},
	{ /** index: 193 */
		"qu",
		[3]Script{Latin},
	},
	{ /** index: 194 */
		"quz",
		[3]Script{Latin},
	},
	{ /** index: 195 */
		"raj",
		[3]Script{Devanagari},
	},
	{ /** index: 196 */
		"rif",
		[3]Script{Latin},
	},
	{ /** index: 197 */
		"rm",
		[3]Script{Latin},
	},
	{ /** index: 198 */
		"rn",
		[3]Script{Latin},
	},
	{ /** index: 199 */
		"ro",
		[3]Script{Latin},
	},
	{ /** index: 200 */
		"ru",
		[3]Script{Cyrillic},
	},
	{ /** index: 201 */
		"rw",
		[3]Script{Latin},
	},
	{ /** index: 202 */
		"sa",
		[3]Script{Devanagari},
	},
	{ /** index: 203 */
		"sah",
		[3]Script{Cyrillic},
	},
	{ /** index: 204 */
		"sat",
		[3]Script{Devanagari},
	},
	{ /** index: 205 */
		"sc",
		[3]Script{Latin},
	},
	{ /** index: 206 */
		"sco",
		[3]Script{Latin},
	},
	{ /** index: 207 */
		"sd",
		[3]Script{Arabic},
	},
	{ /** index: 208 */
		"se",
		[3]Script{Latin},
	},
	{ /** index: 209 */
		"sel",
		[3]Script{Cyrillic},
	},
	{ /** index: 210 */
		"sg",
		[3]Script{Latin},
	},
	{ /** index: 211 */
		"sgs",
		[3]Script{Latin},
	},
	{ /** index: 212 */
		"sh",
		[3]Script{Cyrillic, Latin},
	},
	{ /** index: 213 */
		"shn",
		[3]Script{Myanmar},
	},
	{ /** index: 214 */
		"shs",
		[3]Script{Latin},
	},
	{ /** index: 215 */
		"si",
		[3]Script{Sinhala},
	},
	{ /** index: 216 */
		"sid",
		[3]Script{Ethiopic},
	},
	{ /** index: 217 */
		"sk",
		[3]Script{Latin},
	},
	{ /** index: 218 */
		"sl",
		[3]Script{Latin},
	},
	{ /** index: 219 */
		"sm",
		[3]Script{Latin},
	},
	{ /** index: 220 */
		"sma",
		[3]Script{Latin},
	},
	{ /** index: 221 */
		"smj",
		[3]Script{Latin},
	},
	{ /** index: 222 */
		"smn",
		[3]Script{Latin},
	},
	{ /** index: 223 */
		"sms",
		[3]Script{Latin},
	},
	{ /** index: 224 */
		"sn",
		[3]Script{Latin},
	},
	{ /** index: 225 */
		"so",
		[3]Script{Latin},
	},
	{ /** index: 226 */
		"sq",
		[3]Script{Latin},
	},
	{ /** index: 227 */
		"sr",
		[3]Script{Cyrillic},
	},
	{ /** index: 228 */
		"ss",
		[3]Script{Latin},
	},
	{ /** index: 229 */
		"st",
		[3]Script{Latin},
	},
	{ /** index: 230 */
		"su",
		[3]Script{Latin},
	},
	{ /** index: 231 */
		"sv",
		[3]Script{Latin},
	},
	{ /** index: 232 */
		"sw",
		[3]Script{Latin},
	},
	{ /** index: 233 */
		"syr",
		[3]Script{Syriac},
	},
	{ /** index: 234 */
		"szl",
		[3]Script{Latin},
	},
	{ /** index: 235 */
		"ta",
		[3]Script{Tamil},
	},
	{ /** index: 236 */
		"tcy",
		[3]Script{Kannada},
	},
	{ /** index: 237 */
		"te",
		[3]Script{Telugu},
	},
	{ /** index: 238 */
		"tg",
		[3]Script{Cyrillic},
	},
	{ /** index: 239 */
		"th",
		[3]Script{Thai},
	},
	{ /** index: 240 */
		"the",
		[3]Script{Devanagari},
	},
	{ /** index: 241 */
		"ti-er",
		[3]Script{Ethiopic},
	},
	{ /** index: 242 */
		"ti-et",
		[3]Script{Ethiopic},
	},
	{ /** index: 243 */
		"tig",
		[3]Script{Ethiopic},
	},
	{ /** index: 244 */
		"tk",
		[3]Script{Latin},
	},
	{ /** index: 245 */
		"tl",
		[3]Script{Latin},
	},
	{ /** index: 246 */
		"tn",
		[3]Script{Latin},
	},
	{ /** index: 247 */
		"to",
		[3]Script{Latin},
	},
	{ /** index: 248 */
		"tpi",
		[3]Script{Latin},
	},
	{ /** index: 249 */
		"tr",
		[3]Script{Latin},
	},
	{ /** index: 250 */
		"ts",
		[3]Script{Latin},
	},
	{ /** index: 251 */
		"tt",
		[3]Script{Cyrillic},
	},
	{ /** index: 252 */
		"tw",
		[3]Script{Latin},
	},
	{ /** index: 253 */
		"ty",
		[3]Script{Latin},
	},
	{ /** index: 254 */
		"tyv",
		[3]Script{Cyrillic},
	},
	{ /** index: 255 */
		"ug",
		[3]Script{Arabic},
	},
	{ /** index: 256 */
		"uk",
		[3]Script{Cyrillic},
	},
	{ /** index: 257 */
		"und-zmth",
		[3]Script{Greek, Latin},
	},
	{ /** index: 258 */
		"und-zsye",
		[3]Script{},
	},
	{ /** index: 259 */
		"unm",
		[3]Script{Latin},
	},
	{ /** index: 260 */
		"ur",
		[3]Script{Arabic},
	},
	{ /** index: 261 */
		"uz",
		[3]Script{Latin},
	},
	{ /** index: 262 */
		"ve",
		[3]Script{Latin},
	},
	{ /** index: 263 */
		"vi",
		[3]Script{Latin},
	},
	{ /** index: 264 */
		"vo",
		[3]Script{Latin},
	},
	{ /** index: 265 */
		"vot",
		[3]Script{Latin},
	},
	{ /** index: 266 */
		"wa",
		[3]Script{Latin},
	},
	{ /** index: 267 */
		"wae",
		[3]Script{Latin},
	},
	{ /** index: 268 */
		"wal",
		[3]Script{Ethiopic},
	},
	{ /** index: 269 */
		"wen",
		[3]Script{Latin},
	},
	{ /** index: 270 */
		"wo",
		[3]Script{Latin},
	},
	{ /** index: 271 */
		"xh",
		[3]Script{Latin},
	},
	{ /** index: 272 */
		"yap",
		[3]Script{Latin},
	},
	{ /** index: 273 */
		"yi",
		[3]Script{Hebrew},
	},
	{ /** index: 274 */
		"yo",
		[3]Script{Latin},
	},
	{ /** index: 275 */
		"yue",
		[3]Script{Han},
	},
	{ /** index: 276 */
		"yuw",
		[3]Script{Latin},
	},
	{ /** index: 277 */
		"za",
		[3]Script{Latin},
	},
	{ /** index: 278 */
		"zh-cn",
		[3]Script{Han},
	},
	{ /** index: 279 */
		"zh-hk",
		[3]Script{Han},
	},
	{ /** index: 280 */
		"zh-mo",
		[3]Script{Han},
	},
	{ /** index: 281 */
		"zh-sg",
		[3]Script{Han},
	},
	{ /** index: 282 */
		"zh-tw",
		[3]Script{Han},
	},
	{ /** index: 283 */
		"zu",
		[3]Script{Latin},
	},
	{ /** index: 284 */
		"bku",
		[3]Script{},
	},
	{ /** index: 285 */
		"bug",
		[3]Script{},
	},
	{ /** index: 286 */
		"hnn",
		[3]Script{},
	},
	{ /** index: 287 */
		"ks-devanagari",
		[3]Script{},
	},
	{ /** index: 288 */
		"ml-in",
		[3]Script{},
	},
	{ /** index: 289 */
		"mn",
		[3]Script{},
	},
	{ /** index: 290 */
		"peo",
		[3]Script{},
	},
	{ /** index: 291 */
		"sd-devanagari",
		[3]Script{},
	},
	{ /** index: 292 */
		"syl",
		[3]Script{},
	},
	{ /** index: 293 */
		"tbw",
		[3]Script{},
	},
	{ /** index: 294 */
		"uga",
		[3]Script{},
	},
}

const (
	LangAa            LangID = 1
	LangAb            LangID = 2
	LangAf            LangID = 3
	LangAgr           LangID = 4
	LangAk            LangID = 5
	LangAm            LangID = 6
	LangAn            LangID = 7
	LangAnp           LangID = 8
	LangAr            LangID = 9
	LangAs            LangID = 10
	LangAst           LangID = 11
	LangAv            LangID = 12
	LangAy            LangID = 13
	LangAyc           LangID = 14
	LangAz_Az         LangID = 15
	LangAz_Ir         LangID = 16
	LangBa            LangID = 17
	LangBe            LangID = 18
	LangBem           LangID = 19
	LangBer_Dz        LangID = 20
	LangBer_Ma        LangID = 21
	LangBg            LangID = 22
	LangBh            LangID = 23
	LangBhb           LangID = 24
	LangBho           LangID = 25
	LangBi            LangID = 26
	LangBin           LangID = 27
	LangBm            LangID = 28
	LangBn            LangID = 29
	LangBo            LangID = 30
	LangBr            LangID = 31
	LangBrx           LangID = 32
	LangBs            LangID = 33
	LangBua           LangID = 34
	LangByn           LangID = 35
	LangCa            LangID = 36
	LangCe            LangID = 37
	LangCh            LangID = 38
	LangChm           LangID = 39
	LangChr           LangID = 40
	LangCkb           LangID = 41
	LangCmn           LangID = 42
	LangCo            LangID = 43
	LangCop           LangID = 44
	LangCrh           LangID = 45
	LangCs            LangID = 46
	LangCsb           LangID = 47
	LangCu            LangID = 48
	LangCv            LangID = 49
	LangCy            LangID = 50
	LangDa            LangID = 51
	LangDe            LangID = 52
	LangDoi           LangID = 53
	LangDsb           LangID = 54
	LangDv            LangID = 55
	LangDz            LangID = 56
	LangEe            LangID = 57
	LangEl            LangID = 58
	LangEn            LangID = 59
	LangEo            LangID = 60
	LangEs            LangID = 61
	LangEt            LangID = 62
	LangEu            LangID = 63
	LangFa            LangID = 64
	LangFat           LangID = 65
	LangFf            LangID = 66
	LangFi            LangID = 67
	LangFil           LangID = 68
	LangFj            LangID = 69
	LangFo            LangID = 70
	LangFr            LangID = 71
	LangFur           LangID = 72
	LangFy            LangID = 73
	LangGa            LangID = 74
	LangGd            LangID = 75
	LangGez           LangID = 76
	LangGl            LangID = 77
	LangGn            LangID = 78
	LangGot           LangID = 79
	LangGu            LangID = 80
	LangGv            LangID = 81
	LangHa            LangID = 82
	LangHak           LangID = 83
	LangHaw           LangID = 84
	LangHe            LangID = 85
	LangHi            LangID = 86
	LangHif           LangID = 87
	LangHne           LangID = 88
	LangHo            LangID = 89
	LangHr            LangID = 90
	LangHsb           LangID = 91
	LangHt            LangID = 92
	LangHu            LangID = 93
	LangHy            LangID = 94
	LangHz            LangID = 95
	LangIa            LangID = 96
	LangId            LangID = 97
	LangIe            LangID = 98
	LangIg            LangID = 99
	LangIi            LangID = 100
	LangIk            LangID = 101
	LangIo            LangID = 102
	LangIs            LangID = 103
	LangIt            LangID = 104
	LangIu            LangID = 105
	LangJa            LangID = 106
	LangJv            LangID = 107
	LangKa            LangID = 108
	LangKaa           LangID = 109
	LangKab           LangID = 110
	LangKi            LangID = 111
	LangKj            LangID = 112
	LangKk            LangID = 113
	LangKl            LangID = 114
	LangKm            LangID = 115
	LangKn            LangID = 116
	LangKo            LangID = 117
	LangKok           LangID = 118
	LangKr            LangID = 119
	LangKs            LangID = 120
	LangKu_Am         LangID = 121
	LangKu_Iq         LangID = 122
	LangKu_Ir         LangID = 123
	LangKu_Tr         LangID = 124
	LangKum           LangID = 125
	LangKv            LangID = 126
	LangKw            LangID = 127
	LangKwm           LangID = 128
	LangKy            LangID = 129
	LangLa            LangID = 130
	LangLah           LangID = 131
	LangLb            LangID = 132
	LangLez           LangID = 133
	LangLg            LangID = 134
	LangLi            LangID = 135
	LangLij           LangID = 136
	LangLn            LangID = 137
	LangLo            LangID = 138
	LangLt            LangID = 139
	LangLv            LangID = 140
	LangLzh           LangID = 141
	LangMag           LangID = 142
	LangMai           LangID = 143
	LangMfe           LangID = 144
	LangMg            LangID = 145
	LangMh            LangID = 146
	LangMhr           LangID = 147
	LangMi            LangID = 148
	LangMiq           LangID = 149
	LangMjw           LangID = 150
	LangMk            LangID = 151
	LangMl            LangID = 152
	LangMn_Cn         LangID = 153
	LangMn_Mn         LangID = 154
	LangMni           LangID = 155
	LangMnw           LangID = 156
	LangMo            LangID = 157
	LangMr            LangID = 158
	LangMs            LangID = 159
	LangMt            LangID = 160
	LangMy            LangID = 161
	LangNa            LangID = 162
	LangNan           LangID = 163
	LangNb            LangID = 164
	LangNds           LangID = 165
	LangNe            LangID = 166
	LangNg            LangID = 167
	LangNhn           LangID = 168
	LangNiu           LangID = 169
	LangNl            LangID = 170
	LangNn            LangID = 171
	LangNo            LangID = 172
	LangNqo           LangID = 173
	LangNr            LangID = 174
	LangNso           LangID = 175
	LangNv            LangID = 176
	LangNy            LangID = 177
	LangOc            LangID = 178
	LangOm            LangID = 179
	LangOr            LangID = 180
	LangOs            LangID = 181
	LangOta           LangID = 182
	LangPa            LangID = 183
	LangPa_Pk         LangID = 184
	LangPap_An        LangID = 185
	LangPap_Aw        LangID = 186
	LangPes           LangID = 187
	LangPl            LangID = 188
	LangPrs           LangID = 189
	LangPs_Af         LangID = 190
	LangPs_Pk         LangID = 191
	LangPt            LangID = 192
	LangQu            LangID = 193
	LangQuz           LangID = 194
	LangRaj           LangID = 195
	LangRif           LangID = 196
	LangRm            LangID = 197
	LangRn            LangID = 198
	LangRo            LangID = 199
	LangRu            LangID = 200
	LangRw            LangID = 201
	LangSa            LangID = 202
	LangSah           LangID = 203
	LangSat           LangID = 204
	LangSc            LangID = 205
	LangSco           LangID = 206
	LangSd            LangID = 207
	LangSe            LangID = 208
	LangSel           LangID = 209
	LangSg            LangID = 210
	LangSgs           LangID = 211
	LangSh            LangID = 212
	LangShn           LangID = 213
	LangShs           LangID = 214
	LangSi            LangID = 215
	LangSid           LangID = 216
	LangSk            LangID = 217
	LangSl            LangID = 218
	LangSm            LangID = 219
	LangSma           LangID = 220
	LangSmj           LangID = 221
	LangSmn           LangID = 222
	LangSms           LangID = 223
	LangSn            LangID = 224
	LangSo            LangID = 225
	LangSq            LangID = 226
	LangSr            LangID = 227
	LangSs            LangID = 228
	LangSt            LangID = 229
	LangSu            LangID = 230
	LangSv            LangID = 231
	LangSw            LangID = 232
	LangSyr           LangID = 233
	LangSzl           LangID = 234
	LangTa            LangID = 235
	LangTcy           LangID = 236
	LangTe            LangID = 237
	LangTg            LangID = 238
	LangTh            LangID = 239
	LangThe           LangID = 240
	LangTi_Er         LangID = 241
	LangTi_Et         LangID = 242
	LangTig           LangID = 243
	LangTk            LangID = 244
	LangTl            LangID = 245
	LangTn            LangID = 246
	LangTo            LangID = 247
	LangTpi           LangID = 248
	LangTr            LangID = 249
	LangTs            LangID = 250
	LangTt            LangID = 251
	LangTw            LangID = 252
	LangTy            LangID = 253
	LangTyv           LangID = 254
	LangUg            LangID = 255
	LangUk            LangID = 256
	LangUnd_Zmth      LangID = 257
	LangUnd_Zsye      LangID = 258
	LangUnm           LangID = 259
	LangUr            LangID = 260
	LangUz            LangID = 261
	LangVe            LangID = 262
	LangVi            LangID = 263
	LangVo            LangID = 264
	LangVot           LangID = 265
	LangWa            LangID = 266
	LangWae           LangID = 267
	LangWal           LangID = 268
	LangWen           LangID = 269
	LangWo            LangID = 270
	LangXh            LangID = 271
	LangYap           LangID = 272
	LangYi            LangID = 273
	LangYo            LangID = 274
	LangYue           LangID = 275
	LangYuw           LangID = 276
	LangZa            LangID = 277
	LangZh_Cn         LangID = 278
	LangZh_Hk         LangID = 279
	LangZh_Mo         LangID = 280
	LangZh_Sg         LangID = 281
	LangZh_Tw         LangID = 282
	LangZu            LangID = 283
	LangBku           LangID = 284
	LangBug           LangID = 285
	LangHnn           LangID = 286
	LangKs_Devanagari LangID = 287
	LangMl_In         LangID = 288
	LangMn            LangID = 289
	LangPeo           LangID = 290
	LangSd_Devanagari LangID = 291
	LangSyl           LangID = 292
	LangTbw           LangID = 293
	LangUga           LangID = 294

	// languages in languagesInfos[knownLangsCount:] have no orthographic samples in fontconfig source,
	// but are used in substitutions
	knownLangsCount LangID = 284
)
