package i18nmod

import (
	"fmt"
	"sync"

	"gopkg.in/fatih/set.v0"

	"github.com/moisespsena/template/html/template"
)

type DB map[string]*Translation

type ChildDB struct {
	Group string
	Prefix string
	DB     DB
}

func (this ChildDB) With(prefix string, cb func(db *ChildDB)) {
	this.Prefix += prefix + "."
	cb(&this)
}

func (this *ChildDB) Get(key string) *Translation {
	return this.DB[this.Prefix+key]
}

func (this *ChildDB) Set(t ...*Translation) *ChildDB {
	for _, t := range t {
		func(t Translation) {
			t.Group = &this.Group
			if this.Prefix != "" {
				t.Key = this.Prefix + t.Key
			}
			this.DB[t.Key] = &t
		}(*t)
	}
	return this
}

type Translator struct {
	Backends                 []Backend
	Groups                   map[string]map[string]DB
	ContextFactory           func(t *Translator, translate TranslateFunc, locale string, defaultLocale ...string) Context
	OnContextCreateCallbacks []func(context Context)
	Cache                    *Cache
	Locales                  []string
	DefaultLocale            string
	sync.RWMutex
	preloaded           map[string]bool
	groupLoadedCallback map[string][]func(lang string, db *ChildDB)
}

func NewTranslator() *Translator {
	return &Translator{
		Backends:                 []Backend{},
		ContextFactory:           DefaultContextFactory,
		Cache:                    NewCache(),
		OnContextCreateCallbacks: []func(context Context){},
		groupLoadedCallback:      make(map[string][]func(lang string, db *ChildDB)),
		Groups:                   make(map[string]map[string]DB),
	}
}

func (t *Translator) AfterGroupLoad(groupName string, cb func(lang string, db *ChildDB)) {
	t.groupLoadedCallback[groupName] = append(t.groupLoadedCallback[groupName], cb)
	if data, ok := t.Groups[groupName]; ok {
		for lang, db := range data {
			cb(lang, &ChildDB{groupName, "", db})
		}
	}
}

func (t *Translator) OnContextCreate(callbacks ...func(context Context)) *Translator {
	t.OnContextCreateCallbacks = append(t.OnContextCreateCallbacks, callbacks...)
	return t
}

func (t *Translator) AddBackend(backends ...Backend) {
	t.Backends = append(t.Backends, backends...)
}

func (tr *Translator) LoadGroupTranslations(locale string, group string) (items DB, err error) {
	tree := &Tree{}

	for _, bc := range tr.Backends {
		t, err := bc.LoadTranslations(locale, group)
		if err != nil {
			return nil, fmt.Errorf("Failed to load group '%v' translations of '%v' locale: %v", group, locale, err)
		}
		tree.Merge(t)
	}

	items = make(DB)

	_ = tree.WalkT(func(key string, t *Translation) error {
		t.Key = key
		t.Group = &group
		items[key] = t
		return nil
	})

	return
}

func (tr *Translator) SetGroup(lang string, group string, tree *Tree) {
	items := make(DB)

	_ = tree.WalkT(func(key string, t *Translation) error {
		t.Key = key
		t.Group = &group
		items[key] = t
		return nil
	})

	if _, ok := tr.Groups[group]; !ok {
		tr.Groups[group] = map[string]DB{}
	}
	tr.Groups[group][lang] = items

	if callbacks := tr.groupLoadedCallback[group]; callbacks != nil {
		for _, cb := range callbacks {
			cb(lang, &ChildDB{Group: group, DB: tr.Groups[group][lang]})
		}
	}
}

func (tr *Translator) NewGroup(locale string, group string, cb func(t *Tree)) {
	var tree Tree
	cb(&tree)
	tr.SetGroup(locale, group, &tree)
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
		if t.preloaded == nil {
			t.preloaded = map[string]bool{}
		}
		mn.Each(func(item interface{}) bool {
			name := item.(string)
			if _, ok := t.preloaded[name]; ok {
				return true
			}
			t.preloaded[name] = true
			names[i] = name
			i++
			return true
		})
	}

	for _, name := range names {
		if _, ok := t.Groups[name]; !ok {
			t.Groups[name] = make(map[string]DB)
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

			if callbacks := t.groupLoadedCallback[name]; callbacks != nil {
				for _, cb := range callbacks {
					cb(lang, &ChildDB{Group: name, DB: t.Groups[name][lang]})
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

func (t *Translator) Translate(context Context, tl *T) (r *Result) {
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
					tn.Translate(context, lang, tl, r)
					return
				}
			}
		}
	}

	if tl.DefaultValue != nil {
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
		} else {
			var exec *template.Executor
			switch dvt := tl.DefaultValue.(type) {
			case string:
				r.value = dvt
				return
			case template.HTML:
				r.value = dvt
				return
			case *template.Template:
				exec = dvt.CreateExecutor()
			case *template.Executor:
				exec = dvt
			case func() string:
				r.value = dvt()
				return
			case func() *T:
				r.value = dvt().Get()
				return
			case func() template.HTML:
				r.value = dvt()
				return
			case func() *template.Executor:
				exec = dvt()
			case func(Context) *T:
				r.value = dvt(context).Get()
				return
			case func(Context) template.HTML:
				r.value = dvt(context)
				return
			case func(Context) *template.Executor:
				exec = dvt(context)
			default:
				r.Error = fmt.Errorf("Invalid default value of %q as %T = %s", tl.Key.Key, tl.DefaultValue, tl.DefaultValue)
				return
			}
			dataValue := tl.DataValue
			if dataValue == nil {
				dataValue = map[interface{}]interface{}{}
			}
			data, err := exec.ExecuteString(tl.DataValue)
			if err != nil {
				r.Error = err
				return
			}
			r.value = data
		}
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
