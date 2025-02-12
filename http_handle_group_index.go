package main

import (
	"fmt"
	"net/http"
)

func handle_group_repos(w http.ResponseWriter, r *http.Request, params map[string]any) {
	group_name := params["group_name"]

	var names []string
	rows, err := database.Query(r.Context(), "SELECT r.name FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1;", group_name)
	if err != nil {
		fmt.Fprintln(w, "Error getting groups:", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			fmt.Fprintln(w, "Error scanning row:", err.Error())
			return
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		fmt.Fprintln(w, "Error iterating over rows:", err.Error())
		return
	}

	params["repos"] = names

	err = templates.ExecuteTemplate(w, "group_repos", params)
	if err != nil {
		fmt.Fprintln(w, "Error rendering template:", err.Error())
		return
	}
}
