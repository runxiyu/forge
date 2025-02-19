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

func handle_login(w http.ResponseWriter, r *http.Request, params map[string]any) {
	if r.Method != "POST" {
		render_template(w, "login", params)
		return
	}

	var user_id int
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	var password_hash string
	err := database.QueryRow(r.Context(), "SELECT id, COALESCE(password, '') FROM users WHERE username = $1", username).Scan(&user_id, &password_hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			params["login_error"] = "Unknown username"
			render_template(w, "login", params)
			return
		}
		http.Error(w, "Error querying user information: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if password_hash == "" {
		params["login_error"] = "User has no password"
		render_template(w, "login", params)
		return
	}

	match, err := argon2id.ComparePasswordAndHash(password, password_hash)
	if err != nil {
		http.Error(w, "Error comparing password and hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !match {
		params["login_error"] = "Invalid password"
		render_template(w, "login", params)
		return
	}

	cookie_value, err := random_urlsafe_string(16)
	if err != nil {
		http.Error(w, "Error getting random string: "+err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	expiry := now.Add(time.Duration(config.HTTP.CookieExpiry) * time.Second)

	cookie := http.Cookie{
		Name:     "session",
		Value:    cookie_value,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   false, // TODO
		Expires:  expiry,
		Path:     "/",
		// TODO: Expire
	}

	http.SetCookie(w, &cookie)

	_, err = database.Exec(r.Context(), "INSERT INTO sessions (user_id, session_id) VALUES ($1, $2)", user_id, cookie_value)
	if err != nil {
		http.Error(w, "Error inserting session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// random_urlsafe_string generates a random string of the given entropic size
// using the URL-safe base64 encoding. The actual size of the string returned
// will be 4*sz.
func random_urlsafe_string(sz int) (string, error) {
	r := make([]byte, 3*sz)
	_, err := rand.Read(r)
	if err != nil {
		return "", fmt.Errorf("error generating random string: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(r), nil
}
