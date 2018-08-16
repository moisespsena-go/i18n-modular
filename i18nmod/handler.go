package i18nmod

type HandlerFunc func(handler *Handler, t *T) *Result

type Handler struct {
	Context Context
	Prev    *Handler
	Handler HandlerFunc
}

func (h *Handler) Handle(t *T) *Result {
	return h.Handler(h, t)
}
