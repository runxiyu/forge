// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"net/http"
	"runtime"

	"github.com/dustin/go-humanize"
)

// httpHandleIndex provides the main index page which includes a list of groups
// and some global information such as SSH keys.
func (s *Server) httpHandleIndex(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var err error
	var groups []nameDesc

	groups, err = s.queryNameDesc(request.Context(), "SELECT name, COALESCE(description, '') FROM groups WHERE parent_group IS NULL")
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
