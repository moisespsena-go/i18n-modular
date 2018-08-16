package i18nmod

type DefaultContext struct {
	Translator *Translator
	Languages  []string
	Groups     *map[string]map[string]map[string]*Translation
	FoundHandlers       []func(handler *Handler, r *Result)
	NotFoundHandlers    []func(handler *Handler, t *T)
	handler   *Handler
	LogOkEnabled bool
	LogFaultEnabled bool
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

func (c *DefaultContext) Langs() []string {
	return c.Languages
}

func (c *DefaultContext) AddFoundHandler(handler func(handler *Handler, r *Result)) Context {
	c.FoundHandlers = append(c.FoundHandlers, handler)
	return c
}

func (c *DefaultContext) AddNotFoundHandler(handler func(handler *Handler, t *T)) Context {
	c.NotFoundHandlers = append(c.NotFoundHandlers, handler)
	return c
}


func DefaultContextFactory(t *Translator, translate TranslateFunc, lang string, other_langs ... string) Context {
	c := &DefaultContext{
		Translator: t,
		Languages: append([]string{lang}, other_langs...),
		Groups: &t.Groups,
	}
	c.AddHandler(func(handler *Handler, tl *T) *Result {
		r :=  translate(tl)
		if r.Translation == nil {
			for _, h := range c.NotFoundHandlers {
				h(handler, tl)
			}
		} else {
			for _, h := range c.FoundHandlers {
				h(handler, r)
			}
		}
		return r
	})
	return c
}