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
		err := templates.ExecuteTemplate(w, "login", params)
		if err != nil {
			fmt.Fprintln(w, "Error rendering template:", err.Error())
		}
		return
	}

	var user_id int
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	var password_hash string
	err := database.QueryRow(r.Context(), "SELECT id, password FROM users WHERE username = $1", username).Scan(&user_id, &password_hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			params["login_error"] = "Unknown username"
			err := templates.ExecuteTemplate(w, "login", params)
			if err != nil {
				fmt.Fprintln(w, "Error rendering template:", err.Error())
			}
			return
		}
		fmt.Fprintln(w, "Error querying user information:", err.Error())
		return
	}

	match, err := argon2id.ComparePasswordAndHash(password, password_hash)
	if err != nil {
		fmt.Fprintln(w, "Error comparing password and hash:", err.Error())
		return
	}

	if !match {
		params["login_error"] = "Invalid password"
		err := templates.ExecuteTemplate(w, "login", params)
		if err != nil {
			fmt.Fprintln(w, "Error rendering template:", err.Error())
			return
		}
		return
	}

	cookie_value, err := random_urlsafe_string(16)
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
		fmt.Fprintln(w, "Error inserting session:", err.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func random_urlsafe_string(sz int) (string, error) {
	r := make([]byte, 3*sz)
	_, err := rand.Read(r)
	if err != nil {
		return "", fmt.Errorf("error generating random string: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(r), nil
}
