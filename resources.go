// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

//go:embed LICENSE source.tar.gz
var sourceFS embed.FS

var sourceHandler = http.StripPrefix(
	"/-/source/",
	http.FileServer(http.FS(sourceFS)),
)

//go:embed templates/* static/* hookc/hookc man/*.html man/*.txt man/*.css
var resourcesFS embed.FS

var templates *template.Template

// loadTemplates minifies and loads HTML templates.
func loadTemplates() (err error) {
	minifier := minify.New()
	minifierOptions := html.Minifier{
		TemplateDelims:      [2]string{"{{", "}}"},
		KeepDefaultAttrVals: true,
	} //exhaustruct:ignore
	minifier.Add("text/html", &minifierOptions)

	templates = template.New("templates").Funcs(template.FuncMap{
		"first_line":        firstLine,
		"path_escape":       pathEscape,
		"query_escape":      queryEscape,
		"dereference_error": dereferenceOrZero[error],
		"minus":             minus,
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

			minified, err := minifier.Bytes("text/html", content)
			if err != nil {
				return err
			}

			_, err = templates.Parse(bytesToString(minified))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

var (
	staticHandler http.Handler
	manHandler    http.Handler
)

// This init sets up static and man handlers. The resulting handlers must be
// used in the HTTP router, and do nothing unless called from elsewhere.
func init() {
	staticFS, err := fs.Sub(resourcesFS, "static")
	if err != nil {
		panic(err)
	}
	staticHandler = http.StripPrefix("/-/static/", http.FileServer(http.FS(staticFS)))
	manFS, err := fs.Sub(resourcesFS, "man")
	if err != nil {
		panic(err)
	}
	manHandler = http.StripPrefix("/-/man/", http.FileServer(http.FS(manFS)))
}
