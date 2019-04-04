// +build assetfs_bindataCompile

package repositories

import (
	"github.com/moisespsena-go/i18n-modular/cmd/app"
	"github.com/moisespsena-go/i18n-modular/repository"
)

func Main() {
	app.Init(repository.AssetFS, repository.Repository)
	repository.CallCallbacks()
}
