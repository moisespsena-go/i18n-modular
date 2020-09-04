package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/logging"
	path_helpers "github.com/moisespsena-go/path-helpers"

	"github.com/moisespsena-go/i18n-modular/i18nmod"
	"github.com/moisespsena-go/i18n-modular/i18nmod/backends/yaml"
)

var log = logging.GetOrCreateLogger(path_helpers.GetCalledDir())

type AssetFSPlugin struct {
	Backend *yaml.Backend
	Prefix  string
}

func (p *AssetFSPlugin) Init(fs assetfsapi.Interface) {
	p.LoadFileSystem(fs)
}

var Pattern = assetfs.NewGlobPattern(">\f{*.yml,*.yaml}")

func (p *AssetFSPlugin) load(basePth string, info assetfsapi.FileInfo) (err error) {
	pth := filepath.Join(basePth, info.Path())
	group := i18nmod.FormatGroupName(strings.Replace(filepath.Dir(pth), string(os.PathSeparator), ":", -1))
	log.Debug("+", group, "-->", info)
	return p.Backend.AddFileToGroup(group, func() ([]byte, error) {
		return assetfs.Data(info)
	}, info.Path())
}
func (p *AssetFSPlugin) LoadFileSystem(fs assetfsapi.Interface) error {
	basePth := fs.GetPath()
	if basePth == p.Prefix {
		basePth = ""
	} else {
		basePth = strings.TrimPrefix(basePth, p.Prefix+string(os.PathSeparator))
		log.Debug(basePth)
		if basePth != "" && basePth[0] == filepath.Separator {
			basePth = basePth[1:]
		}
	}
	return fs.NewGlob(Pattern).Info(func(info assetfsapi.FileInfo) error {
		return p.load(basePth, info)
	})
}

func (p *AssetFSPlugin) PathRegisterCallback(fs assetfsapi.Interface) {
	err := p.LoadFileSystem(fs)
	if err != nil {
		panic(fmt.Errorf("AssetFSPlugin.PathRegisterCallback failed: %v", err))
	}
}
