// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/web"
)

// httpHandleUsers is a useless stub.
func (s *Server) httpHandleUsers(writer http.ResponseWriter, _ *http.Request, params map[string]any) {
	web.ErrorPage501(s.templates, writer, params)
}
