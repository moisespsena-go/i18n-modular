package i18nmod


// Backend defined methods that needs for translation backend
type Backend interface {
	ListGroups() []string
	ListLanguages() []string
	LoadTranslations(lang string, group string) (*Tree, error)
	SaveTranslation(*Translation) error
	DeleteTranslation(*Translation) error
}
