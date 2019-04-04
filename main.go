// +build !init

package main

import (
	"fmt"
	"github.com/moisespsena-go/i18n-modular/i18nmod/backends/yaml"
	"github.com/theplant/cldr"

	"github.com/moisespsena-go/i18n-modular/i18nmod"
)

func main() {
	cldr.RegisterLocale(&cldr.Locale{Locale: "pt-br"})
	tr := i18nmod.NewTranslator()
	backend := yaml.New()
	backend.AddInput("g1", "_", func() (bytes []byte, e error) {
		return []byte(`a:
  name*:
    p: plural
    s: singular
b&: a`), nil
	})
	tr.AddBackend(backend)

	err := tr.Preload([]string{})
	if err != nil {
		panic(err)
	}

	ctx := tr.NewContext("pt-BR")
	fmt.Println(ctx.T("g1.a.name~p").Get())
}
