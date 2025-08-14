// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package web

import (
	"log/slog"
	"net/http"
)

// renderTemplate abstracts out the annoyances of reporting template rendering
// errors.
func (s *Server) renderTemplate(w http.ResponseWriter, templateName string, params map[string]any) {
	if err := s.templates.ExecuteTemplate(w, templateName, params); err != nil {
		http.Error(w, "error rendering template: "+err.Error(), http.StatusInternalServerError)
		slog.Error("error rendering template", "error", err.Error())
	}
}
