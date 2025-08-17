package web

import (
	"fmt"
	"net/http"
	"strings"
)

func (h *handler) index(w http.ResponseWriter, r *http.Request, p Params) {
	_, _ = w.Write([]byte("index: replace with template render"))
}

func (h *handler) groupIndex(w http.ResponseWriter, r *http.Request, p Params) {
	g := p["group_path"].([]string) // captured by @group
	_, _ = w.Write([]byte("group index for: /" + strings.Join(g, "/") + "/"))
}

func (h *handler) repoIndex(w http.ResponseWriter, r *http.Request, p Params) {
	repo := p["repo"].(string)
	g := p["group_path"].([]string)
	_, _ = w.Write([]byte(fmt.Sprintf("repo index: group=%q repo=%q", "/"+strings.Join(g, "/")+"/", repo)))
}

func (h *handler) repoTree(w http.ResponseWriter, r *http.Request, p Params) {
	repo := p["repo"].(string)
	rest := p["rest"].(string) // may be ""
	if p["dir_mode"].(bool) && rest != "" && !strings.HasSuffix(rest, "/") {
		rest += "/"
	}
	_, _ = w.Write([]byte(fmt.Sprintf("tree: repo=%q path=%q", repo, rest)))
}

func (h *handler) repoRaw(w http.ResponseWriter, r *http.Request, p Params) {
	repo := p["repo"].(string)
	rest := p["rest"].(string)
	if p["dir_mode"].(bool) && rest != "" && !strings.HasSuffix(rest, "/") {
		rest += "/"
	}
	_, _ = w.Write([]byte(fmt.Sprintf("raw: repo=%q path=%q", repo, rest)))
}

func (h *handler) notImplemented(w http.ResponseWriter, _ *http.Request, _ Params) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
