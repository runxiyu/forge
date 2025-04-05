// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package render

import (
	"html"
	"html/template"
)

// EscapeHTML just escapes a string and wraps it in [template.HTML].
func EscapeHTML(s string) template.HTML {
	return template.HTML(html.EscapeString(s)) //#nosec G203
}
