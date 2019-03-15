package i18nmod

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/moisespsena/go-path-helpers"
	"github.com/moisespsena/template/text/template"
	"github.com/pkg/errors"
)

func getExtension(fileName string) (extension string, err error) {
	ar := strings.SplitAfterN(fileName, ".", 2)
	if len(ar) == 2 {
		extension = "." + ar[1]
	} else {
		err = errors.New("No Extension found")
	}
	return
}

func WalkDir(name string, path string, cb func(key string, items []string) error) (err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("i18nmod.WalkDir: Failed to scan '", path, "': ", err)
	}

	var items []string

	for _, f := range files {
		fname := f.Name()
		p := filepath.Join(path, fname)

		if f.IsDir() {
			if name == "" {
				err = WalkDir(fname, p, cb)
			} else {
				err = WalkDir(name+":"+fname, p, cb)
			}
			if err != nil {
				return err
			}
		} else if !strings.HasPrefix(fname, ".") {
			ext, err := getExtension(fname)
			if err != nil {
				return fmt.Errorf("i18nmod.DirIterate: Failed to scan '", path, "': ", err)
			} else if ext == ".yaml" {
				items = append(items, p)
			}
		}
	}

	if len(items) > 0 {
		return cb(name, items)
	}

	return
}

func FormatGroupName(groupname string) string {
	return strings.Replace(groupname, ".", "_", -1)
}

func PkgToGroup(pkgPath string, sub ...string) string {
	p := []string{FormatGroupName(strings.Replace(strings.Replace(pkgPath, "\\", "/", -1),
		"/", ":", -1))}
	return strings.Join(append(p, sub...), ":")
}

func StructGroup(value interface{}) string {
	pkgPath := path_helpers.PkgPathOf(value)
	return PkgToGroup(pkgPath, ModelType(value).Name())
}

func ParseTemplate(text string) *template.Executor {
	t, _ := template.New("").Parse(text)
	return t.CreateExecutor()
}

// ModelType get value's model type
func ModelType(value interface{}) reflect.Type {
	reflectType := reflect.Indirect(reflect.ValueOf(value)).Type()

	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}

	return reflectType
}
