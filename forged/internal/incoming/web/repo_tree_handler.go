package web

import (
    "fmt"
    "net/http"
    "strings"
)

func (h *RepoHTTP) Tree(w http.ResponseWriter, r *http.Request, v Vars) {
    repo := v["repo"]
    rest := v["rest"]
    if Base(r).DirMode && rest != "" && !strings.HasSuffix(rest, "/") {
        rest += "/"
    }
    _, _ = fmt.Fprintf(w, "tree: repo=%q path=%q", repo, rest)
}

