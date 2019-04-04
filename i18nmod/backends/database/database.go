package database

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/moisespsena-go/i18n-modular/i18nmod"
	"errors"
)

// Translation is a struct used to save translations into databae
type Translation struct {
	Locale string `sql:"size:12;"`
	Group    string `sql:"size:255;"`
	Key    string `sql:"size:4294967295;"`
	Value  string `sql:"size:4294967295"`
}

// New new DB backend for I18n
func New(db *gorm.DB) *Backend {
	db.AutoMigrate(&Translation{})
	model := db.Model(&Translation{})
	if err := model.AddUniqueIndex("idx_translations_key_with_locale", "locale", "key").Error; err != nil {
		fmt.Printf("Failed to create unique index for translations key & locale, got: %v\n", err.Error())
	}
	if err := model.AddIndex("idx_translations_group", "group").Error; err != nil {
		fmt.Printf("Failed to create index for translations group, got: %v\n", err.Error())
	}
	return &Backend{DB: db}
}

// Backend DB backend
type Backend struct {
	DB *gorm.DB
}

// LoadTranslations load translations from DB backend
func (backend *Backend) LoadTranslations() (translations []*Translation) {
	backend.DB.Find(&translations)
	return translations
}

func (backend *Backend) LoadTranslationsInGroup(group string) (translations []*i18nmod.Translation) {
	backend.DB.Where(Translation{Group:group}).Find(&translations)
	return
}

func (backend *Backend) LoadContent(content []byte) (translations []*i18nmod.Translation, err error)  {
	return translations, errors.New("not implemented")
}

// SaveTranslation save translation into DB backend
func (backend *Backend) SaveTranslation(t *i18nmod.Translation) error {
	return backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).
		Assign(Translation{Group: t.Group, Value: t.Value}).
		FirstOrCreate(&Translation{}).Error
}

// DeleteTranslation delete translation into DB backend
func (backend *Backend) DeleteTranslation(t *i18nmod.Translation) error {
	return backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Delete(&Translation{}).Error
}
