package app

import (

	"github.com/theplant/cldr"
	"fmt"
	"github.com/moisespsena/go-assetfs/repository"
	"github.com/moisespsena/go-i18n-modular/i18nmod"
	"github.com/moisespsena/go-assetfs"
)

func Init(fs assetfs.Interface, repo repository.Interface) {
	// ..
	fs.RegisterPath("test_data")
}

func Main(fs assetfs.Interface, backend i18nmod.Backend) {
	cldr.RegisterLocale(&cldr.Locale{Locale: "pt-br"})
	// ns := fs.NameSpace("g1")
	tr := i18nmod.NewTranslator()
	tr.AddBackend(backend)

	err := tr.Preload([]string{})
	if err != nil {
		panic(err)
	}

	ctx := tr.NewContext("pt-BR")
	fmt.Println(ctx.T("g1.name").Get())
}
