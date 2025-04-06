// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package web

import (
	"html/template"
	"net/http"
)

// ErrorPage404 renders a 404 Not Found error page using the "404" template.
func ErrorPage404(templates *template.Template, w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusNotFound)
	_ = templates.ExecuteTemplate(w, "404", params)
}

// ErrorPage400 renders a 400 Bad Request error page using the "400" template.
// The error message is passed via the "complete_error_msg" template param.
func ErrorPage400(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "400", params)
}

// ErrorPage400Colon renders a 400 Bad Request error page telling the user
// that we migrated from : to -.
func ErrorPage400Colon(templates *template.Template, w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusBadRequest)
	_ = templates.ExecuteTemplate(w, "400_colon", params)
}

// ErrorPage403 renders a 403 Forbidden error page using the "403" template.
// The error message is passed via the "complete_error_msg" template param.
func ErrorPage403(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusForbidden)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "403", params)
}

// ErrorPage451 renders a 451 Unavailable For Legal Reasons error page using the "451" template.
// The error message is passed via the "complete_error_msg" template param.
func ErrorPage451(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusUnavailableForLegalReasons)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "451", params)
}

// ErrorPage500 renders a 500 Internal Server Error page using the "500" template.
// The error message is passed via the "complete_error_msg" template param.
func ErrorPage500(templates *template.Template, w http.ResponseWriter, params map[string]any, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	params["complete_error_msg"] = msg
	_ = templates.ExecuteTemplate(w, "500", params)
}

// ErrorPage501 renders a 501 Not Implemented error page using the "501" template.
func ErrorPage501(templates *template.Template, w http.ResponseWriter, params map[string]any) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = templates.ExecuteTemplate(w, "501", params)
}
