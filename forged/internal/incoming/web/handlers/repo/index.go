package repo

import (
	"net/http"
	"strings"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

func (h *HTTP) Index(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repo := v["repo"]
	_ = h.r.Render(w, "repo/index.html", struct {
		Group string
		Repo  string
	}{
		Group: "/" + strings.Join(base.GroupPath, "/") + "/",
		Repo:  repo,
	})
}
