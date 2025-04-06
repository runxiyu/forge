// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/web"
)

// httpHandleIndex provides the main index page which includes a list of groups
// and some global information such as SSH keys.
func (s *Server) httpHandleIndex(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var err error
	var groups []nameDesc

	groups, err = s.queryNameDesc(request.Context(), "SELECT name, COALESCE(description, '') FROM groups WHERE parent_group IS NULL")
	if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error querying groups: "+err.Error())
		return
	}
	params["groups"] = groups

	s.renderTemplate(writer, "index", params)
}
