// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
)

func getUserFromRequest(r *http.Request) (id int, username string, err error) {
	var sessionCookie *http.Cookie

	if sessionCookie, err = r.Cookie("session"); err != nil {
		return
	}

	err = database.QueryRow(
		r.Context(),
		"SELECT user_id, COALESCE(username, '') FROM users u JOIN sessions s ON u.id = s.user_id WHERE s.session_id = $1;",
		sessionCookie.Value,
	).Scan(&id, &username)

	return
}
