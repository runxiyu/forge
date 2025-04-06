// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"html/template"
	"io/fs"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"go.lindenii.runxiyu.org/forge/forged/internal/embed"
	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

// loadTemplates minifies and loads HTML templates.
func (s *Server) loadTemplates() (err error) {
	minifier := minify.New()
	minifierOptions := html.Minifier{
		TemplateDelims:      [2]string{"{{", "}}"},
		KeepDefaultAttrVals: true,
	} //exhaustruct:ignore
	minifier.Add("text/html", &minifierOptions)

	s.templates = template.New("templates").Funcs(template.FuncMap{
		"first_line":        misc.FirstLine,
		"path_escape":       misc.PathEscape,
		"query_escape":      misc.QueryEscape,
		"dereference_error": misc.DereferenceOrZero[error],
		"minus":             misc.Minus,
	})

	err = fs.WalkDir(embed.Resources, "forged/templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, err := fs.ReadFile(embed.Resources, path)
			if err != nil {
				return err
			}

			minified, err := minifier.Bytes("text/html", content)
			if err != nil {
				return err
			}

			_, err = s.templates.Parse(misc.BytesToString(minified))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
