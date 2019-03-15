package i18nmod

import (
	"fmt"
)

type pKeyCount struct {
	cond   uint8
	value  string
	format string
}

func (k pKeyCount) Accept(value interface{}) bool {
	var s string
	if k.format == "" {
		s = fmt.Sprint(value)
	} else {
		s = fmt.Sprintf(k.format, value)
	}
	switch k.cond {
	case '=':
		return k.value == s
	case '>':
		return k.value > s
	case '<':
		return k.value < s
	default:
		return false
	}
}

type Plural struct {
	cases    map[interface{}]interface{}
	expCases map[pKeyCount]interface{}
}

func (p *Plural) AddCase(key, value interface{}) {
	switch kt := key.(type) {
	case string:
		switch kt[0] {
		case '=', '>', '<':
			if p.expCases == nil {
				p.expCases = map[pKeyCount]interface{}{}
			}
			v := kt[1:]
			var format string
			if v[0] == '%' {
				format = v[:2]
				v = v[2:]
			}
			p.expCases[pKeyCount{kt[0], v, format}] = value
			return
		}
	}

	if p.cases == nil {
		p.cases = map[interface{}]interface{}{}
	}
	switch key {
	case "p":
		if _, ok := p.cases["other"]; !ok {
			p.cases["other"] = value
		}
	case "s":
		if _, ok := p.cases["one"]; !ok {
			p.cases["one"] = value
		}
	}
	p.cases[key] = value
}

func (p *Plural) MustFind(count interface{}) (v interface{}) {
	v, _ = p.Find(count)
	return
}

func (p *Plural) Find(count interface{}) (v interface{}, ok bool) {
	if p.cases != nil {
		if v, ok = p.cases[count]; ok {
			return
		}
	}
	if p.expCases != nil {
		var k pKeyCount
		for k, v = range p.expCases {
			if k.Accept(count) {
				ok = true
				return
			}
		}
	}

	if p.cases != nil {
		if v, ok = p.cases["other"]; ok {
			return
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
