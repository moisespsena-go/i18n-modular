// +build assetfs_bindataCompile,assetfs_bindataClean

package init_

import (
	"github.com/moisespsena-go/i18n-modular/repository"
)

func Main() {
	fs := assetfs.NewAssetFileSystem()
	repo := fs.NewRepository(filepath.Join(path_helpers.GetCalledDir(), "..", "..", "repository"))
	repo.RegisterPlugin(&yamlrepository.RepositoryPlugin{Prefix:"p"})
	app.Init(fs, repo)
	repository.CallCallbacks()
}
