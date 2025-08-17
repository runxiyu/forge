package web

import (
    "fmt"
    "net/http"
    "strings"
)

type RepoHTTP struct{}

func (h *RepoHTTP) Index(w http.ResponseWriter, r *http.Request, v Vars) {
    repo := v["repo"]
    gp := Base(r).GroupPath
    _, _ = fmt.Fprintf(w, "repo index: group=%q repo=%q", "/"+strings.Join(gp, "/")+"/", repo)
}

