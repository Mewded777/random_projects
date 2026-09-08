package domain

type Book struct {
	ID        string
	Title     string
	Author    string
	Copies    int
	Available int
	Language  string
	Genre     string
}

// Catalog languages. The librarian picks one of these first when adding a book, which in turn determines which genres are offered.
const (
	LanguageEnglish = "English"
	LanguageAmharic = "Amharic"
)

func Languages() []string {
	return []string{LanguageEnglish, LanguageAmharic}
}

// GenresByLanguage holds the set of genres available for each supported language.
// Each language keeps its own list rather than sharing one

var GenresByLanguage = map[string][]string{
	LanguageEnglish: {
		"Comedy",
		"Tragedy",
		"Romance",
		"Drama",
		"Fantasy",
		"Mystery",
		"Science Fiction",
		"Horror",
		"Poetry",
		"Biography",
	},
	LanguageAmharic: {
		"አስቂኝ (Comedy)",
		"አሳዛኝ (Tragedy)",
		"የፍቅር (Romance)",
		"ድራማ (Drama)",
		"ልቦለድ (Fiction)",
		"ግጥም (Poetry)",
		"ታሪክ (History)",
		"የሕይወት ታሪክ (Biography)",
		"ሃይማኖታዊ (Religious)",
		"ተረት (Folktale)",
	},
}

// GenresFor returns the genre list for a given language, or nil if the language isn't recognized.
func GenresFor(language string) []string {
	return GenresByLanguage[language]
}

// IsValidLanguage reports whether language is one of the supported catalog languages.
func IsValidLanguage(language string) bool {
	_, ok := GenresByLanguage[language]
	return ok
}

// IsValidGenre reports whether genre is one of the genres defined for language.
func IsValidGenre(language, genre string) bool {
	for _, g := range GenresByLanguage[language] {
		if g == genre {
			return true
		}
	}
	return false
}
