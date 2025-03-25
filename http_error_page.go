// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func errorPage404(w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusNotFound)
	_ = templates.ExecuteTemplate(w, "404", params)
}

func errorPage400(w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "400", params)
}

func errorPage403(w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusForbidden)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "403", params)
}

func errorPage451(w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusUnavailableForLegalReasons)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "451", params)
}

func errorPage500(w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "500", params)
}

func errorPage501(w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = templates.ExecuteTemplate(w, "501", params)
}
