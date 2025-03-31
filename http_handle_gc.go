// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"runtime"
)

// httpHandleGC handles an HTTP request by calling the garbage collector and
// redirecting the user back to the home page.
//
// TODO: This should probably be removed or hidden behind an administrator's
// control panel, in the future.
func httpHandleGC(writer http.ResponseWriter, request *http.Request, _ map[string]any) {
	runtime.GC()
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}
