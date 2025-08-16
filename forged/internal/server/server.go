package server

import (
	"context"
	"fmt"

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

func New(configPath string) (server *Server, err error) {
	server = &Server{}

	server.config, err = config.Open(configPath)
	if err != nil {
		return server, fmt.Errorf("open config: %w", err)
	}

	server.hookServer = hooks.New(server.config.Hooks)
	server.lmtpServer = lmtp.New(server.config.LMTP)
	server.webServer = web.New(server.config.Web)
	server.sshServer, err = ssh.New(server.config.SSH)
	if err != nil {
		return server, fmt.Errorf("create SSH server: %w", err)
	}

	return server, nil
}

func (server *Server) Run(ctx context.Context) (err error) {
	// TODO: Not running git2d because it should be run separately.
	// This needs to be documented somewhere, hence a TODO here for now.

	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	server.database, err = database.Open(subCtx, server.config.DB)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	errCh := make(chan error)

	go func() {
		if err := server.hookServer.Run(subCtx); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		if err := server.lmtpServer.Run(subCtx); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		if err := server.webServer.Run(subCtx); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	go func() {
		if err := server.sshServer.Run(subCtx); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
	}

	return nil
}
