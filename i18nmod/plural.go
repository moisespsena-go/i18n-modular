package i18nmod

import (
	"fmt"
)

type PluralKeyCount struct {
	Cond   uint8
	Value  string
	Format string
}

func (k PluralKeyCount) Accept(value interface{}) bool {
	var s string
	if k.Format == "" {
		s = fmt.Sprint(value)
	} else {
		s = fmt.Sprintf(k.Format, value)
	}
	switch k.Cond {
	case '=':
		return k.Value == s
	case '>':
		return k.Value > s
	case '<':
		return k.Value < s
	default:
		return false
	}
}

type Plural struct {
	Cases    map[interface{}]interface{}
	ExpCases map[PluralKeyCount]interface{}
}

func (p *Plural) AddCase(key, value interface{}) {
	switch kt := key.(type) {
	case string:
		switch kt[0] {
		case '=', '>', '<':
			if p.ExpCases == nil {
				p.ExpCases = map[PluralKeyCount]interface{}{}
			}
			v := kt[1:]
			var format string
			if v[0] == '%' {
				format = v[:2]
				v = v[2:]
			}
			p.ExpCases[PluralKeyCount{kt[0], v, format}] = value
			return
		}
	}

	if p.Cases == nil {
		p.Cases = map[interface{}]interface{}{}
	}
	switch key {
	case "p":
		if _, ok := p.Cases["other"]; !ok {
			p.Cases["other"] = value
		}
	case "s":
		if _, ok := p.Cases["one"]; !ok {
			p.Cases["one"] = value
		}
	}
	p.Cases[key] = value
}

func (p *Plural) SetCase(key, value interface{}) {
	switch kt := key.(type) {
	case string:
		switch kt[0] {
		case '=', '>', '<':
			if p.ExpCases == nil {
				p.ExpCases = map[PluralKeyCount]interface{}{}
			}
			v := kt[1:]
			var format string
			if v[0] == '%' {
				format = v[:2]
				v = v[2:]
			}
			p.ExpCases[PluralKeyCount{kt[0], v, format}] = value
			return
		}
	}

	if p.Cases == nil {
		p.Cases = map[interface{}]interface{}{}
	}
	switch key {
	case "p":
		old := p.Cases[key]
		if o, ok := p.Cases["other"]; !ok || (old != nil && o == old) {
			p.Cases["other"] = value
		}
	case "s":
		old := p.Cases[key]
		if o, ok := p.Cases["one"]; !ok || (old != nil && o == old) {
			p.Cases["one"] = value
		}
	}
	p.Cases[key] = value
}

func (p *Plural) MustFind(count interface{}) (v interface{}) {
	v, _ = p.Find(count)
	return
}

func (p *Plural) Find(count interface{}) (v interface{}, ok bool) {
	if p.Cases != nil {
		if v, ok = p.Cases[count]; ok {
			return
		}
	}
	if p.ExpCases != nil {
		var k PluralKeyCount
		for k, v = range p.ExpCases {
			if k.Accept(count) {
				ok = true
				return
			}
		}
	}

	if p.Cases != nil {
		switch count {
		case 1:
			if v, ok = p.Cases["one"]; ok {
				return
			}
		default:
			if v, ok = p.Cases["other"]; ok {
				return
			}
		}
	}
	v = nil
	return
}

func ParsePlural(data interface{}) *Plural {
	p := &Plural{}
	switch d := data.(type) {
	case [][]string:
		for _, pair := range d {
			p.AddCase(pair[0], pair[1])
		}
	case [][]interface{}:
		for _, pair := range d {
			p.AddCase(pair[0], pair[1])
		}
	case map[string]string:
		for key, value := range d {
			p.AddCase(key, value)
		}
	case map[string]interface{}:
		for key, value := range d {
			p.AddCase(key, value)
		}
	case map[interface{}]interface{}:
		for key, value := range d {
			p.AddCase(key, value)
		}
	}
	return p
}
