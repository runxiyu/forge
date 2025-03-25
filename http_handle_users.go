// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func httpHandleUsers(writer http.ResponseWriter, _ *http.Request, params map[string]any) {
	errorPage501(writer, params)
}
