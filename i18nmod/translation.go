package i18nmod

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/moisespsena/template/text/template"
)

type Translation struct {
	Group         *string
	Key           string
	Value         string
	ValueTemplate *template.Executor
	Plural        *Plural
	Source        *string
	Alias         string
	TemplateCache *template.Executor
}

func (t *Translation) Pluralize(count interface{}, data interface{}) string {
	v, ok := t.Plural.Find(count)
	if !ok {
		return ""
	}

	if tpl, ok := v.(*template.Executor); ok {
		s, err := tpl.ExecuteString(data, map[string]interface{}{
			"count": func() interface{} {
				return count
			},
		})
		if err != nil {
			return fmt.Sprint("Translation template error [%v]: %v", t.Key, err)
		}
		return s
	}
	return v.(string)
}

func (t *Translation) PluralValue() interface{} {
	return t.Plural.MustFind("p")
}

func (t *Translation) SingularValue() interface{} {
	return t.Plural.MustFind("s")
}

func (t *Translation) Translate(context Context, lang string, tl *T, r *Result) {
	r.Translation = t

	if t.Alias != "" {
		alias := t.Alias
		if alias[0:1] == "." {
			alias = tl.Key.GroupName + alias
		}
		r.Alias = alias
		return
	}

	if t.Plural != nil {
		var value interface{}
		if tl.CountValue != nil {
			value = t.Pluralize(tl.CountValue, tl.DataValue)
		} else if tl.Key.IsSingular {
			value = t.SingularValue()
		} else if tl.Key.IsPlural {
			value = t.PluralValue()
		} else {
			r.Error = errors.New("error: isn't singular or Plural or not have count value")
			return
		}
		switch vt := value.(type) {
		case *template.Executor:
			var data interface{}
			if tfd, ok := tl.DataValue.(TemplateFuncsData); ok {
				data = tfd.Data()
				vt = vt.Funcs(tl.funcMaps...).Funcs(tfd.Funcs()).FuncsValues(tfd.FuncValues())
			} else {
				data = tl.DataValue
				vt = vt.Funcs(tl.funcMaps...).FuncsValues(tl.funcValues...)
			}

			var err error
			if r.value, err = vt.ExecuteString(data); err != nil {
				r.Error = fmt.Errorf("Execute template failed: %v", err)
			}
		default:
			r.value = vt.(string)
		}
		return
	} else if t.ValueTemplate != nil {
		var buf bytes.Buffer
		var err error

		if tfd, ok := tl.DataValue.(TemplateFuncsData); ok {
			err = t.ValueTemplate.Funcs(tfd.Funcs()).Execute(&buf, tfd.Data())
		} else {
			err = t.ValueTemplate.Funcs(tl.funcMaps...).Execute(&buf, tl.DataValue)
		}

		if err != nil {
			r.Error = err
			return
		}

		r.value = buf.String()
		return
	} else if tl.AsTemplateResult || t.ValueTemplate != nil {
		var tpl *template.Executor
		if t.TemplateCache != nil {
			tpl = t.TemplateCache
		} else if t.ValueTemplate == nil {
			tpl, err := template.New("").Parse(t.Value)
			if err != nil {
				r.Error = fmt.Errorf("Parse Value failed: %v", err)
				return
			}
			t.TemplateCache = tpl.CreateExecutor()
		} else {
			tpl = t.ValueTemplate
			t.TemplateCache = tpl
		}
		var buf bytes.Buffer
		var err error

		var executor *template.Executor
		var data interface{}

		if tfd, ok := tl.DataValue.(TemplateFuncsData); ok {
			data = tfd.Data()
			executor = tpl.Funcs(tl.funcMaps...).Funcs(tfd.Funcs()).FuncsValues(tfd.FuncValues())
		} else {
			data = tl.DataValue
			executor = tpl.Funcs(tl.funcMaps...).FuncsValues(tl.funcValues...)
		}

		err = executor.Execute(&buf, data)

		if err != nil {
			r.Error = fmt.Errorf("Execute template failed: %v", err)
			return
		}

		r.value = buf.String()
		return
	}
	r.value = t.Value
	return
}
