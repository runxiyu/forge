package handlers

import (
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type NotImplementedHTTP struct {
	r templates.Renderer
}

func NewNotImplementedHTTP(r templates.Renderer) *NotImplementedHTTP {
	return &NotImplementedHTTP{
		r: r,
	}
}

func (h *NotImplementedHTTP) Handle(w http.ResponseWriter, _ *http.Request, _ wtypes.Vars) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
