package i18nmod

import (
	"bytes"
	"fmt"
	"errors"

	"github.com/moisespsena/template/text/template"
)

type Translation struct {
	Key           string
	Value         string
	ValueTemplate *template.Executor
	PluralizeData map[interface{}]interface{}
	Source        *string
	Alias         string
	TemplateCache *template.Executor
}


func (t *Translation) PluralizeValue(count int, data interface{}) string {
	v, ok := t.PluralizeData[count]
	if !ok {
		v = t.PluralizeData["_"]
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

func (t *Translation) Plural() string {
	if v, ok := t.PluralizeData["p"]; ok {
		return v.(string)
	}
	return t.PluralizeData["_"].(string)
}

func (t *Translation) Singular() string {
	if v, ok := t.PluralizeData["s"]; ok {
		return v.(string)
	}
	return t.PluralizeData["_"].(string)
}

func (t *Translation) Translate(tl *T, r *Result) {
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

		r.Text = buf.String()
		return
	} else {
		if t.PluralizeData != nil {
			if tl.CountValue != -1 {
				r.Text = t.PluralizeValue(tl.CountValue, tl.DefaultValue)
			} else if tl.Key.IsSingular {
				r.Text = t.Singular()
				return
			} else if tl.Key.IsPlural {
				r.Text = t.Plural()
				return
			}
			r.Error = errors.New("error: isn't singular or plural or not have count value")
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

			r.Text = buf.String()
			return
		}
		r.Text = t.Value
	}
	return
}
