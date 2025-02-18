package main

import (
	"net/http"

	"github.com/go-git/go-git/v5"
)

type id_title_status_t struct {
	ID     int
	Title  string
	Status string
}

func handle_repo_contrib_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	_ = params["repo"].(*git.Repository)

	rows, err := database.Query(r.Context(), "SELECT id, title, status FROM merge_requests WHERE repo_id = $1", params["repo_id"])
	if err != nil {
		http.Error(w, "Error querying merge requests: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := []id_title_status_t{}
	for rows.Next() {
		var id int
		var title, status string
		if err := rows.Scan(&id, &title, &status); err != nil {
			http.Error(w, "Error scanning merge request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, id_title_status_t{id, title, status})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error ranging over merge requests: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["merge_requests"] = result

	render_template(w, "repo_contrib_index", params)
}
