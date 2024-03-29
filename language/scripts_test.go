package language

import (
	"testing"
	"unicode"
)

func TestParseScript(t *testing.T) {
	tests := []struct {
		args    string
		want    Script
		wantErr bool
	}{
		{"xxx", 0, true},
		{"bamu", Bamum, false},
		{"bamu_to_long", Bamum, false},
		{"cyrl", Cyrillic, false},
		{"samr", Samaritan, false},
		{"ARAB", Arabic, false},
		{"arab", Arabic, false},
		{"Arab", Arabic, false},
		{"Samr", Samaritan, false},
	}
	for _, tt := range tests {
		got, err := ParseScript(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseScript() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if got != tt.want {
			t.Errorf("ParseScript() = %v, want %v", got, tt.want)
		}
	}
}

func TestScript_String(t *testing.T) {
	if Bamum.String() != "Bamu" {
		t.Fatal()
	}
}

// used as benchmark reference
func lookupScriptNaive(r rune) Script {
	for name, table := range unicode.Scripts {
		if unicode.Is(table, r) {
			return scriptToTag[name]
		}
	}
	return Unknown
}

func TestFastLookup(t *testing.T) {
	for _, r := range scriptsSample {
		g1, g2 := lookupScriptNaive(r), LookupScript(r)
		if g1 != g2 {
			t.Fatalf("for rune 0x%x, expected %s, got %s", r, g1, g2)
		}
	}
}

func BenchmarkLookupScript(b *testing.B) {
	b.Run("naive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range scriptsSample {
				_ = lookupScriptNaive(r)
			}
		}
	})
	b.Run("optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range scriptsSample {
				_ = LookupScript(r)
			}
		}
	})
}

