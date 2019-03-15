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
	return &Backend{make(map[string]map[string][]*Input)}
}

type FileReader func(name string) ([]byte, error)
type Reader func() ([]byte, error)

type Input struct {
	Name   string
	Reader Reader
}

// Backend YAML backend
type Backend struct {
	inputs map[string]map[string][]*Input
}

func mapToPlural(scope []string, parentkey string, value yaml.MapSlice) (*i18nmod.Plural, error) {
	var plural = &i18nmod.Plural{}

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
				plural.AddCase(i, v.Executor())
			} else {
				plural.AddCase(key, v.Executor())
			}
		} else {
			plural.AddCase(e.Key, e.Value)
		}
	}

	return plural, nil
}

type importer struct {
	tree  i18nmod.Tree
	links map[string]string
}

func (i *importer) Add(t ...*i18nmod.Translation) {
	for _, t := range t {
		i.tree.Add(t.Key, t)
	}
}

func (i *importer) Link(from, to string) {
	i.tree.Tree(from).Link(to)
}

func (i *importer) Import(name *string, value interface{}, scopes []string) (err error) {
	switch v := value.(type) {
	case yaml.MapSlice:
		for _, e := range v {
			key := fmt.Sprint(e.Key)

			if strings.HasSuffix(key, "*") {
				switch mps := e.Value.(type) {
				case yaml.MapSlice:
					key := key[0 : len(key)-1]
					plural, err := mapToPlural(scopes, key, mps)

					if err != nil {
						return err
					}

					i.Add(&i18nmod.Translation{
						Key:    strings.Join(append(scopes, key), "."),
						Plural: plural,
						Source: name,
					})
				}
			} else {
				if err = i.Import(name, e.Value, append(scopes, key)); err != nil {
					return
				}
			}
		}
	case [][]string:
		key := scopes[len(scopes)-1]
		if strings.HasSuffix(key, "*") {
			key = key[0 : len(key)-1]
			scopes[len(scopes)-1] = key
			plural := &i18nmod.Plural{}

			for i, value := range v {
				var (
					k             = value[0]
					v interface{} = value[1]
				)

				if strings.HasSuffix(k, "~") {
					k = k[0 : len(k)-1]
					t, err := template.New("").Parse(value[1])
					if err != nil {
						return fmt.Errorf("Parse translation [%v][%d][1] template failed: %v",
							strings.Join(scopes, "."), i, err)
					}
					v = t.CreateExecutor()
				}

				plural.AddCase(k, v)
			}

			i.Add(&i18nmod.Translation{
				Key:    strings.Join(append(scopes, key), "."),
				Plural: plural,
				Source: name,
			})
		}
	case [][]interface{}:
		key := scopes[len(scopes)-1]
		if strings.HasSuffix(key, "*") {
			key = key[0 : len(key)-1]
			scopes[len(scopes)-1] = key
			plural := &i18nmod.Plural{}

			for i, value := range v {
				var (
					k = value[0]
					v = value[1]
				)

				if s, ok := k.(string); ok {
					if strings.HasSuffix(s, "~") {
						k = s[0 : len(s)-1]
						t, err := template.New("").Parse(v.(string))
						if err != nil {
							return fmt.Errorf("Parse translation [%v][%d][1] template failed: %v",
								strings.Join(scopes, "."), i, err)
						}
						v = t.CreateExecutor()
						k = s
					}
				}

				plural.AddCase(k, v)
			}

			i.Add(&i18nmod.Translation{
				Key:    strings.Join(append(scopes, key), "."),
				Plural: plural,
				Source: name,
			})
		}
	case string:
		key := scopes[len(scopes)-1]
		if strings.HasSuffix(key, "~") {
			key = key[0 : len(key)-1]
			scopes[len(scopes)-1] = key

			t, err := template.New("").Parse(v)
			if err != nil {
				return fmt.Errorf("Parse translation [%v] template failed: %v",
					strings.Join(scopes, "."), err)
			}

			tpl := t.CreateExecutor()

			i.Add(&i18nmod.Translation{
				Key:           strings.Join(scopes, "."),
				ValueTemplate: tpl,
				Source:        name,
			})
		} else if strings.HasSuffix(key, "@") {
			key = key[0 : len(key)-1]
			scopes[len(scopes)-1] = key

			i.Add(&i18nmod.Translation{
				Key:    strings.Join(scopes, "."),
				Alias:  fmt.Sprint(v),
				Source: name,
			})
		} else if strings.HasSuffix(key, "&") {
			key = key[0 : len(key)-1]
			i.Link(key, v)
		} else {
			i.Add(&i18nmod.Translation{
				Key:    strings.Join(scopes, "."),
				Value:  fmt.Sprint(v),
				Source: name,
			})
		}
	default:
		return fmt.Errorf("Invalid value of scope '%v': %v",
			strings.Join(scopes, "."), value)
	}
	return
}

