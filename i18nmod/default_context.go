package i18nmod

import (
	"context"
)

type DefaultContext struct {
	context.Context
	Translator       *Translator
	locales          []string
	Groups           map[string]map[string]DB
	FoundHandlers    []func(handler *Handler, r *Result)
	NotFoundHandlers []func(handler *Handler, t *T)
	cache            map[string]*Result
	handler          *Handler
	LogOkEnabled     bool
	LogFaultEnabled  bool
}

func (c *DefaultContext) AddHandler(fn HandlerFunc) Context {
	c.handler = &Handler{Context: c, Prev: c.handler, Handler: fn}
	return c
}

func (c *DefaultContext) Handler() *Handler {
	return c.handler
}

func (c *DefaultContext) T(key string) *T {
	return NewT(c, key)
}

func (c *DefaultContext) TT(key string) *T {
	return NewT(c, key).AsTemplate()
}

func (c *DefaultContext) Locales() []string {
	return c.locales
}

func (c *DefaultContext) AddFoundHandler(handler func(handler *Handler, r *Result)) Context {
	c.FoundHandlers = append(c.FoundHandlers, handler)
	return c
}

func (c *DefaultContext) AddNotFoundHandler(handler func(handler *Handler, t *T)) Context {
	c.NotFoundHandlers = append(c.NotFoundHandlers, handler)
	return c
}

func (c DefaultContext) WithContext(ctx context.Context) Context {
	c.Context = ctx
	return &c
}

func DefaultContextFactory(t *Translator, translate TranslateFunc, lang string, defaultLocale ...string) Context {
	var (
		locales = []string{lang}
	)

	if len(defaultLocale) > 0 && defaultLocale[0] != "" {
		if defaultLocale[0] != lang {
			locales = append(locales, defaultLocale[0])
		}
	} else if t.DefaultLocale != "" {
		if t.DefaultLocale != lang {
			locales = append(locales, t.DefaultLocale)
		}
	}

	c := &DefaultContext{
		Context:    context.Background(),
		Translator: t,
		locales:    locales,
		Groups:     t.Groups,
		cache:      map[string]*Result{},
	}

	c.AddHandler(func(handler *Handler, tl *T) (r *Result) {
		if tl.Key.Cached {
			if r = c.cache[tl.Key.Key]; r != nil {
				return
			}
		}
		r = translate(handler.Context, tl)
		if r.Translation == nil {
			for _, h := range c.NotFoundHandlers {
				h(handler, tl)
			}
		} else {
			if tl.Key.Cached {
				c.cache[tl.Key.Key] = r
			}
			for _, h := range c.FoundHandlers {
				h(handler, r)
			}
		}
		return
	})
	return c
}
