// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func get_user_info_from_request(r *http.Request) (id int, username string, err error) {
	session_cookie, err := r.Cookie("session")
	if err != nil {
		return
	}

	err = database.QueryRow(r.Context(), "SELECT user_id, COALESCE(username, '') FROM users u JOIN sessions s ON u.id = s.user_id WHERE s.session_id = $1;", session_cookie.Value).Scan(&id, &username)

	return
}
