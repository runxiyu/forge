// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package render

import (
	"bytes"
	"html"
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.lindenii.runxiyu.org/forge/misc"
)

var markdownConverter = goldmark.New(goldmark.WithExtensions(extension.GFM))

// renderReadme renders and sanitizes README content from a byte slice and filename.
func Readme(data []byte, filename string) (string, template.HTML) {
	switch strings.ToLower(filename) {
	case "readme":
		return "README", template.HTML("<pre>" + html.EscapeString(misc.BytesToString(data)) + "</pre>") //#nosec G203
	case "readme.md":
		var buf bytes.Buffer
		if err := markdownConverter.Convert(data, &buf); err != nil {
			return "Error fetching README", EscapeHTML("Unable to render README: " + err.Error())
		}
		return "README.md", template.HTML(bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())) //#nosec G203
	case "readme.org":
		htmlStr, err := org.New().Parse(strings.NewReader(misc.BytesToString(data)), filename).Write(org.NewHTMLWriter())
		if err != nil {
			return "Error fetching README", EscapeHTML("Unable to render README: " + err.Error())
		}
		return "README.org", template.HTML(bluemonday.UGCPolicy().Sanitize(htmlStr)) //#nosec G203
	default:
		return filename, template.HTML("<pre>" + html.EscapeString(misc.BytesToString(data)) + "</pre>") //#nosec G203
	}
}
