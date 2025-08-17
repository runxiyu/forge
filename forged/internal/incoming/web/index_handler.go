package web

import (
    "fmt"
    "net/http"
)

type IndexHTTP struct{}

func (h *IndexHTTP) Index(w http.ResponseWriter, r *http.Request, _ Vars) {
    _ = Base(r)
    _, _ = fmt.Fprint(w, "index: replace with template render")
}