// LoadYAMLContent load YAML content
func (backend *Backend) LoadContent(name *string, content []byte) (tree *i18nmod.Tree, err error) {
	var slice yaml.MapSlice

	if err = yaml.Unmarshal(content, &slice); err == nil {
		imp := &importer{links: map[string]string{}}
		err = imp.Import(name, slice, []string{})
		if err != nil {
			return
		}
		tree = &imp.tree
	}

	return
}

func (backend *Backend) LoadTranslations(language string, group string) (*i18nmod.Tree, error) {
	tree := &i18nmod.Tree{}

	if gfiles, ok := backend.inputs[group]; ok {
		if inputs, ok := gfiles[language]; ok {
			for _, input := range inputs {
				if content, err := input.Reader(); err == nil {
					t, err := backend.LoadContent(&input.Name, content)
					if err != nil {
						return nil, fmt.Errorf("Load group '%v' of input '%v' failed: %v", group, input.Name, err)
					}
					tree.Merge(t)
				} else {
					return nil, fmt.Errorf("Load group '%v' of input '%v' failed: %v", group, input.Name, err)
				}
			}
		}
	}

	return tree, nil
}

// SaveTranslation save translation into YAML backend, not implemented
func (backend *Backend) SaveTranslation(t *i18nmod.Translation) error {
	return errors.New("not implemented")
}

// DeleteTranslation delete translation into YAML backend, not implemented
func (backend *Backend) DeleteTranslation(t *i18nmod.Translation) error {
	return errors.New("not implemented")
}

func (backend *Backend) GetFiles() map[string]map[string][]*Input {
	return backend.inputs
}

func (backend *Backend) ListGroups() []string {
	keys := make([]string, len(backend.inputs))

	i := 0
	for k := range backend.inputs {
		keys[i] = k
		i++
	}

	return keys
}

func (backend *Backend) ListLanguages() (langs []string) {
	st := set.New(set.NonThreadSafe)
	for group := range backend.inputs {
		for lang := range backend.inputs[group] {
			st.Add(lang)
		}
	}

	st.Each(func(item interface{}) bool {
		langs = append(langs, item.(string))
		return true
	})

	return langs
}

func (backend *Backend) AddFileToGroup(group string, reader Reader, files ...string) error {
	for _, f := range files {
		lang := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(f), ".yaml"), ".yml")
		return backend.addInput("file", group, lang, reader)
	}
	return nil
}

func (backend *Backend) LoadDir(path string) (errs []error) {
	err := i18nmod.WalkDir(".", path, func(group string, items []string) error {
		return backend.AddFileToGroup(i18nmod.FormatGroupName(group), func() (bytes []byte, e error) {
			return ioutil.ReadFile(path)
		}, items...)
	})
	if err != nil {
		errs = append(errs, err)
	}
	return
}

func (backend *Backend) AddInput(group, lang string, reader func() ([]byte, error)) (err error) {
	return backend.addInput("raw", group, lang, reader)
}

func (backend *Backend) addInput(typ, group, lang string, reader func() ([]byte, error)) (err error) {
	if lang != i18nmod.AnyLang {
		langs := language.Parse(lang)
		if l := len(langs); l == 0 || l > 1 {
			return fmt.Errorf("Invalid language name %q", lang)
		}
		lang = langs[0].Tag
		parts := strings.Split(lang, "-")

		if len(parts) > 1 {
			lang = parts[0] + "-" + strings.ToUpper(parts[1])
		}
	}

	if _, ok := backend.inputs[group]; !ok {
		backend.inputs[group] = make(map[string][]*Input)
	}

	if _, ok := backend.inputs[group][lang]; !ok {
		backend.inputs[group][lang] = []*Input{}
	}

	if typ != "" {
		typ = "+" + typ
	}

	backend.inputs[group][lang] = append(backend.inputs[group][lang], &Input{"yaml" + typ + "://" + group + "[" + lang + "]", reader})
	return nil
}
