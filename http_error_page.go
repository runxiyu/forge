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

func errorPage400(w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(400)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "400", params)
}

func errorPage451(w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(451)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "451", params)
}
