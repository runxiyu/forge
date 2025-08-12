package server

import (
	"context"
	"fmt"

	"go.lindenii.runxiyu.org/forge/forged/internal/config"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/global"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/hooks"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/lmtp"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	config config.Config

	database   database.Database
	hookServer *hooks.Server
	lmtpServer *lmtp.Server
	webServer  *web.Server
	sshServer  *ssh.Server

	global global.Global
}

func New(configPath string) (server *Server, err error) {
	server = &Server{} //exhaustruct:ignore

	server.config, err = config.Open(configPath)
	if err != nil {
		return server, fmt.Errorf("open config: %w", err)
	}

	queries := queries.New(&server.database)

	server.global.ForgeVersion = "unknown" // TODO
	server.global.ForgeTitle = server.config.General.Title
	server.global.Config = &server.config
	server.global.Queries = queries

	server.hookServer = hooks.New(&server.global)
	server.lmtpServer = lmtp.New(&server.global)
	server.webServer = web.New(&server.global)
	server.sshServer, err = ssh.New(&server.global)
	if err != nil {
		return server, fmt.Errorf("create SSH server: %w", err)
	}

	return server, nil
}

func (server *Server) Run(ctx context.Context) (err error) {
	// TODO: Not running git2d because it should be run separately.
	// This needs to be documented somewhere, hence a TODO here for now.

	g, gctx := errgroup.WithContext(ctx)

	server.database, err = database.Open(gctx, server.config.DB.Conn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer server.database.Close()

	// TODO: neater way to do this for transactions in querypool?
	server.global.DB = &server.database

	g.Go(func() error { return server.hookServer.Run(gctx) })
	g.Go(func() error { return server.lmtpServer.Run(gctx) })
	g.Go(func() error { return server.webServer.Run(gctx) })
	g.Go(func() error { return server.sshServer.Run(gctx) })

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	err = ctx.Err()
	if err != nil {
		return fmt.Errorf("context exceeded: %w", err)
	}

	return nil
}
