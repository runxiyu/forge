package repo

import (
	"fmt"
	"net/http"
	"strings"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type HTTP struct{}

func NewHTTP() *HTTP { return &HTTP{} }

func (h *HTTP) Index(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repo := v["repo"]
	_, _ = w.Write([]byte(fmt.Sprintf("repo index: group=%q repo=%q",
		"/"+strings.Join(base.GroupPath, "/")+"/", repo)))
}
