package server

import (
	"context"
	"fmt"
	"log"

	"go.lindenii.runxiyu.org/forge/forged/internal/config"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/hooks"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/lmtp"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web"
)

type Server struct {
	config config.Config

	database   database.Database
	hookServer *hooks.Server
	lmtpServer *lmtp.Server
	webServer  *web.Server
	sshServer  *ssh.Server

	globalData struct {
		SSHPubkey      string
		SSHFingerprint string
		Version        string
	}
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

	server.hookServer = hooks.New(server.config.Hooks)

	server.lmtpServer = lmtp.New(server.config.LMTP)

	// TODO: Add HTTP and SSH servers

	return server, nil
}

func (s *Server) Run() error {
	// TODO: Not running git2d because it should be run separately.
	// This needs to be documented somewhere, hence a TODO here for now.

	go func() {
		if err := s.hookServer.Run(); err != nil {
			log.Fatalf("run hook pool: %v", err)
		}
	}()

	go func() {
		if err := s.lmtpServer.Run(); err != nil {
			log.Fatalf("run LMTP pool: %v", err)
		}
	}()

	return nil
}
