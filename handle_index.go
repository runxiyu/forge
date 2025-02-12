package main

import (
	"net/http"
)

func handle_index(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)

	rows, err := database.Query(r.Context(), "SELECT name FROM groups")
	if err != nil {
		_, _ = w.Write([]byte("Error querying groups: " + err.Error()))
		return
	}
	defer rows.Close()

	groups := []string{}
	for rows.Next() {
		var groupName string
		if err := rows.Scan(&groupName); err != nil {
			_, _ = w.Write([]byte("Error scanning group name: " + err.Error()))
			return
		}
		groups = append(groups, groupName)
	}

	if err := rows.Err(); err != nil {
		_, _ = w.Write([]byte("Error iterating over rows: " + err.Error()))
		return
	}

	data["groups"] = groups

	err = templates.ExecuteTemplate(w, "index", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
