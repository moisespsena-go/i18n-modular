package i18nmod

import (
	"strings"
	"github.com/pkg/errors"
	"fmt"
	"reflect"
	"io/ioutil"
	"path/filepath"
	"github.com/moisespsena/template/text/template"
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

func Walk(name string, path string, cb func(key string, items []string) error) (err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("i18nmod.Walk: Failed to scan '", path, "': ", err)
	}

	var items []string

	for _, f := range files {
		fname := f.Name()
		p := filepath.Join(path, fname)

		if f.IsDir() {
			if name == "" {
				err = Walk(fname, p, cb)
			} else {
				err = Walk(name+":"+fname, p, cb)
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
	pkgPath := reflect.TypeOf(value).Elem().PkgPath()
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
