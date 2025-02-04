package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/* static/*
var resources_fs embed.FS

var templates *template.Template

func load_templates() (err error) {
	templates, err = template.ParseFS(resources_fs, "templates/*")
	return err
}

func serve_static() (err error) {
	static_fs, err := fs.Sub(resources_fs, "static")
	if err != nil {
		return err
	}
	http.Handle("/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.FS(static_fs)),
		),
	)
	return nil
}
