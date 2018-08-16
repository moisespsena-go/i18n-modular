package i18nmod

import (
	"strings"
	"fmt"
	html "html/template"
	"github.com/moisespsena/template/funcs"
)

type Key struct {
	Key        string
	GroupName  string
	Prev       *Key
	IsSingular bool
	IsPlural   bool
}

func (k *Key) Stringfy() string {
	return k.Key
}

func (k *Key) Plural() *Key {
	k.IsPlural = true
	return k
}

func (k *Key) Singular() *Key {
	k.IsSingular = true
	return k
}

func NewKey(key string, prev *Key) *Key {
	pos := strings.Index(key, ".")
	var groupName string

	if pos != -1 {
		groupName = key[0:pos]
	}

	var plural bool
	var singular bool
	if strings.HasSuffix(key, "+") {
		plural = true
		key = key[0:len(key)-1]
	} else if strings.HasSuffix(key, "~p") {
		plural = true
		key = key[0:len(key)-2]
	} else if strings.HasSuffix(key, "-") {
		singular = true
		key = key[0:len(key)-1]
	} else if strings.HasSuffix(key, "~s") {
		singular = true
		key = key[0:len(key)-2]
	}

	return &Key{Key: key, GroupName: groupName, Prev: prev, IsPlural: plural, IsSingular: singular}
}

type T struct {
	Handler          *Handler
	Langs            []string
	Key              *Key
	Group            string
	DefaultValue     string
	DataValue        interface{}
	CountValue       int
	AsTemplateResult bool
	funcMaps         []funcs.FuncMap
	funcValues       []*funcs.FuncValues
}

func NewT(context Context, key string) *T {
	return &T{Handler: context.Handler(), Langs: context.Langs(), Key: NewKey(key, nil), CountValue: -1,
		DefaultValue: key}
}

func (t *T) Funcs(funcMaps... funcs.FuncMap) *T {
	t.funcMaps = funcMaps
	return t
}
func (t *T) FuncValues(funcValues... *funcs.FuncValues) *T {
	t.funcValues = funcValues
	return t
}

func (t *T) Count(value int) *T {
	t.CountValue = value
	return t
}

func (t *T) Plural(value int) *T {
	t.Key.Plural()
	return t
}

func (t *T) Singular(value int) *T {
	t.Key.Singular()
	return t
}

func (t *T) With(key string) *T {
	if t.DefaultValue == t.Key.Key {
		t.DefaultValue = key
	}
	t.Key = NewKey(key, t.Key)
	return t
}

func (t *T) Default(value string) *T {
	t.DefaultValue = value
	return t
}

func (t *T) DefaultArgs(values... string) *T {
	if len(values) > 0 {
		t.DefaultValue = values[0]
	}
	return t
}

func (t *T) Data(value interface{}) *T {
	t.DataValue = value
	return t
}

func (t *T) DefaultAndDataFromArgs(args ... interface{}) *T {
	if len(args) > 0 {
		t.DataValue = args[0].(string)
		args = args[1:]

		if len(args) > 0 {
			t.DataValue = args[0]
			args = args[1:]
		}
	}

	if len(args) > 0 {
		panic("Invalid args.")
	}

	return t
}

func (t *T) AsTemplate() *T {
	t.AsTemplateResult = true
	return t
}

var FOLLOW = 5

func (t *T) Get() string {
	var r *Result
	for i := 0; i < FOLLOW; i++ {
		r = t.Handler.Handle(t)
		if r.Error != nil || r.Alias == "" {
			break
		}
		t.With(r.Alias)
	}

	if r.Error != nil {
		return fmt.Sprint("ERROR: ", r.Error)
	}

	return r.Text
}

func (t *T) GetText() string {
	return t.Get()
}

func (t *T) GetHtml() html.HTML {
	return html.HTML(t.Get())
}

type Result struct {
	Text        string
	Alias       string
	Error       error
	Translation *Translation
}
