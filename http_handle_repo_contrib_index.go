// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"

	"github.com/jackc/pgx/v5"
)

type id_title_status_t struct {
	ID     int
	Title  string
	Status string
}

func handle_repo_contrib_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var rows pgx.Rows
	var result []id_title_status_t
	var err error

	if rows, err = database.Query(r.Context(),
		"SELECT id, COALESCE(title, 'Untitled'), status FROM merge_requests WHERE repo_id = $1",
		params["repo_id"],
	); err != nil {
		http.Error(w, "Error querying merge requests: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title, status string
		if err = rows.Scan(&id, &title, &status); err != nil {
			http.Error(w, "Error scanning merge request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, id_title_status_t{id, title, status})
	}
	if err = rows.Err(); err != nil {
		http.Error(w, "Error ranging over merge requests: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["merge_requests"] = result

	render_template(w, "repo_contrib_index", params)
}
