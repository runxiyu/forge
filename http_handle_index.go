// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func handle_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	groups, err := query_name_desc_list(r.Context(), "SELECT name, COALESCE(description, '') FROM groups")
	if err != nil {
		http.Error(w, "Error querying groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["groups"] = groups
	render_template(w, "index", params)
}
