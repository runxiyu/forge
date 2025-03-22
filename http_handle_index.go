// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"runtime"

	"github.com/dustin/go-humanize"
)

func httpHandleIndex(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var err error
	var groups []nameDesc

	groups, err = queryNameDesc(request.Context(), "SELECT name, COALESCE(description, '') FROM groups WHERE parent_group IS NULL")
	if err != nil {
		errorPage500(writer, params, "Error querying groups: "+err.Error())
		return
	}
	params["groups"] = groups

	// Memory currently allocated
	memstats := runtime.MemStats{} //exhaustruct:ignore
	runtime.ReadMemStats(&memstats)
	params["mem"] = humanize.IBytes(memstats.Alloc)
	renderTemplate(writer, "index", params)
}
