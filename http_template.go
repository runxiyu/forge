// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"

	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

// renderTemplate abstracts out the annoyances of reporting template rendering
// errors.
func renderTemplate(w http.ResponseWriter, templateName string, params map[string]any) {
	if err := templates.ExecuteTemplate(w, templateName, params); err != nil {
		http.Error(w, "error rendering template: "+err.Error(), http.StatusInternalServerError)
		clog.Error(err.Error())
	}
}
