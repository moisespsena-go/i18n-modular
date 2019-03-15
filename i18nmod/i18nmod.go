package i18nmod

import "github.com/moisespsena/template/funcs"

type Group struct {
	Name  string
	Items map[string]*Translation
}

type ContextGroup struct {
	Group
	Translator *Translator
}

type TranslateFunc func(tl *T) *Result

type TemplateFuncsData interface {
	Funcs() map[string]interface{}
	FuncValues() *funcs.FuncValues
	Data() interface{}
}
