package web

import (
	"net/http"
	"path/filepath"

	handlers "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/handlers"
	repoHandlers "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/handlers/repo"
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

	// Feature handler instances
	indexHTTP := handlers.NewIndexHTTP()
	repoHTTP := repoHandlers.NewHTTP()

	// Index
	h.r.GET("/", indexHTTP.Index)

	// Top-level utilities
	h.r.ANY("-/login", h.notImplemented)
	h.r.ANY("-/users", h.notImplemented)

	// Group index (kept local for now; migrate later)
	h.r.GET("@group/", h.groupIndex)

	// Repo index (handled by repoHTTP)
	h.r.GET("@group/-/repos/:repo/", repoHTTP.Index)

	// Repo (kept local for now)
	h.r.ANY("@group/-/repos/:repo/info", h.notImplemented)
	h.r.ANY("@group/-/repos/:repo/git-upload-pack", h.notImplemented)

	// Repo features (kept local for now)
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
