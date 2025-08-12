package templates

import (
	"bytes"
	"html/template"
	"log/slog"
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
	var buf bytes.Buffer
	if err := r.t.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("template render failed", "name", name, "error", err)
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	n, err := w.Write(buf.Bytes())
	if err != nil {
		return err
	}
	slog.Info("template rendered", "name", name, "bytes", n)
	return nil
}
