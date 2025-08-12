// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package web provides web-facing components of the forge.
package web

import (
	"html/template"
	"io/fs"
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/config"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/embed"
	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

// Server handles HTTP requests for the forge.
type Server struct {
	config        config.Config
	database      database.Database
	sourceHandler http.Handler
	staticHandler http.Handler
	templates     *template.Template
	globalData    map[string]any
}

// New creates a new web server.
func New(cfg config.Config, db database.Database, pubkeyStr, pubkeyFP *string, version string) (*Server, error) {
	s := &Server{config: cfg, database: db}
	s.globalData = map[string]any{
		"server_public_key_string":      pubkeyStr,
		"server_public_key_fingerprint": pubkeyFP,
		"forge_version":                 version,
		"forge_title":                   cfg.General.Title,
	}

	s.sourceHandler = http.StripPrefix(
		"/-/source/",
		http.FileServer(http.FS(embed.Source)),
	)
	staticFS, err := fs.Sub(embed.Resources, "forged/static")
	if err != nil {
		return s, err
	}
	s.staticHandler = http.StripPrefix("/-/static/", http.FileServer(http.FS(staticFS)))

	if err = s.loadTemplates(); err != nil {
		return s, err
	}

	return s, nil
}
