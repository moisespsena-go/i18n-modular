package yaml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"path/filepath"
	"strconv"

	"github.com/moisespsena/go-i18n-modular/i18nmod"
	"github.com/moisespsena/template/text/template"
	"github.com/nicksnyder/go-i18n/i18n/language"
	"gopkg.in/fatih/set.v0"
	"gopkg.in/yaml.v2"
)

var _ i18nmod.Backend = &Backend{}

// New new YAML backend for I18n
func New() *Backend {
	return &Backend{make(map[string]map[string][]*File)}
}

type FileReader func(name string) ([]byte, error)

type File struct {
	Path   string
	Reader FileReader
}

// Backend YAML backend
type Backend struct {
	files map[string]map[string][]*File
}

func loadPluralizableValue(scope []string, parentkey string, value yaml.MapSlice) (map[interface{}]interface{}, error) {
	mp := make(map[interface{}]interface{})

	for _, e := range value {
		key := fmt.Sprint(e.Key)

		if strings.HasSuffix(key, "~") {
			v, err := template.New("").Parse(fmt.Sprint(e.Value))
			if err != nil {
				return nil, fmt.Errorf("Parse translation [%v.%v.%v] template failed: %v",
					strings.Join(scope, "."), parentkey, key, err)
			}

			key = key[0 : len(key)-1]

			if i, err := strconv.Atoi(key); err == nil {
				mp[i] = v.Executor()
			} else {
				mp[key] = v.Executor()
			}
		} else {
			mp[e.Key] = fmt.Sprint(e.Value)
		}
	}

	return mp, nil
}

func loadTranslationsFromYaml(file *string, value interface{}, scopes []string) (translations []*i18nmod.Translation, err error) {
	switch v := value.(type) {
	case yaml.MapSlice:
		for _, e := range v {
			key := fmt.Sprint(e.Key)

			if strings.HasSuffix(key, "*") {
				switch mps := e.Value.(type) {
				case yaml.MapSlice:
					key := key[0 : len(key)-1]
					pValue, err := loadPluralizableValue(scopes, key, mps)

					if err != nil {
						return nil, err
					}

					translations = append(translations, &i18nmod.Translation{
						Key:           strings.Join(append(scopes, key), "."),
						PluralizeData: pValue,
						Source:        file,
					})
				}
			} else {
				results, err := loadTranslationsFromYaml(file, e.Value, append(scopes, key))
				if err != nil {
					return results, err
				}
				translations = append(translations, results...)
			}
		}
	case string:
		key := scopes[len(scopes)-1]
		if strings.HasSuffix(key, "~") {
			key = key[0 : len(key)-1]
			scopes[len(scopes)-1] = key

			t, err := template.New("").Parse(v)
			if err != nil {
				return nil, fmt.Errorf("Parse translation [%v] template failed: %v",
					strings.Join(scopes, "."), err)
			}

			tpl := t.CreateExecutor()

			translations = append(translations, &i18nmod.Translation{
				Key:           strings.Join(scopes, "."),
				ValueTemplate: tpl,
				Source:        file,
			})
		} else if strings.HasSuffix(key, "@") {
			key = key[0 : len(key)-1]
			scopes[len(scopes)-1] = key

			translations = append(translations, &i18nmod.Translation{
				Key:    strings.Join(scopes, "."),
				Alias:  fmt.Sprint(v),
				Source: file,
			})
		} else {
			translations = append(translations, &i18nmod.Translation{
				Key:    strings.Join(scopes, "."),
				Value:  fmt.Sprint(v),
				Source: file,
			})
		}
	default:
		return []*i18nmod.Translation{}, fmt.Errorf("Invalid value of scope '%v': %v",
			strings.Join(scopes, "."), value)
	}
	return translations, nil
}

// LoadYAMLContent load YAML content
func (backend *Backend) LoadContent(file *string, content []byte) (translations []*i18nmod.Translation, err error) {
	var slice yaml.MapSlice

	if err = yaml.Unmarshal(content, &slice); err == nil {
		return loadTranslationsFromYaml(file, slice, []string{})
	}

	return translations, err
}

func (backend *Backend) LoadTranslations(language string, group string) ([]*i18nmod.Translation, error) {
	var translations []*i18nmod.Translation
	if gfiles, ok := backend.files[group]; ok {
		if files, ok := gfiles[language]; ok {
			for _, file := range files {
				if content, err := file.Reader(file.Path); err == nil {
					items, err := backend.LoadContent(&file.Path, content)

					if err != nil {
						return nil, fmt.Errorf("Load group '%v' of file '%v' failed: %v", group, file.Path, err)
					}

					translations = append(translations, items...)
				} else {
					return nil, fmt.Errorf("Load group '%v' of file '%v' failed: %v", group, file.Path, err)
				}
			}
		}
	}

	return translations, nil
}

// SaveTranslation save translation into YAML backend, not implemented
func (backend *Backend) SaveTranslation(t *i18nmod.Translation) error {
	return errors.New("not implemented")
}

// DeleteTranslation delete translation into YAML backend, not implemented
func (backend *Backend) DeleteTranslation(t *i18nmod.Translation) error {
	return errors.New("not implemented")
}

func (backend *Backend) GetFiles() map[string]map[string][]*File {
	return backend.files
}

func (backend *Backend) ListGroups() []string {
	keys := make([]string, len(backend.files))

	i := 0
	for k := range backend.files {
		keys[i] = k
		i++
	}

	return keys
}

func (backend *Backend) ListLanguages() (langs []string) {
	st := set.New(set.NonThreadSafe)
	for group := range backend.files {
		for lang := range backend.files[group] {
			st.Add(lang)
		}
	}

	st.Each(func(item interface{}) bool {
		langs = append(langs, item.(string))
		return true
	})

	return langs
}

func (backend *Backend) AddFileToGroup(group string, reader FileReader, files ...string) error {
	for _, f := range files {
		basename := filepath.Base(f)
		langs := language.Parse(basename)
		switch l := len(langs); {
		case l == 0:
			return fmt.Errorf("no language found in %v", f)
		case l > 1:
			return fmt.Errorf("multiple languages found in filename %v: %v; expected one", f, langs)
		}

		if _, ok := backend.files[group]; !ok {
			backend.files[group] = make(map[string][]*File)
		}

		lang := langs[0].Tag
		parts := strings.Split(lang, "-")

		if len(parts) > 1 {
			lang = parts[0] + "-" + strings.ToUpper(parts[1])
		}

		if _, ok := backend.files[group][lang]; !ok {
			backend.files[group][lang] = []*File{}
		}

		backend.files[group][lang] = append(backend.files[group][lang], &File{f, reader})
	}
	return nil
}

func (backend *Backend) LoadDir(path string) (errs []error) {
	err := i18nmod.Walk(".", path, func(group string, items []string) error {
		return backend.AddFileToGroup(i18nmod.FormatGroupName(group), ioutil.ReadFile, items...)
	})
	if err != nil {
		errs = append(errs, err)
	}
	return
}
