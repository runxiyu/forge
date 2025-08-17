package repo

import (
	"net/http"
	"strings"

	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type HTTP struct {
	r templates.Renderer
}

func NewHTTP(r templates.Renderer) *HTTP { return &HTTP{r: r} }

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

