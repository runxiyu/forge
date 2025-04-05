// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/forge/internal/web"
)

// idTitleStatus describes properties of a merge request that needs to be
// present in MR listings.
type idTitleStatus struct {
	ID     int
	Title  string
	Status string
}

// httpHandleRepoContribIndex provides an index to merge requests of a repo.
func (s *Server) httpHandleRepoContribIndex(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var rows pgx.Rows
	var result []idTitleStatus
	var err error

	if rows, err = s.database.Query(request.Context(),
		"SELECT repo_local_id, COALESCE(title, 'Untitled'), status FROM merge_requests WHERE repo_id = $1",
		params["repo_id"],
	); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error querying merge requests: "+err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var mrID int
		var mrTitle, mrStatus string
		if err = rows.Scan(&mrID, &mrTitle, &mrStatus); err != nil {
			web.ErrorPage500(s.templates, writer, params, "Error scanning merge request: "+err.Error())
			return
		}
		result = append(result, idTitleStatus{mrID, mrTitle, mrStatus})
	}
	if err = rows.Err(); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error ranging over merge requests: "+err.Error())
		return
	}
	params["merge_requests"] = result

	s.renderTemplate(writer, "repo_contrib_index", params)
}
