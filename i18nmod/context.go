package i18nmod

import "context"

type Context interface {
	context.Context
	AddFoundHandler(func(handler *Handler, r *Result)) Context
	AddNotFoundHandler(func(handler *Handler, t *T)) Context
	AddHandler(HandlerFunc) Context
	Handler() *Handler
	Locales() []string
	T(key string) *T
	TT(key string) *T
	WithContext(ctx context.Context) Context
}

type Translater interface {
	Translate(ctx Context) string
}
