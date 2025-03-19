// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import "net/http"

// renderTemplate abstracts out the annoyances of reporting template rendering
// errors.
func renderTemplate(w http.ResponseWriter, templateName string, params map[string]any) {
	if err := templates.ExecuteTemplate(w, templateName, params); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}
