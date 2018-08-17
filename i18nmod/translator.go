package i18nmod

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"

	"github.com/moisespsena/template/common"
	"gopkg.in/fatih/set.v0"
)

type Translator struct {
	Backends                 []Backend
	Groups                   map[string]map[string]map[string]*Translation
	ContextFactory           func(t *Translator, translate TranslateFunc, lang string, other_langs ...string) Context
	OnContextCreateCallbacks []func(context Context)
	Cache                    *Cache
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

func (t *Translator) LoadGroupTranslations(lang string, group string) (map[string]*Translation, error) {
	itemsMap := make(map[string]*Translation)

	for _, bc := range t.Backends {
		items, err := bc.LoadTranslations(lang, group)
		if err != nil {
			return nil, fmt.Errorf("Failed to load group '%v' translations of '%v' language: %v", group, lang, err)
		}
		for _, item := range items {
			item.Key = group + "." + item.Key
			if _, ok := itemsMap[item.Key]; !ok {
				itemsMap[item.Key] = item
			}
		}
	}

	return itemsMap, nil
}

func (t *Translator) PreloadAll() error {
	return t.Preload([]string{})
}

func (t *Translator) Preload(languages []string, names ...string) error {
	if len(languages) == 0 {
		mn := set.New(set.ThreadSafe)
		for _, backend := range t.Backends {
			for _, lang := range backend.ListLanguages() {
				mn.Add(lang)
			}
		}
		languages = make([]string, mn.Size())
		i := 0
		mn.Each(func(item interface{}) bool {
			languages[i] = item.(string)
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
		for _, lang := range languages {
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

func (t *Translator) NewContext(lang string, other_langs ...string) (c Context) {
	c = t.ContextFactory(t, t.Translate, lang, other_langs...)
	for _, cb := range t.OnContextCreateCallbacks {
		cb(c)
	}
	return c
}

func (t *Translator) Translate(tl *T) (r *Result) {
	key := tl.Key.Key
	r = &Result{Text: tl.DefaultValue}
	if r.Text == "" {
		r.Text = key
	}

	for _, lang := range tl.Langs {
		if group, ok := t.Groups[tl.Key.GroupName]; ok {
			if _, ok := group[lang]; ok {
				if tn, ok := t.Groups[tl.Key.GroupName][lang][key]; ok {
					tn.Translate(tl, r)
					return
				}
			}
		}
	}

	return
}

func (t *Translator) Dump(writer io.Writer) {
	d := &Dumper{Writer: writer}

	d.
		Wl("import (").
		With(func(d *Dumper) {
			d.
				Wl("\"", reflect.TypeOf(t).Elem().PkgPath(), "\"")
		}).
		Wl(")").
		Wl("func Groups() map[string]map[string]map[string]*i18nmod.Translation {").
		With(func(d *Dumper) {
			d.Wl("return map[string]map[string]map[string]*i18nmod.Translation {").With(func(d *Dumper) {
				for gname, langs := range t.Groups {
					d.Wl("\"", gname, "\": {").
						With(func(d *Dumper) {
							for lang, tls := range langs {
								d.Wl("\"", lang, "\": {").
									With(func(d *Dumper) {
										for _, tl := range tls {
											d.Wl("\"", tl.Key, "\": &i18nmod.Translation{").
												With(func(d *Dumper) {
													d.Wl("Key: \"", tl.Key, "\",")

													if tl.Value != "" {
														d.Wl("Value: ", strconv.Quote(tl.Value), ",")
													} else if tl.ValueTemplate != nil {
														d.Wl("ValueTemplate: i18nmod.ParseTemplate(", strconv.Quote(tl.ValueTemplate.Template().RawText()), "),")
													}

													if tl.Alias != "" {
														d.Wl("Alias: \"", tl.Alias, "\",")
													}

													if tl.PluralizeData != nil {
														d.Wl("PluralizeData: map[interface{}]interface{} {").
															With(func(d *Dumper) {
																for k, v := range tl.PluralizeData {
																	d.W()
																	switch kv := k.(type) {
																	case string:
																		d.R("\"", kv, "\":")
																	case int:
																		d.R(strconv.Itoa(kv), ":")
																	}

																	switch vv := v.(type) {
																	case string:
																		d.R(strconv.Quote(vv))
																	case common.TemplateExecutorInterface:
																		d.R("i18nmod.ParseTemplate(", strconv.Quote(vv.Template().RawText()), ")")
																	}

																	d.R(",\n")
																}
															}).
															Wl("},")
													}

												}).
												Wl("},")
										}
									}).
									Wl("},")
							}
						}).
						Wl("},")
				}
			}).
				Wl("}")
		}).
		Wl("}")
}
