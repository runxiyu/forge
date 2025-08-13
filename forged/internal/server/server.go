package server

import (
	"context"
	"fmt"
	"html/template"

	"go.lindenii.runxiyu.org/forge/forged/internal/config"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/hooki"
	"go.lindenii.runxiyu.org/forge/forged/internal/store"
)

type Server struct {
	config config.Config

	database database.Database
	stores *store.Set
	hookis *hooki.Pool
	templates *template.Template
}

func New(ctx context.Context, config config.Config) (*Server, error) {
	database, err := database.Open(ctx, config.DB)	
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return &Server{
		database: database,
	}, nil
}
