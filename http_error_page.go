// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func errorPage404(w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(404)
	_ = templates.ExecuteTemplate(w, "404", params)
}
