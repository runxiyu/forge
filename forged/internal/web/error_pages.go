// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package web

import (
	"html/template"
	"net/http"
)

func ErrorPage404(templates *template.Template, w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusNotFound)
	_ = templates.ExecuteTemplate(w, "404", params)
}

func ErrorPage400(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "400", params)
}

func ErrorPage400Colon(templates *template.Template, w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusBadRequest)
	_ = templates.ExecuteTemplate(w, "400_colon", params)
}

func ErrorPage403(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusForbidden)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "403", params)
}

func ErrorPage451(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusUnavailableForLegalReasons)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "451", params)
}

func ErrorPage500(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "500", params)
}

func ErrorPage501(templates *template.Template, w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = templates.ExecuteTemplate(w, "501", params)
}
