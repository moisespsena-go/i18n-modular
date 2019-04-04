package locales
import (
	"github.com/moisespsena-go/i18n-modular/i18nmod/backends/yaml"
	"github.com/moisespsena-go/i18n-modular/i18nmod/backends/yaml/repository"
)

var (
	Backend           = yaml.New()
)

func init() {
	&repository.RepositoryPlugin{Prefix:"p"}
}