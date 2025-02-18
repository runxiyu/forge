package main

import (
	"net/http"
)

func handle_repo_contrib_num(w http.ResponseWriter, r *http.Request, params map[string]any) {
	render_template(w, "repo_contrib_num", params)
}
