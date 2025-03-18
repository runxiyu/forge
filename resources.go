// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// We embed all source for easy AGPL compliance.
//
//go:embed .gitignore .gitattributes
//go:embed LICENSE README.md
//go:embed *.go go.mod go.sum
//go:embed *.scfg
//go:embed Makefile
//go:embed static/* templates/* scripts/* sql/*
//go:embed hookc/*.c
//go:embed vendor/*
var source_fs embed.FS

var source_handler = http.StripPrefix(
	"/:/source/",
	http.FileServer(http.FS(source_fs)),
)

//go:embed templates/* static/* hookc/hookc
var resourcesFS embed.FS

var templates *template.Template

func loadTemplates() (err error) {
	m := minify.New()
	m.Add("text/html", &html.Minifier{TemplateDelims: [2]string{"{{", "}}"}, KeepDefaultAttrVals: true})

	templates = template.New("templates").Funcs(template.FuncMap{
		"first_line":   first_line,
		"base_name":    base_name,
		"path_escape":  path_escape,
		"query_escape": query_escape,
	})

	err = fs.WalkDir(resourcesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, err := fs.ReadFile(resourcesFS, path)
			if err != nil {
				return err
			}

			minified, err := m.Bytes("text/html", content)
			if err != nil {
				return err
			}

			_, err = templates.Parse(string(minified))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

var static_handler http.Handler

func init() {
	static_fs, err := fs.Sub(resourcesFS, "static")
	if err != nil {
		panic(err)
	}
	static_handler = http.StripPrefix("/:/static/", http.FileServer(http.FS(static_fs)))
}
