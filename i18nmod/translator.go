package i18nmod

import (
	"fmt"
	"sync"

	"github.com/moisespsena/template/html/template"
	"gopkg.in/fatih/set.v0"
)

type Translator struct {
	Backends                 []Backend
	Groups                   map[string]map[string]map[string]*Translation
	ContextFactory           func(t *Translator, translate TranslateFunc, locale string, defaultLocale ...string) Context
	OnContextCreateCallbacks []func(context Context)
	Cache                    *Cache
	Locales                  []string
	DefaultLocale            string
	sync.RWMutex
}

func NewTranslator() *Translator {
	return &Translator{
		Backends:       []Backend{},
		ContextFactory: DefaultContextFactory,
		Cache:          NewCache(),
		OnContextCreateCallbacks: []func(context Context){},
	}
}

func (t *Translator) OnContextCreate(callbacks ...func(context Context)) *Translator {
	t.OnContextCreateCallbacks = append(t.OnContextCreateCallbacks, callbacks...)
	return t
}

func (t *Translator) AddBackend(backends ...Backend) {
	t.Backends = append(t.Backends, backends...)
}

func (tr *Translator) LoadGroupTranslations(locale string, group string) (items map[string]*Translation, err error) {
	tree := &Tree{}

	for _, bc := range tr.Backends {
		t, err := bc.LoadTranslations(locale, group)
		if err != nil {
			return nil, fmt.Errorf("Failed to load group '%v' translations of '%v' locale: %v", group, locale, err)
		}
		tree.Merge(t)
	}

	items = map[string]*Translation{}

	_ = tree.WalkT(func(key string, t *Translation) error {
		t.Key = key
		t.Group = &group
		items[key] = t
		return nil
	})

	return
}

func (t *Translator) PreloadAll() error {
	return t.Preload([]string{})
}

func (t *Translator) Preload(locales []string, names ...string) error {
	if len(locales) == 0 {
		mn := set.New(set.ThreadSafe)
		for _, backend := range t.Backends {
			for _, lang := range backend.ListLanguages() {
				mn.Add(lang)
			}
		}
		locales = make([]string, mn.Size())
		i := 0
		mn.Each(func(item interface{}) bool {
			locales[i] = item.(string)
			i++
			return true
		})
	}

	if len(names) == 0 {
		mn := set.New(set.NonThreadSafe)
		for _, backend := range t.Backends {
			for _, name := range backend.ListGroups() {
				mn.Add(name)
			}
		}
		names = make([]string, mn.Size())
		i := 0
		mn.Each(func(item interface{}) bool {
			names[i] = item.(string)
			i++
			return true
		})
	}

	if t.Groups == nil {
		t.Groups = make(map[string]map[string]map[string]*Translation)
	}

	for _, name := range names {
		if _, ok := t.Groups[name]; !ok {
			t.Groups[name] = make(map[string]map[string]*Translation)
		}
		for _, lang := range locales {
			items, err := t.LoadGroupTranslations(lang, name)

			if err != nil {
				return err
			}

			if _, ok := t.Groups[name][lang]; !ok {
				t.Groups[name][lang] = items
			} else {
				for k, tr := range items {
					if _, ok := t.Groups[name][lang][k]; !ok {
						t.Groups[name][lang][k] = tr
					}
				}
			}

		}
	}
	return nil
}

func (t *Translator) NewContext(lang string, defaultLocale ...string) (c Context) {
	c = t.ContextFactory(t, t.Translate, lang, append(defaultLocale, AnyLang)...)
	for _, cb := range t.OnContextCreateCallbacks {
		cb(c)
	}
	return c
}

func (t *Translator) Translate(tl *T) (r *Result) {
	r = &Result{}
	if tl.DefaultValue != nil {
		r.defaultValue = tl.DefaultValue
	} else {
		r.defaultValue = tl.Key.Key
	}

	name := tl.Key.Name()

	for _, lang := range tl.Locales {
		if group, ok := t.Groups[tl.Key.GroupName]; ok {
			if data, ok := group[lang]; ok {
				if tn, ok := data[name]; ok {
					tn.Translate(lang, tl, r)
					return
				}
			}
		}
	}

	if tl.AsTemplateResult {
		var exec *template.Executor
		switch dvt := tl.DefaultValue.(type) {
		case string:
			tpl, err := template.New(tl.Key.Key).Parse(dvt)
			if err != nil {
				r.Error = err
				return
			}
			exec = tpl.CreateExecutor()
		case *template.Template:
			exec = dvt.CreateExecutor()
		case *template.Executor:
			exec = dvt
		default:
			r.Error = fmt.Errorf("Invalid template value of %q", tl.Key.Key)
			return
		}
		data, err := exec.ExecuteString(tl.DataValue)
		if err != nil {
			r.Error = err
			return
		}
		r.value = data
	}

	return
}

func (tr *Translator) ValidOrDefaultLocale(l string) string {
	if l != "" {
		for _, loc := range tr.Locales {
			if l == loc {
				return l
			}
		}
	}
	return tr.DefaultLocale
}
