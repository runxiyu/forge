// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"runtime"
)

func httpHandleGC(writer http.ResponseWriter, request *http.Request, _ map[string]any) {
	runtime.GC()
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}
