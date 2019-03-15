package i18nmod

type Context interface {
	AddFoundHandler(func(handler *Handler, r *Result)) Context
	AddNotFoundHandler(func(handler *Handler, t *T)) Context
	AddHandler(HandlerFunc) Context
	Handler() *Handler
	Locales() []string
	T(key string) *T
	TT(key string) *T
}
