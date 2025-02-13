package main

import (
	"fmt"
	"net/http"
)

func handle_users(w http.ResponseWriter, r *http.Request, params map[string]any) {
	fmt.Fprintln(w, "Not implemented")
}
