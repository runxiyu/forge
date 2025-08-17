package templates

import (
	"html/template"
	"net/http"
)

type Renderer interface {
	Render(w http.ResponseWriter, name string, data any) error
}

type tmplRenderer struct {
	t *template.Template
}

func New(t *template.Template) Renderer {
	return &tmplRenderer{t: t}
}

func (r *tmplRenderer) Render(w http.ResponseWriter, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return r.t.ExecuteTemplate(w, name, data)
}
