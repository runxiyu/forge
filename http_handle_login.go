package main

import (
	"fmt"
	"net/http"
)

func handle_login(w http.ResponseWriter, r *http.Request, params map[string]any) {
	if r.Method != "POST" {
		err := templates.ExecuteTemplate(w, "login", params)
		if err != nil {
			fmt.Fprintln(w, "Error rendering template:", err.Error())
			return
		}
	}

	_ = r.PostFormValue("username")
	_ = r.PostFormValue("password")
}
