// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func handle_group_repos(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var group_name string
	var repos []name_desc_t
	var err error

	group_name = params["group_name"].(string)
	repos, err = query_name_desc_list(r.Context(), "SELECT r.name, COALESCE(r.description, '') FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1;", group_name)
	if err != nil {
		http.Error(w, "Error getting groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["repos"] = repos

	render_template(w, "group_repos", params)
}
