package repo

import (
	"fmt"
	"net/http"
	"strings"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

func (h *HTTP) Raw(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repo := v["repo"]
	rest := v["rest"]
	if base.DirMode && rest != "" && !strings.HasSuffix(rest, "/") {
		rest += "/"
	}
	_, _ = w.Write([]byte(fmt.Sprintf("raw: repo=%q path=%q", repo, rest)))
}
