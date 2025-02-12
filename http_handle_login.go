package main

import (
	"net/http"
)

func handle_login(w http.ResponseWriter, r *http.Request, params map[string]any) {
	if r.Method != "POST" {
		err := templates.ExecuteTemplate(w, "login", params)
		if err != nil {
			_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
			return
		}
	}

	_ = r.PostFormValue("username")
	_ = r.PostFormValue("password")
}
