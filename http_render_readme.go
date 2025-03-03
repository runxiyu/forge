// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	"html"
	"html/template"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var markdown_converter = goldmark.New(goldmark.WithExtensions(extension.GFM))

func render_readme_at_tree(tree *object.Tree) (readme_filename string, readme_content template.HTML) {
	var readme_rendered_unsafe bytes.Buffer

	readme_file, err := tree.File("README")
	if err == nil {
		readme_file_contents, err := readme_file.Contents()
		if err != nil {
			return "Error fetching README", string_escape_html("Unable to fetch contents of README: " + err.Error())
		}
		return "README", template.HTML("<pre>" + html.EscapeString(readme_file_contents) + "</pre>")
	}

	readme_file, err = tree.File("README.md")
	if err == nil {
		readme_file_contents, err := readme_file.Contents()
		if err != nil {
			return "Error fetching README", string_escape_html("Unable to fetch contents of README: " + err.Error())
		}
		err = markdown_converter.Convert([]byte(readme_file_contents), &readme_rendered_unsafe)
		if err != nil {
			return "Error fetching README", string_escape_html("Unable to render README: " + err.Error())
		}
		return "README.md", template.HTML(bluemonday.UGCPolicy().SanitizeBytes(readme_rendered_unsafe.Bytes()))
	}

	readme_file, err = tree.File("README.org")
	if err == nil {
		readme_file_contents, err := readme_file.Contents()
		if err != nil {
			return "Error fetching README", string_escape_html("Unable to fetch contents of README: " + err.Error())
		}
		org_html, err := org.New().Parse(strings.NewReader(readme_file_contents), readme_filename).Write(org.NewHTMLWriter())
		if err != nil {
			return "Error fetching README", string_escape_html("Unable to render README: " + err.Error())
		}
		return "README.org", template.HTML(bluemonday.UGCPolicy().Sanitize(org_html))
	}

	return "", ""
}

func string_escape_html(s string) template.HTML {
	return template.HTML(html.EscapeString(s))
}
