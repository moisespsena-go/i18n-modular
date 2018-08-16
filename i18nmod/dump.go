package i18nmod

import (
	"io"
	"strings"
)

type Dumper struct {
	Writer io.Writer
	Spaces int
}

func (d *Dumper) R(args... string) *Dumper {
	for _, a := range args {
		d.Writer.Write([]byte(a))
	}
	return d
}

func (d *Dumper) W(args... string) *Dumper {
	if d.Spaces > 0 {
		d.Writer.Write([]byte(strings.Repeat("    ", d.Spaces)))
	}

	for _, a := range args {
		d.Writer.Write([]byte(a))
	}
	return d
}

func (d *Dumper) Wl(data... string) *Dumper {
	d.W(data...)
	d.Writer.Write([]byte("\n"))
	return d
}

func (d *Dumper) With(f func(d *Dumper)) *Dumper {
	d.Spaces++
	defer func() {
		d.Spaces--
	}()

	f(d)
	return d
}

type DumpLine struct {
	Dumper *Dumper
	Data []string
}

func (l *DumpLine) End(values... string) *Dumper {
	l.A(values...)
	l.Dumper.Wl(l.Data...)
	return l.Dumper
}

func (l *DumpLine) A(values... string) *DumpLine {
	l.Data = append(l.Data, values...)
	return l
}

func (d *Dumper) L() *DumpLine {
	return &DumpLine{d, []string{}}
}

