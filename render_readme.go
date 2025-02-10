package main

import (
	"bytes"
	"html/template"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func render_readme_at_tree(tree *object.Tree) any {
	readme_file, err := tree.File("README.md")
	if err != nil {
		return ""
	}
	readme_file_contents, err := readme_file.Contents()
	if err != nil {
		return "Unable to fetch contents of README: " + err.Error()
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)

	var readme_rendered_unsafe bytes.Buffer
	err = md.Convert([]byte(readme_file_contents), &readme_rendered_unsafe)
	if err != nil {
		return "Unable to render README: " + err.Error()
	}
	return template.HTML(bluemonday.UGCPolicy().SanitizeBytes(readme_rendered_unsafe.Bytes()))
}
