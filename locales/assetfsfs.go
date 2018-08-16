// +build !bindata

package locales

import (
	"github.com/moisespsena/go-assetfs"
)

var (
	FileSystem                   = assetfs.NewAssetFileSystem()
	AssetFS    assetfs.Interface = FileSystem
)

func init() {
	FileSystem.OnPathRegister(func(fs assetfs.Interface) {
		Repository.AddSource(fs.(*assetfs.AssetFileSystem).GetPath())
	})
}
