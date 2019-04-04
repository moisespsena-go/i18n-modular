// +build !bindata

package locales

import (
	"github.com/moisespsena-go/assetfs"
)

func init() {
    FileSystem.OnPathRegister(func(fs assetfs.Interface) {
    	println("@@")
		BackendRepository.LoadFileSystem(fs.(*assetfs.AssetFileSystem), "p")
	})
}
