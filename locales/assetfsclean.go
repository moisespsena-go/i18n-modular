// +build bindata_clean

package locales

func init() {
	AddCallback(Repository.Clean)
}
