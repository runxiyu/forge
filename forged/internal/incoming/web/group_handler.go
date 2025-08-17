package web

import (
    "fmt"
    "net/http"
    "strings"
)

type GroupHTTP struct{}

func (h *GroupHTTP) Index(w http.ResponseWriter, r *http.Request, _ Vars) {
    gp := Base(r).GroupPath
    _, _ = fmt.Fprint(w, "group index for: /"+strings.Join(gp, "/")+"/")
}

