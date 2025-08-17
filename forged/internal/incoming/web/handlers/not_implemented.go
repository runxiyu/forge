package handlers

import (
	"net/http"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type NotImplementedHTTP struct{}

func NewNotImplementedHTTP() *NotImplementedHTTP { return &NotImplementedHTTP{} }

func (h *NotImplementedHTTP) Handle(w http.ResponseWriter, _ *http.Request, _ wtypes.Vars) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
