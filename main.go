// +build !init

package main

import (
	"github.com/moisespsena/go-i18n-modular/setup"
)

func main() {
	cmd.Setup()
	/*
	cldr.RegisterLocale(&cldr.Locale{Locale: "pt-br"})
	backend := locales.Backend
	fs := locales.FS
	locales.LoadDir("test_data")
	fs.Init("locales")
	tr := i18nmod.NewTranslator()
	tr.AddBackend(backend)

	err := tr.Preload([]string{})
	if err != nil {
		panic(err)
	}

	ctx := tr.NewContext("pt-BR")
	fmt.Println(ctx.T("g1.name").Get())
	*/
}
