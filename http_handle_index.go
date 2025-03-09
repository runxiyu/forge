// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"runtime"
)

func handle_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var err error
	var groups []name_desc_t

	groups, err = query_name_desc_list(r.Context(), "SELECT name, COALESCE(description, '') FROM groups WHERE parent_group IS NULL")
	if err != nil {
		http.Error(w, "Error querying groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["groups"] = groups

	// Memory currently allocated
	memstats := runtime.MemStats{}
	runtime.ReadMemStats(&memstats)
	params["mem"] = memstats.Alloc
	render_template(w, "index", params)
}
