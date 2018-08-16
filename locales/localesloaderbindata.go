// +build bindata

package locales

func init() {
	BackendRepository.LoadFileSystem(AssetFS, "p")
}
