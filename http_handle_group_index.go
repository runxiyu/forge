package main

import (
	"net/http"
)

func handle_group_repos(w http.ResponseWriter, r *http.Request, params map[string]any) {
	group_name := params["group_name"]

	names, err := query_list[string](r.Context(), "SELECT r.name FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1;", group_name)
	if err != nil {
		http.Error(w, "Error getting groups:: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["repos"] = names

	err = templates.ExecuteTemplate(w, "group_repos", params)
	if err != nil {
		http.Error(w, "Error rendering template:: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
