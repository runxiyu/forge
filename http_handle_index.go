package main

import (
	"net/http"
)

func handle_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	rows, err := database.Query(r.Context(), "SELECT name FROM groups")
	if err != nil {
		http.Error(w, "Error querying groups: : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	groups := []string{}
	for rows.Next() {
		var groupName string
		if err := rows.Scan(&groupName); err != nil {
			http.Error(w, "Error scanning group name: : "+err.Error(), http.StatusInternalServerError)
			return
		}
		groups = append(groups, groupName)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows: : "+err.Error(), http.StatusInternalServerError)
		return
	}

	params["groups"] = groups

	err = templates.ExecuteTemplate(w, "index", params)
	if err != nil {
		http.Error(w, "Error rendering template: : "+err.Error(), http.StatusInternalServerError)
		return
	}
}
