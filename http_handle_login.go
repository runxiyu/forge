package main

import (
	"errors"
	"fmt"
	"net/http"

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

}
