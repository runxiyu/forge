// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5"
)

func httpHandleLogin(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var username, password string
	var userID int
	var passwordHash string
	var err error
	var passwordMatches bool
	var cookieValue string
	var now time.Time
	var expiry time.Time
	var cookie http.Cookie

	if request.Method != http.MethodPost {
		renderTemplate(writer, "login", params)
		return
	}

	username = request.PostFormValue("username")
	password = request.PostFormValue("password")

	err = database.QueryRow(request.Context(),
		"SELECT id, COALESCE(password, '') FROM users WHERE username = $1",
		username,
	).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			params["login_error"] = "Unknown username"
			renderTemplate(writer, "login", params)
			return
		}
		http.Error(writer, "Error querying user information: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if passwordHash == "" {
		params["login_error"] = "User has no password"
		renderTemplate(writer, "login", params)
		return
	}

	if passwordMatches, err = argon2id.ComparePasswordAndHash(password, passwordHash); err != nil {
		http.Error(writer, "Error comparing password and hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !passwordMatches {
		params["login_error"] = "Invalid password"
		renderTemplate(writer, "login", params)
		return
	}

	if cookieValue, err = randomUrlsafeStr(16); err != nil {
		http.Error(writer, "Error getting random string: "+err.Error(), http.StatusInternalServerError)
		return
	}

	now = time.Now()
	expiry = now.Add(time.Duration(config.HTTP.CookieExpiry) * time.Second)

	cookie = http.Cookie{
		Name:     "session",
		Value:    cookieValue,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   false, // TODO
		Expires:  expiry,
		Path:     "/",
	} //exhaustruct:ignore

	http.SetCookie(writer, &cookie)

	_, err = database.Exec(request.Context(), "INSERT INTO sessions (user_id, session_id) VALUES ($1, $2)", userID, cookieValue)
	if err != nil {
		http.Error(writer, "Error inserting session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

// randomUrlsafeStr generates a random string of the given entropic size
// using the URL-safe base64 encoding. The actual size of the string returned
// will be 4*sz.
func randomUrlsafeStr(sz int) (string, error) {
	r := make([]byte, 3*sz)
	_, err := rand.Read(r)
	if err != nil {
		return "", fmt.Errorf("error generating random string: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(r), nil
}
