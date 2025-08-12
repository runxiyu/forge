package templates

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
)

func MustParseDir(dir string, funcs template.FuncMap) *template.Template {
	base := template.New("").Funcs(funcs)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = base.Parse(string(b))
		return err
	})
	if err != nil {
		panic(err)
	}
	return base
}
