// +build !init,!assetfs_bindataCompile,!assetfs_bindataClean

package repositories

import (
	"github.com/moisespsena-go/i18n-modular/cmd/app"
	"github.com/moisespsena-go/i18n-modular/repository"
)

func Main() {
	app.Init(repository.AssetFS, repository.Repository)
	repository.CallCallbacks()
	app.Main(repository.AssetFS, repository.I18nModBackend)
}
