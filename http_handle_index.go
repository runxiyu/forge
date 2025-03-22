// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"runtime"

	"github.com/dustin/go-humanize"
)

func httpHandleIndex(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var err error
	var groups []nameDesc

	groups, err = queryNameDesc(r.Context(), "SELECT name, COALESCE(description, '') FROM groups WHERE parent_group IS NULL")
	if err != nil {
		http.Error(w, "Error querying groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["groups"] = groups

	// Memory currently allocated
	memstats := runtime.MemStats{} //exhaustruct:ignore
	runtime.ReadMemStats(&memstats)
	params["mem"] = humanize.IBytes(memstats.Alloc)
	renderTemplate(w, "index", params)
}
