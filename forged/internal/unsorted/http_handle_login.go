// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/forge/forged/internal/argon2id"
	"go.lindenii.runxiyu.org/forge/forged/internal/web"
)

// httpHandleLogin provides the login page for local users.
func (s *Server) httpHandleLogin(writer http.ResponseWriter, request *http.Request, params map[string]any) {
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
		s.renderTemplate(writer, "login", params)
		return
	}

	username = request.PostFormValue("username")
	password = request.PostFormValue("password")

	err = s.database.QueryRow(request.Context(),
		"SELECT id, COALESCE(password, '') FROM users WHERE username = $1",
		username,
	).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			params["login_error"] = "Unknown username"
			s.renderTemplate(writer, "login", params)
			return
		}
		web.ErrorPage500(s.templates, writer, params, "Error querying user information: "+err.Error())
		return
	}
	if passwordHash == "" {
		params["login_error"] = "User has no password"
		s.renderTemplate(writer, "login", params)
		return
	}

	if passwordMatches, err = argon2id.ComparePasswordAndHash(password, passwordHash); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error comparing password and hash: "+err.Error())
		return
	}

	if !passwordMatches {
		params["login_error"] = "Invalid password"
		s.renderTemplate(writer, "login", params)
		return
	}

	if cookieValue, err = randomUrlsafeStr(16); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting random string: "+err.Error())
		return
	}

	now = time.Now()
	expiry = now.Add(time.Duration(s.config.HTTP.CookieExpiry) * time.Second)

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

	_, err = s.database.Exec(request.Context(), "INSERT INTO sessions (user_id, session_id) VALUES ($1, $2)", userID, cookieValue)
	if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error inserting session: "+err.Error())
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
