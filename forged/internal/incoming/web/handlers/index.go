package handlers

import (
	"net/http"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type IndexHTTP struct{}

func NewIndexHTTP() *IndexHTTP { return &IndexHTTP{} }

func (h *IndexHTTP) Index(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	_, _ = w.Write([]byte("index: replace with template render"))
}
