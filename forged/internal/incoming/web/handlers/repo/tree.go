package repo

import (
	"fmt"
	"net/http"
	"strings"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

func (h *HTTP) Tree(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repo := v["repo"]
	rest := v["rest"] // may be ""
	if base.DirMode && rest != "" && !strings.HasSuffix(rest, "/") {
		rest += "/"
	}
	_, _ = w.Write([]byte(fmt.Sprintf("tree: repo=%q path=%q", repo, rest)))
}
