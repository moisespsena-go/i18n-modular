package templates

func Locales(prefix string) string {
	d := `package {{.Package}}
import (
	"github.com/moisespsena-go/i18n-modular/i18nmod/backends/yaml"
	"github.com/moisespsena-go/i18n-modular/i18nmod/backends/yaml/repository"
)

var (
	I18nModBackend         = yaml.New()
	I18nModFSPlugin = &repository.AssetFSPlugin{Backend:I18nModBackend, Prefix:"` + prefix + `"}
)

func init() {
	AssetFS.`
	if prefix != "" {
		d += `NameSpace("` + prefix + `").`
	}
	d += `AssetFS.RegisterPlugin(I18nModFSPlugin)
}
`
	return d
}