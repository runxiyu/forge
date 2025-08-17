// internal/incoming/web/handler.go
package web

import (
	"net/http"
	"path/filepath"
)

type handler struct {
	r *Router
}

func NewHandler(cfg Config) http.Handler {
	h := &handler{r: NewRouter().ReverseProxy(cfg.ReverseProxy)}

	// Static files
	staticDir := filepath.Join(cfg.Root, "static")
	staticFS := http.FileServer(http.Dir(staticDir))
	h.r.ANYHTTP("-/static/*rest",
		http.StripPrefix("/-/static/", staticFS),
		WithDirIfEmpty("rest"),
	)

	// Index
	h.r.GET("/", h.index)

	// Top-level utilities
	h.r.ANY("-/login", h.notImplemented)
	h.r.ANY("-/users", h.notImplemented)

	// Group index
	h.r.GET("@group/", h.groupIndex)

	// Repo index
	h.r.GET("@group/-/repos/:repo/", h.repoIndex)

	// Repo
	h.r.ANY("@group/-/repos/:repo/info", h.notImplemented)
	h.r.ANY("@group/-/repos/:repo/git-upload-pack", h.notImplemented)

	// Repo features
	h.r.GET("@group/-/repos/:repo/branches/", h.notImplemented)
	h.r.GET("@group/-/repos/:repo/log/", h.notImplemented)
	h.r.GET("@group/-/repos/:repo/commit/:commit", h.notImplemented)
	h.r.GET("@group/-/repos/:repo/tree/*rest", h.repoTree, WithDirIfEmpty("rest"))
	h.r.GET("@group/-/repos/:repo/raw/*rest", h.repoRaw, WithDirIfEmpty("rest"))
	h.r.GET("@group/-/repos/:repo/contrib/", h.notImplemented)
	h.r.GET("@group/-/repos/:repo/contrib/:mr", h.notImplemented)

	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
