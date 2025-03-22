// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func httpHandleUsers(writer http.ResponseWriter, _ *http.Request, _ map[string]any) {
	http.Error(writer, "Not implemented", http.StatusNotImplemented)
}
