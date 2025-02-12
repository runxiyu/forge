package main

import (
	"fmt"
	"net/http"
)

func handle_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	rows, err := database.Query(r.Context(), "SELECT name FROM groups")
	if err != nil {
		fmt.Fprintln(w, "Error querying groups: " + err.Error())
		return
	}
	defer rows.Close()

	groups := []string{}
	for rows.Next() {
		var groupName string
		if err := rows.Scan(&groupName); err != nil {
			fmt.Fprintln(w, "Error scanning group name: " + err.Error())
			return
		}
		groups = append(groups, groupName)
	}

	if err := rows.Err(); err != nil {
		fmt.Fprintln(w, "Error iterating over rows: " + err.Error())
		return
	}

	params["groups"] = groups

	err = templates.ExecuteTemplate(w, "index", params)
	if err != nil {
		fmt.Fprintln(w, "Error rendering template: " + err.Error())
		return
	}
}
