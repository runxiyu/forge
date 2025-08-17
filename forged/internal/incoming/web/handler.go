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
	groupHTTP := handlers.NewGroupHTTP()
	repoHTTP := repoHandlers.NewHTTP()
	notImpl := handlers.NewNotImplementedHTTP()

	// Index
	h.r.GET("/", indexHTTP.Index)

	// Top-level utilities
	h.r.ANY("-/login", notImpl.Handle)
	h.r.ANY("-/users", notImpl.Handle)

	// Group index
	h.r.GET("@group/", groupHTTP.Index)

	// Repo index
	h.r.GET("@group/-/repos/:repo/", repoHTTP.Index)

	// Repo (not implemented yet)
	h.r.ANY("@group/-/repos/:repo/info", notImpl.Handle)
	h.r.ANY("@group/-/repos/:repo/git-upload-pack", notImpl.Handle)

	// Repo features
	h.r.GET("@group/-/repos/:repo/branches/", notImpl.Handle)
	h.r.GET("@group/-/repos/:repo/log/", notImpl.Handle)
	h.r.GET("@group/-/repos/:repo/commit/:commit", notImpl.Handle)
	h.r.GET("@group/-/repos/:repo/tree/*rest", repoHTTP.Tree, WithDirIfEmpty("rest"))
	h.r.GET("@group/-/repos/:repo/raw/*rest", repoHTTP.Raw, WithDirIfEmpty("rest"))
	h.r.GET("@group/-/repos/:repo/contrib/", notImpl.Handle)
	h.r.GET("@group/-/repos/:repo/contrib/:mr", notImpl.Handle)

	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
