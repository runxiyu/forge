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

        // Handlers
        indexHTTP := &IndexHTTP{}
        groupHTTP := &GroupHTTP{}
        repoHTTP := &RepoHTTP{}

        notImpl := func(w http.ResponseWriter, _ *http.Request, _ Vars) {
                http.Error(w, "not implemented", http.StatusNotImplemented)
        }

        // Index
        h.r.GET("/", indexHTTP.Index)

        // Top-level utilities
        h.r.ANY("-/login", notImpl)
        h.r.ANY("-/users", notImpl)

        // Group index
        h.r.GET("@group/", groupHTTP.Index)

        // Repo index
        h.r.GET("@group/-/repos/:repo/", repoHTTP.Index)

        // Repo
        h.r.ANY("@group/-/repos/:repo/info", notImpl)
        h.r.ANY("@group/-/repos/:repo/git-upload-pack", notImpl)

        // Repo features
        h.r.GET("@group/-/repos/:repo/branches/", notImpl)
        h.r.GET("@group/-/repos/:repo/log/", notImpl)
        h.r.GET("@group/-/repos/:repo/commit/:commit", notImpl)
        h.r.GET("@group/-/repos/:repo/tree/*rest", repoHTTP.Tree, WithDirIfEmpty("rest"))
        h.r.GET("@group/-/repos/:repo/raw/*rest", repoHTTP.Raw, WithDirIfEmpty("rest"))
        h.r.GET("@group/-/repos/:repo/contrib/", notImpl)
        h.r.GET("@group/-/repos/:repo/contrib/:mr", notImpl)

        return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
