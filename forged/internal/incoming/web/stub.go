package web

import (
	"fmt"
	"net/http"
	"strings"

	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

func (h *handler) groupIndex(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	base := wtypes.Base(r)
	_, _ = w.Write([]byte("group index for: /" + strings.Join(base.GroupPath, "/") + "/"))
}

func (h *handler) repoTree(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repo := v["repo"]
	rest := v["rest"] // may be ""
	if base.DirMode && rest != "" && !strings.HasSuffix(rest, "/") {
		rest += "/"
	}
	_, _ = w.Write([]byte(fmt.Sprintf("tree: repo=%q path=%q", repo, rest)))
}

func (h *handler) repoRaw(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repo := v["repo"]
	rest := v["rest"]
	if base.DirMode && rest != "" && !strings.HasSuffix(rest, "/") {
		rest += "/"
	}
	_, _ = w.Write([]byte(fmt.Sprintf("raw: repo=%q path=%q", repo, rest)))
}

func (h *handler) notImplemented(w http.ResponseWriter, _ *http.Request, _ wtypes.Vars) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
