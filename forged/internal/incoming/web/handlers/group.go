package handlers

import (
	"net/http"
	"strings"

	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type GroupHTTP struct {
	r templates.Renderer
}

func NewGroupHTTP(r templates.Renderer) *GroupHTTP {
	return &GroupHTTP{
		r: r,
	}
}

func (h *GroupHTTP) Index(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	base := wtypes.Base(r)
	_ = h.r.Render(w, "group/index.html", struct {
		GroupPath string
	}{
		GroupPath: "/" + strings.Join(base.GroupPath, "/") + "/",
	})
}
