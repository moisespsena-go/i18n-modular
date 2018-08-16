// +build pre_compile

package locales

func init() {
	AddCallback(Repository.Sync)
}
