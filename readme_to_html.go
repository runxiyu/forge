// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

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

var markdownConverter = goldmark.New(goldmark.WithExtensions(extension.GFM))

// escapeHTML just escapes a string and wraps it in [template.HTML].
func escapeHTML(s string) template.HTML {
	return template.HTML(html.EscapeString(s)) //#nosec G203
}

// renderReadmeAtTree looks for README files in the supplied Git tree and
// returns its filename and rendered (and sanitized) HTML.
func renderReadmeAtTree(tree *object.Tree) (string, template.HTML) {
	for _, name := range []string{"README", "README.md", "README.org"} {
		file, err := tree.File(name)
		if err != nil {
			continue
		}
		contents, err := file.Contents()
		if err != nil {
			return "Error fetching README", escapeHTML("Unable to fetch contents of " + name + ": " + err.Error())
		}
		return renderReadme(stringToBytes(contents), name)
	}
	return "", ""
}

// renderReadme renders and sanitizes README content from a byte slice and filename.
func renderReadme(data []byte, filename string) (string, template.HTML) {
	switch strings.ToLower(filename) {
	case "readme":
		return "README", template.HTML("<pre>" + html.EscapeString(bytesToString(data)) + "</pre>") //#nosec G203
	case "readme.md":
		var buf bytes.Buffer
		if err := markdownConverter.Convert(data, &buf); err != nil {
			return "Error fetching README", escapeHTML("Unable to render README: " + err.Error())
		}
		return "README.md", template.HTML(bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())) //#nosec G203
	case "readme.org":
		htmlStr, err := org.New().Parse(strings.NewReader(bytesToString(data)), filename).Write(org.NewHTMLWriter())
		if err != nil {
			return "Error fetching README", escapeHTML("Unable to render README: " + err.Error())
		}
		return "README.org", template.HTML(bluemonday.UGCPolicy().Sanitize(htmlStr)) //#nosec G203
	default:
		return filename, template.HTML("<pre>" + html.EscapeString(bytesToString(data)) + "</pre>") //#nosec G203
	}
}
