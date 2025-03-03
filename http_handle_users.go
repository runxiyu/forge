// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func handle_users(w http.ResponseWriter, r *http.Request, params map[string]any) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
