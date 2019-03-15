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
		var buf bytes.Buffer

		err := tpl.Execute(&buf, map[string]interface{}{
			"Count": count,
			"Data":  data,
		})

		if err != nil {
			return fmt.Sprint("Translation template error [%v]: %v", t.Key, err)
		}
		return buf.String()
	}
	return v.(string)
}

func (t *Translation) PluralValue() string {
	return t.Plural.MustFind("p").(string)
}

func (t *Translation) SingularValue() string {
	return t.Plural.MustFind("s").(string)
}

func (t *Translation) Translate(lang string, tl *T, r *Result) {
	r.Translation = t

	if t.Alias != "" {
		alias := t.Alias
		if alias[0:1] == "." {
			alias = tl.Key.GroupName + alias
		}
		r.Alias = alias
		return
	}

	if tl.AsTemplateResult {
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
	} else {
		if t.Plural != nil {
			if tl.CountValue != nil {
				r.value = t.Pluralize(tl.CountValue, tl.DataValue)
			} else if tl.Key.IsSingular {
				r.value = t.SingularValue()
			} else if tl.Key.IsPlural {
				r.value = t.PluralValue()
			} else {
				r.Error = errors.New("error: isn't singular or Plural or not have count value")
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
		}
		r.value = t.Value
	}
	return
}
