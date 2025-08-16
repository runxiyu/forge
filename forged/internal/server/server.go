package server

import (
	"context"
	"fmt"
	"log"

	"go.lindenii.runxiyu.org/forge/forged/internal/config"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/hooki"
	"go.lindenii.runxiyu.org/forge/forged/internal/lmtp"
)

type Server struct {
	config config.Config

	database database.Database
	hookPool hooki.Pool
	lmtpPool lmtp.Pool
}

func New(ctx context.Context, configPath string) (server *Server, err error) {
	server = &Server{}

	server.config, err = config.Open(configPath)
	if err != nil {
		return server, fmt.Errorf("open config: %w", err)
	}

	// TODO: Should this belong here, or in Run()?
	server.database, err = database.Open(ctx, server.config.DB)
	if err != nil {
		return server, fmt.Errorf("open database: %w", err)
	}

	return server, nil
}

func (s *Server) Run() error {
	// TODO: Not running git2d because it should be run separately.
	// This needs to be documented somewhere, hence a TODO here for now.

	go func() {
		s.hookPool = hooki.New(s.config.Hooks)
		if err := s.hookPool.Run(); err != nil {
			log.Fatalf("run hook pool: %v", err)
		}
	}()

	go func() {
		s.lmtpPool = lmtp.New(s.config.LMTP)
		if err := s.lmtpPool.Run(); err != nil {
			log.Fatalf("run LMTP pool: %v", err)
		}
	}()

	return nil
}
