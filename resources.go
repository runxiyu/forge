package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed .gitignore LICENSE README.md
//go:embed *.go go.mod go.sum
//go:embed *.scfg
//go:embed Makefile
//go:embed schema.sql
//go:embed static/* templates/*
var source_fs embed.FS

var source_handler http.Handler

func init() {
	source_handler = http.StripPrefix(
		"/:/source/",
		http.FileServer(http.FS(source_fs)),
	)
}

//go:embed templates/* static/*
var resources_fs embed.FS

var templates *template.Template

func load_templates() (err error) {
	templates, err = template.New("templates").Funcs(template.FuncMap{
		"first_line": first_line,
		"base_name":  base_name,
	}).ParseFS(resources_fs, "templates/*")
	return err
}

var static_handler http.Handler

func init() {
	static_fs, err := fs.Sub(resources_fs, "static")
	if err != nil {
		panic(err)
	}
	static_handler = http.StripPrefix("/:/static/", http.FileServer(http.FS(static_fs)))
}
