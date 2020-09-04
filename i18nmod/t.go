package i18nmod

import (
	"fmt"
	html "html/template"
	"strings"

	"github.com/moisespsena/template/funcs"
)

type Key struct {
	Key        string
	GroupName  string
	Prev       *Key
	IsSingular bool
	IsPlural   bool
	Cached     bool
}

func (k *Key) Stringfy() string {
	return k.Key
}

func (k *Key) Name() string {
	if k.GroupName == "" {
		return k.Key
	}
	return strings.TrimPrefix(k.Key, k.GroupName+".")
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
	var cached bool
	if key[0] == '^' {
		cached = true
		key = key[1:]
	}
	pos := strings.Index(key, ".")
	var groupName string

	if pos != -1 {
		groupName = key[0:pos]
	}

	var plural bool
	var singular bool
	if strings.HasSuffix(key, "+") {
		plural = true
		key = key[0 : len(key)-1]
	} else if strings.HasSuffix(key, "~p") {
		plural = true
		key = key[0 : len(key)-2]
	} else if strings.HasSuffix(key, "-") {
		singular = true
		key = key[0 : len(key)-1]
	} else if strings.HasSuffix(key, "~s") {
		singular = true
		key = key[0 : len(key)-2]
	}

	return &Key{
		Key:        key,
		GroupName:  groupName,
		Prev:       prev,
		IsPlural:   plural,
		IsSingular: singular,
		Cached:     cached,
	}
}

type T struct {
	Handler          *Handler
	Locales          []string
	Key              *Key
	Group            string
	DefaultValue     interface{}
	DataValue        interface{}
	CountValue       interface{}
	AsTemplateResult bool
	funcMaps         []funcs.FuncMap
	funcValues       []funcs.FuncValues
}

func NewT(context Context, key string) *T {
	return &T{
		Handler:      context.Handler(),
		Locales:      context.Locales(),
		Key:          NewKey(key, nil),
		DefaultValue: key,
	}
}

func (t *T) Funcs(funcMaps ...funcs.FuncMap) *T {
	t.funcMaps = funcMaps
	return t
}
func (t *T) FuncValues(funcValues ...funcs.FuncValues) *T {
	t.funcValues = funcValues
	return t
}

func (t *T) Count(value interface{}) *T {
	t.CountValue = value
	return t
}

func (t *T) Plural(value interface{}) *T {
	t.CountValue = value
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

func (t *T) Default(value interface{}) *T {
	t.DefaultValue = value
	return t
}

func (t *T) DefaultArgs(values ...interface{}) *T {
	if len(values) > 0 {
		t.DefaultValue = values[0]
	}
	return t
}

func (t *T) Data(value interface{}) *T {
	t.DataValue = value
	return t
}

func (t *T) DefaultAndDataFromArgs(args ...interface{}) *T {
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

	if r.value == nil {
		if r.defaultValue != nil {
			switch dvt := r.defaultValue.(type) {
			case func() string:
				r.value = dvt()
			default:
				r.value = dvt
			}
		}
	}

	if r.value == nil {
		return ""
	}

	if s, ok := r.value.(string); ok {
		return s
	}

	return fmt.Sprint(r.value)
}

func (t *T) GetText() string {
	return t.Get()
}

func (t *T) GetHtml() html.HTML {
	return html.HTML(t.Get())
}

func (t *T) String() string {
	return t.Get()
}

type Result struct {
	defaultValue interface{}
	value        interface{}
	Alias        string
	Error        error
	Translation  *Translation
}

func Cached(key string) string {
	return "^" + key
}
