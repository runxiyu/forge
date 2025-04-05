// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"net/http"
)

// httpHandleUsers is a useless stub.
func httpHandleUsers(writer http.ResponseWriter, _ *http.Request, params map[string]any) {
	errorPage501(writer, params)
}
