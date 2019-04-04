// +build init

package init_

import (
	"path/filepath"
	"github.com/moisespsena-go/path-helpers"
	yamlrepository "github.com/moisespsena-go/i18n-modular/i18nmod/backends/yaml/repository"
	"github.com/moisespsena-go/i18n-modular/cmd/app"
	"github.com/moisespsena-go/assetfs"
)

func Main() {
	fs := assetfs.NewAssetFileSystem()
	repo := fs.NewRepository(filepath.Join(path_helpers.GetCalledDir(), "..", "..", "repository"))
	repo.RegisterPlugin(&yamlrepository.RepositoryPlugin{Prefix:"p"})
	app.Init(fs, repo)
	repo.Init()
}
