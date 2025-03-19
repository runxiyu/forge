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

var markdownConverter = goldmark.New(goldmark.WithExtensions(extension.GFM))

func renderReadmeAtTree(tree *object.Tree) (readmeFilename string, readmeRenderedSafeHTML template.HTML) {
	var readmeRenderedUnsafe bytes.Buffer
	var readmeFile *object.File
	var readmeFileContents string
	var err error

	if readmeFile, err = tree.File("README"); err == nil {
		if readmeFileContents, err = readmeFile.Contents(); err != nil {
			return "Error fetching README", escapeHTML("Unable to fetch contents of README: " + err.Error())
		}

		return "README", template.HTML("<pre>" + html.EscapeString(readmeFileContents) + "</pre>") //#nosec G203
	}

	if readmeFile, err = tree.File("README.md"); err == nil {
		if readmeFileContents, err = readmeFile.Contents(); err != nil {
			return "Error fetching README", escapeHTML("Unable to fetch contents of README: " + err.Error())
		}

		if err = markdownConverter.Convert([]byte(readmeFileContents), &readmeRenderedUnsafe); err != nil {
			return "Error fetching README", escapeHTML("Unable to render README: " + err.Error())
		}

		return "README.md", template.HTML(bluemonday.UGCPolicy().SanitizeBytes(readmeRenderedUnsafe.Bytes())) //#nosec G203
	}

	if readmeFile, err = tree.File("README.org"); err == nil {
		if readmeFileContents, err = readmeFile.Contents(); err != nil {
			return "Error fetching README", escapeHTML("Unable to fetch contents of README: " + err.Error())
		}

		orgHTML, err := org.New().Parse(strings.NewReader(readmeFileContents), readmeFilename).Write(org.NewHTMLWriter())
		if err != nil {
			return "Error fetching README", escapeHTML("Unable to render README: " + err.Error())
		}

		return "README.org", template.HTML(bluemonday.UGCPolicy().Sanitize(orgHTML)) //#nosec G203
	}

	return "", ""
}

func escapeHTML(s string) template.HTML {
	return template.HTML(html.EscapeString(s)) //#nosec G203
}
