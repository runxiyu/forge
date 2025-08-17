package web

import (
	"html/template"
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/global"
	handlers "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/handlers"
	repoHandlers "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/handlers/repo"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
)

type handler struct {
	r *Router
}

func NewHandler(cfg Config, global *global.Global, queries *queries.Queries) *handler {
	h := &handler{r: NewRouter().ReverseProxy(cfg.ReverseProxy).Global(global).Queries(queries)}

	staticFS := http.FileServer(http.Dir(cfg.StaticPath))
	h.r.ANYHTTP("-/static/*rest",
		http.StripPrefix("/-/static/", staticFS),
		WithDirIfEmpty("rest"),
	)

	funcs := template.FuncMap{
		"path_escape":       misc.PathEscape,
		"query_escape":      misc.QueryEscape,
		"minus":             misc.Minus,
		"first_line":        misc.FirstLine,
		"dereference_error": misc.DereferenceOrZero[error],
	}
	t := templates.MustParseDir(cfg.TemplatesPath, funcs)
	renderer := templates.New(t)

	indexHTTP := handlers.NewIndexHTTP(renderer)
	groupHTTP := handlers.NewGroupHTTP(renderer)
	repoHTTP := repoHandlers.NewHTTP(renderer)
	notImpl := handlers.NewNotImplementedHTTP(renderer)

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
