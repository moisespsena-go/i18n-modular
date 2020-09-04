package i18nmod

import (
	"errors"
	"fmt"
	"strings"

	template2 "github.com/moisespsena/template/text/template"
)

type TError interface {
	error
	Translater
}

type Errors []error

func (this Errors) Error() string {
	var s = make([]string, len(this))
	for i, err := range this {
		s[i] = err.Error()
	}
	return strings.Join(s, ": ")
}

func (this Errors) Translate(ctx Context) string {
	var s = make([]string, len(this))
	for i, err := range this {
		if t, ok := err.(Translater); ok {
			s[i] = t.Translate(ctx)
		} else {
			s[i] = err.Error()
		}
	}
	return strings.Join(s, ": ")
}

func (this Errors) Cause(err error) Errors {
	return append(this, err)
}

type Err string

func (this Err) Error() string {
	return strings.ReplaceAll(string(this), "_", " ")
}

func (this Err) Translate(ctx Context) string {
	return ctx.T(string(this)).Get()
}

func (this Err) Cause(err error) Errors {
	return (Errors{this}).Cause(err)
}

func ErrorCtx(ctx Context, err error) error {
	if err == nil {
		return nil
	}
	if t, ok := err.(Translater); ok {
		return errors.New(t.Translate(ctx))
	}
	return err
}

func ErrorCtxS(ctx Context, err error) string {
	if t, ok := err.(Translater); ok {
		return t.Translate(ctx)
	}
	return err.Error()
}

type ErrData struct {
	Group, Key, Message, MessageT string
	data                          interface{}
}

func (this ErrData) WithData(data interface{}) ErrData {
	this.data = data
	return this
}

func (this ErrData) Data() interface{} {
	return this.data
}

func (this ErrData) Translate(ctx Context) string {
	return ctx.T(this.Group + "." + this.Key).Default(this.Error()).Data(this.data).Get()
}

func (this ErrData) Error() (msg string) {
	if this.Message == "" {
		msg = strings.ReplaceAll(string(this.Key), "_", " ")
		if this.data != nil {
			msg += fmt.Sprintf(" `%v`", this.data)
		}
	}
	return this.Message
}

func (this ErrData) Cause(err error) (errs Errors) {
	return append(errs, this, err)
}

type ErrDataT struct {
	Group, Key, MessageT string
	TExe                 *template2.Executor
	data                 interface{}
}

func (this ErrDataT) WithData(data interface{}) ErrDataT {
	this.data = data
	return this
}

func (this ErrDataT) Data() interface{} {
	return this.data
}

func (this ErrDataT) Translate(ctx Context) string {
	return ctx.T(this.Group + "." + this.Key).Default(this.Error()).Data(this.data).Get()
}

func (this ErrDataT) Error() (msg string) {
	if this.TExe != nil {
		if msg, _ = this.TExe.ExecuteString(this.data); msg != "" {
			return
		}
	}
	if this.MessageT == "" {
		msg = strings.ReplaceAll(string(this.Key), "_", " ")
		if this.data != nil {
			msg += fmt.Sprintf(" `%v`", this.data)
		}
		return
	}
	if tmpl, err := template2.New(this.Group + "." + this.Key).Parse(this.MessageT); err != nil {
		panic(fmt.Errorf("create template of error %q failed: %s", this.Group+"."+this.Key, err))
	} else {
		this.TExe = tmpl.CreateExecutor()
	}
	return this.Error()
}

func (this ErrDataT) Cause(err error) (errs Errors) {
	return append(errs, this, err)
}
