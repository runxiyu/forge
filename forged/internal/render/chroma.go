// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package render

import (
	"bytes"
	"html/template"

	chromaHTML "github.com/alecthomas/chroma/v2/formatters/html"
	chromaLexers "github.com/alecthomas/chroma/v2/lexers"
	chromaStyles "github.com/alecthomas/chroma/v2/styles"
)

func Highlight(filename, content string) template.HTML {
	lexer := chromaLexers.Match(filename)
	if lexer == nil {
		lexer = chromaLexers.Fallback
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return template.HTML("<pre>Error tokenizing file: " + err.Error() + "</pre>") //#nosec G203`
	}

	var buf bytes.Buffer
	style := chromaStyles.Get("autumn")
	formatter := chromaHTML.New(
		chromaHTML.WithClasses(true),
		chromaHTML.TabWidth(8),
	)

	if err := formatter.Format(&buf, style, iterator); err != nil {
		return template.HTML("<pre>Error formatting file: " + err.Error() + "</pre>") //#nosec G203
	}

	return template.HTML(buf.Bytes()) //#nosec G203
}
