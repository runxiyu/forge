package main

import (
	"net/http"
)

func handle_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	rows, err := database.Query(r.Context(), "SELECT name, COALESCE(description, '') FROM groups")
	if err != nil {
		http.Error(w, "Error querying groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	groups := []name_desc_t{}
	for rows.Next() {
		var groupName, groupDescription string
		if err := rows.Scan(&groupName, &groupDescription); err != nil {
			http.Error(w, "Error scanning group: "+err.Error(), http.StatusInternalServerError)
			return
		}
		groups = append(groups, name_desc_t{groupName, groupDescription})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	params["groups"] = groups

	err = templates.ExecuteTemplate(w, "index", params)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
