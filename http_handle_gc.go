// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"runtime"
)

func handle_gc(w http.ResponseWriter, r *http.Request, params map[string]any) {
	runtime.GC()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
