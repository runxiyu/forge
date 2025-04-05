// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"embed"
	"html/template"
	"io/fs"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"go.lindenii.runxiyu.org/forge/internal/misc"
)

//go:embed LICENSE source.tar.gz
var embeddedSourceFS embed.FS

//go:embed templates/* static/*
//go:embed hookc/hookc git2d/git2d
var embeddedResourcesFS embed.FS

var templates *template.Template //nolint:gochecknoglobals

// loadTemplates minifies and loads HTML templates.
func loadTemplates() (err error) {
	minifier := minify.New()
	minifierOptions := html.Minifier{
		TemplateDelims:      [2]string{"{{", "}}"},
		KeepDefaultAttrVals: true,
	} //exhaustruct:ignore
	minifier.Add("text/html", &minifierOptions)

	templates = template.New("templates").Funcs(template.FuncMap{
		"first_line":        misc.FirstLine,
		"path_escape":       misc.PathEscape,
		"query_escape":      misc.QueryEscape,
		"dereference_error": misc.DereferenceOrZero[error],
		"minus":             misc.Minus,
	})

	err = fs.WalkDir(embeddedResourcesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, err := fs.ReadFile(embeddedResourcesFS, path)
			if err != nil {
				return err
			}

			minified, err := minifier.Bytes("text/html", content)
			if err != nil {
				return err
			}

			_, err = templates.Parse(misc.BytesToString(minified))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
