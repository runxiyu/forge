package handlers

import (
	"net/http"
	"strings"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type GroupHTTP struct{}

func NewGroupHTTP() *GroupHTTP { return &GroupHTTP{} }

func (h *GroupHTTP) Index(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	base := wtypes.Base(r)
	_, _ = w.Write([]byte("group index for: /" + strings.Join(base.GroupPath, "/") + "/"))
}
