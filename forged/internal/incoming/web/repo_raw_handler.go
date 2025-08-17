package web

import (
    "fmt"
    "net/http"
    "strings"
)

func (h *RepoHTTP) Raw(w http.ResponseWriter, r *http.Request, v Vars) {
    repo := v["repo"]
    rest := v["rest"]
    if Base(r).DirMode && rest != "" && !strings.HasSuffix(rest, "/") {
        rest += "/"
    }
    _, _ = fmt.Fprintf(w, "raw: repo=%q path=%q", repo, rest)
}

