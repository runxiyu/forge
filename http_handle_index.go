package main

import (
	"net/http"
)

func handle_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	groups, err := query_name_desc_list(r.Context(), "SELECT name, COALESCE(description, '') FROM groups")
	if err != nil {
		http.Error(w, "Error querying groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["groups"] = groups
	err = templates.ExecuteTemplate(w, "index", params)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
