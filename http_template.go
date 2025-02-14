package main

import "net/http"

func render_template(w http.ResponseWriter, template_name string, params map[string]any) {
	err := templates.ExecuteTemplate(w, template_name, params)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}
