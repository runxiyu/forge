package render

import (
	"html"
	"html/template"
)

// EscapeHTML just escapes a string and wraps it in [template.HTML].
func EscapeHTML(s string) template.HTML {
	return template.HTML(html.EscapeString(s)) //#nosec G203
}
