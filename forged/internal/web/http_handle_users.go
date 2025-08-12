// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package web

import (
	"net/http"
)

// httpHandleUsers is a useless stub.
func (s *Server) httpHandleUsers(writer http.ResponseWriter, _ *http.Request, params map[string]any) {
	ErrorPage501(s.templates, writer, params)
}