//lint:ignore ST1018 for simplicity
const scriptsSample = `
	Ek kan glas eet, maar dit doen my nie skade nie. 
	نص حكيم له سر قاطع وذو شأن عظيم مكتوب على ثوب أخضر ومغلف بجلد أزرق. 
	Gvxam mincetu apocikvyeh: ñizol ce mamvj ka raq kuse bafkeh mew. 
	I koh Glos esa, und es duard ma ned wei. 
	Под южно дърво, цъфтящо в синьо, бягаше малко пухкаво зайче. 
	Mi save kakae glas, hemi no save katem mi. 
	আমি কাঁচ খেতে পারি, তাতে আমার কোনো ক্ষতি হয় না। 
	ཤེལ་སྒོ་ཟ་ནས་ང་ན་གི་མ་རེད། 
	Fin džip, gluh jež i čvrst konjić dođoše bez moljca. 
	Jove xef, porti whisky amb quinze glaçons d'hidrogen, coi! 
	Siña yo' chumocho krestat, ti ha na'lalamen yo'. 
	Příliš žluťoučký kůň úpěl ďábelské ódy. 
	Dw i'n gallu bwyta gwydr, 'dyw e ddim yn gwneud dolur i mi. 
	Quizdeltagerne spiste jordbær med fløde, mens cirkusklovnen Walther spillede på xylofon. 
	Zwölf Boxkämpfer jagen Viktor quer über den großen Sylter Deich. 
	މާއްދާ 1 – ހުރިހާ އިންސާނުން ވެސް އުފަންވަނީ، ދަރަޖަ އާއި ޙައްޤު ތަކުގައި މިނިވަންކަމާއި ހަމަހަމަކަން ލިބިގެންވާ ބައެއްގެ ގޮތުގައެވެ. 
	Θέλει αρετή και τόλμη η ελευθερία. (Ανδρέας Κάλβος) 
	The quick brown fox jumps over the lazy dog.
	Ich canne glas eten and hit hirtiþ me nouȝt. 
	Eĥoŝanĝo ĉiuĵaŭde. 
	Jovencillo emponzoñado de whisky: ¡qué figurota exhibe! 
	See väike mölder jõuab rongile hüpata. 
	Kristala jan dezaket, ez dit minik ematen. 
	«الا یا اَیُّها السّاقی! اَدِرْ کَأساً وَ ناوِلْها!» که عشق آسان نمود اوّل، ولی افتاد مشکل‌ها!
	Viekas kettu punaturkki laiskan koiran takaa kurkki. 
	Voix ambiguë d'un cœur qui, au zéphyr, préfère les jattes de kiwis. 
	Je puis mangier del voirre. Ne me nuit. 
	Chuaigh bé mhórshách le dlúthspád fíorfhinn trí hata mo dhea-phorcáin bhig. 
	S urrainn dhomh gloinne ithe; cha ghoirtich i mi. 
	Eu podo xantar cristais e non cortarme. 
	𐌼𐌰𐌲 𐌲𐌻𐌴𐍃 𐌹̈𐍄𐌰𐌽, 𐌽𐌹 𐌼𐌹𐍃 𐍅𐌿 𐌽𐌳𐌰𐌽 𐌱𐍂𐌹𐌲𐌲𐌹𐌸. 
	હું કાચ ખાઇ શકુ છુ અને તેનાથી મને દર્દ નથી થતુ. 
	Foddym gee glonney agh cha jean eh gortaghey mee. 
	Hiki iaʻu ke ʻai i ke aniani; ʻaʻole nō lā au e ʻeha. 
	דג סקרן שט לו בים זך אך לפתע פגש חבורה נחמדה שצצה כך. 
	नहीं नजर किसी की बुरी नहीं किसी का मुँह काला जो करे सो उपर वाला 
	Deblji krojač: zgužvah smeđ filc u tanjušni džepić. 
	Egy hűtlen vejét fülöncsípő, dühös mexikói úr Wesselényinél mázol Quitóban. 
	Կրնամ ապակի ուտել և ինծի անհանգիստ չըներ։ 
	Kæmi ný öxi hér ykist þjófum nú bæði víl og ádrepa 
	Ma la volpe, col suo balzo, ha raggiunto il quieto Fido. 
	いろはにほへと ちりぬるを 色は匂へど 散りぬるを
	Chruu, a kwik di kwik brong fox a jomp huova di liezi daag de, yu no siit?
	.o'i mu xagji sofybakni cu zvati le purdi 
	Aku isa mangan beling tanpa lara. 
	მინას ვჭამ და არა მტკივა. 
	ខ្ញុំអាចញុំកញ្ចក់បាន ដោយគ្មានបញ្ហារ 

	ನಾನು ಗಾಜನ್ನು ತಿನ್ನಬಲ್ಲೆ ಮತ್ತು ಅದರಿಂದ ನನಗೆ ನೋವಾಗುವುದಿಲ್ಲ. 
	다람쥐 헌 쳇바퀴에 타고파 
	Mý a yl dybry gwéder hag éf ny wra ow ankenya. 
	Sic surgens, dux, zelotypos quam karus haberis
	ຂອ້ຍກິນແກ້ວໄດ້ໂດຍທີ່ມັນບໍ່ໄດ້ເຮັດໃຫ້ຂອ້ຍເຈັບ 

	Įlinkdama fechtuotojo špaga sublykčiojusi pragręžė apvalų arbūzą. 
	Sarkanās jūrascūciņas peld pa jūru. 
	E koʻana e kai i te karahi, mea ʻā, ʻaʻe hauhau. 
	Можам да јадам стакло, а не ме штета. 
	വേദനയില്ലാതെ കുപ്പിചില്ലു് എനിയ്ക്കു് കഴിയ്ക്കാം. 
	ᠪᠢ ᠰᠢᠯᠢ ᠢᠳᠡᠶᠦ ᠴᠢᠳᠠᠨᠠ ᠂ ᠨᠠᠳᠤᠷ ᠬᠣᠤᠷᠠᠳᠠᠢ ᠪᠢᠰᠢ 
	मी काच खाऊ शकतो, मला ते दुखत नाही. 
	Saya boleh makan kaca dan ia tidak mencederakan saya. 
	ဘာသာပြန်နှင့် စာပေပြုစုရေး ကော်မရှင် 
	M' pozz magna' o'vetr, e nun m' fa mal. 
	Vår sære Zulu fra badeøya spilte jo whist og quickstep i min taxi.
	Pa's wijze lynx bezag vroom het fikse aquaduct. 
	Eg kan eta glas utan å skada meg. 
	Vår sære Zulu fra badeøya spilte jo whist og quickstep i min taxi.
	Tsésǫʼ yishą́ągo bííníshghah dóó doo shił neezgai da. 
	Pòdi manjar de veire, me nafrariá pas. 
	ମୁଁ କାଚ ଖାଇପାରେ ଏବଂ ତାହା ମୋର କ୍ଷତି କରିନଥାଏ।. 
	ਮੈਂ ਗਲਾਸ ਖਾ ਸਕਦਾ ਹਾਂ ਅਤੇ ਇਸ ਨਾਲ ਮੈਨੂੰ ਕੋਈ ਤਕਲੀਫ ਨਹੀਂ. 
	Ch'peux mingi du verre, cha m'foé mie n'ma. 
	Pchnąć w tę łódź jeża lub ośm skrzyń fig. 
	Vejam a bruxa da raposa Salta-Pocinhas e o cão feliz que dorme regalado. 
	À noite, vovô Kowalsky vê o ímã cair no pé do pingüim queixoso e vovó põe açúcar no chá de tâmaras do jabuti feliz. 
	Fumegând hipnotic sașiul azvârle mreje în bălți. 
	В чащах юга жил бы цитрус? Да, но фальшивый экземпляр! 
	काचं शक्नोम्यत्तुम् । नोपहिनस्ति माम् ॥ 
	Puotsu mangiari u vitru, nun mi fa mali. 
	මනොපුබ්‌බඞ්‌ගමා ධම්‌මා, මනොසෙට්‌ඨා මනොමයා; මනසා චෙ පදුට්‌ඨෙන, භාසති වා කරොති වා; තතො නං දුක්‌ඛමන්‌වෙති, චක්‌කංව වහතො පදං.
	Starý kôň na hŕbe kníh žuje tíško povädnuté ruže, na stĺpe sa ďateľ učí kvákať novú ódu o živote.
	Šerif bo za vajo spet kuhal domače žgance. 
	Unë mund të ha qelq dhe nuk më gjen gjë. 
	Чешће цeђење мрeжастим џаком побољшава фертилизацију генских хибрида. 
	Flygande bäckasiner söka strax hwila på mjuka tuvor. 
	I kå Glas frässa, ond des macht mr nix! 
	நான் கண்ணாடி சாப்பிடுவேன், அதனால் எனக்கு ஒரு கேடும் வராது. 
	నేను గాజు తినగలను అయినా నాకు యేమీ కాదు. 
	เป็นมนุษย์สุดประเสริฐเลิศคุณค่า - กว่าบรรดาฝูงสัตว์เดรัจฉาน - จงฝ่าฟันพัฒนาวิชาการ อย่าล้างผลาญฤๅเข่นฆ่าบีฑาใคร - ไม่ถือโทษโกรธแช่งซัดฮึดฮัดด่า - หัดอภัยเหมือนกีฬาอัชฌาสัย - ปฏิบัติประพฤติกฎกำหนดใจ - พูดจาให้จ๊ะ ๆ จ๋า ๆ น่าฟังเอยฯ 
	Kaya kong kumain nang bubog at hindi ako masaktan. 
	Pijamalı hasta yağız şoföre çabucak güvendi. 
	Metumi awe tumpan, ɜnyɜ me hwee. 
	Чуєш їх, доцю, га? Кумедна ж ти, прощайся без ґольфів! 
	میں کانچ کھا سکتا ہوں اور مجھے تکلیف نہیں ہوتی ۔ 
	Mi posso magnare el vetro, no'l me fa mae. 
	Con sói nâu nhảy qua con chó lười.
	Dji pou magnî do vêre, çoula m' freut nén må. 
	איך קען עסן גלאָז און עס טוט מיר נישט װײ. 
	Mo lè je̩ dígí, kò ní pa mí lára. 
	我能吞下玻璃而不伤身体。 
	我能吞下玻璃而不傷身體。 
	我能吞下玻璃而不伤身体。 
	我能吞下玻璃而不傷身體。 
	Saya boleh makan kaca dan ia tidak mencederakan saya. 
`
