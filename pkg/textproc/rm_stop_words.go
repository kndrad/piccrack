package textproc

import (
	"github.com/bbalet/stopwords"
	"github.com/pemistahl/lingua-go"
)

func defaultLanguages() []lingua.Language {
	return []lingua.Language{
		lingua.English,
		lingua.Polish,
	}
}

func RmStopWords(text string, langs ...lingua.Language) string {
	if text == "" {
		return ""
	}
	if len(langs) == 0 {
		langs = defaultLanguages()
	}
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(langs...).
		Build()

	lang, exists := detector.DetectLanguageOf(text)
	if !exists {
		return text
	}

	return stopwords.CleanString(text, lang.IsoCode639_1().String(), false)
}
