// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import "net/http"

// render_template abstracts out the annoyances of reporting template rendering
// errors.
func render_template(w http.ResponseWriter, template_name string, params map[string]any) {
	if err := templates.ExecuteTemplate(w, template_name, params); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}
