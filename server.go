// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.lindenii.runxiyu.org/lindenii-common/cmap"
	goSSH "golang.org/x/crypto/ssh"
)

type Server struct {
	config Config

	// database serves as the primary database handle for this entire application.
	// Transactions or single reads may be used from it. A [pgxpool.Pool] is
	// necessary to safely use pgx concurrently; pgx.Conn, etc. are insufficient.
	database *pgxpool.Pool

	sourceHandler http.Handler
	staticHandler http.Handler

	ircSendBuffered   chan string
	ircSendDirectChan chan errorBack[string]

	// globalData is passed as "global" when rendering HTML templates.
	globalData map[string]any

	serverPubkeyString string
	serverPubkeyFP     string
	serverPubkey       goSSH.PublicKey

	// packPasses contains hook cookies mapped to their packPass.
	packPasses cmap.Map[string, packPass]
}

func (s *Server) Setup() {
	s.sourceHandler = http.StripPrefix(
		"/-/source/",
		http.FileServer(http.FS(embeddedSourceFS)),
	)
	staticFS, err := fs.Sub(embeddedResourcesFS, "static")
	if err != nil {
		panic(err)
	}
	s.staticHandler = http.StripPrefix("/-/static/", http.FileServer(http.FS(staticFS)))
	s.globalData = map[string]any{
		"server_public_key_string":      &s.serverPubkeyString,
		"server_public_key_fingerprint": &s.serverPubkeyFP,
		"forge_version":                 VERSION,
		// Some other ones are populated after config parsing
	}
}

func (s *Server) Run() {
	if err := s.deployHooks(); err != nil {
		slog.Error("deploying hooks", "error", err)
		os.Exit(1)
	}
	if err := loadTemplates(); err != nil {
		slog.Error("loading templates", "error", err)
		os.Exit(1)
	}
	if err := s.deployGit2D(); err != nil {
		slog.Error("deploying git2d", "error", err)
		os.Exit(1)
	}

	// Launch Git2D
	go func() {
		cmd := exec.Command(s.config.Git.DaemonPath, s.config.Git.Socket) //#nosec G204
		cmd.Stderr = log.Writer()
		cmd.Stdout = log.Writer()
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}()

	// UNIX socket listener for hooks
	{
		hooksListener, err := net.Listen("unix", s.config.Hooks.Socket)
		if errors.Is(err, syscall.EADDRINUSE) {
			slog.Warn("removing existing socket", "path", s.config.Hooks.Socket)
			if err = syscall.Unlink(s.config.Hooks.Socket); err != nil {
				slog.Error("removing existing socket", "path", s.config.Hooks.Socket, "error", err)
				os.Exit(1)
			}
			if hooksListener, err = net.Listen("unix", s.config.Hooks.Socket); err != nil {
				slog.Error("listening hooks", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening hooks", "error", err)
			os.Exit(1)
		}
		slog.Info("listening hooks on unix", "path", s.config.Hooks.Socket)
		go func() {
			if err = s.serveGitHooks(hooksListener); err != nil {
				slog.Error("serving hooks", "error", err)
				os.Exit(1)
			}
		}()
	}

	// UNIX socket listener for LMTP
	{
		lmtpListener, err := net.Listen("unix", s.config.LMTP.Socket)
		if errors.Is(err, syscall.EADDRINUSE) {
			slog.Warn("removing existing socket", "path", s.config.LMTP.Socket)
			if err = syscall.Unlink(s.config.LMTP.Socket); err != nil {
				slog.Error("removing existing socket", "path", s.config.LMTP.Socket, "error", err)
				os.Exit(1)
			}
			if lmtpListener, err = net.Listen("unix", s.config.LMTP.Socket); err != nil {
				slog.Error("listening LMTP", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening LMTP", "error", err)
			os.Exit(1)
		}
		slog.Info("listening LMTP on unix", "path", s.config.LMTP.Socket)
		go func() {
			if err = s.serveLMTP(lmtpListener); err != nil {
				slog.Error("serving LMTP", "error", err)
				os.Exit(1)
			}
		}()
	}

	// SSH listener
	{
		sshListener, err := net.Listen(s.config.SSH.Net, s.config.SSH.Addr)
		if errors.Is(err, syscall.EADDRINUSE) && s.config.SSH.Net == "unix" {
			slog.Warn("removing existing socket", "path", s.config.SSH.Addr)
			if err = syscall.Unlink(s.config.SSH.Addr); err != nil {
				slog.Error("removing existing socket", "path", s.config.SSH.Addr, "error", err)
				os.Exit(1)
			}
			if sshListener, err = net.Listen(s.config.SSH.Net, s.config.SSH.Addr); err != nil {
				slog.Error("listening SSH", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening SSH", "error", err)
			os.Exit(1)
		}
		slog.Info("listening SSH on", "net", s.config.SSH.Net, "addr", s.config.SSH.Addr)
		go func() {
			if err = s.serveSSH(sshListener); err != nil {
				slog.Error("serving SSH", "error", err)
				os.Exit(1)
			}
		}()
	}

	// HTTP listener
	{
		httpListener, err := net.Listen(s.config.HTTP.Net, s.config.HTTP.Addr)
		if errors.Is(err, syscall.EADDRINUSE) && s.config.HTTP.Net == "unix" {
			slog.Warn("removing existing socket", "path", s.config.HTTP.Addr)
			if err = syscall.Unlink(s.config.HTTP.Addr); err != nil {
				slog.Error("removing existing socket", "path", s.config.HTTP.Addr, "error", err)
				os.Exit(1)
			}
			if httpListener, err = net.Listen(s.config.HTTP.Net, s.config.HTTP.Addr); err != nil {
				slog.Error("listening HTTP", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening HTTP", "error", err)
			os.Exit(1)
		}
		server := http.Server{
			Handler:      s,
			ReadTimeout:  time.Duration(s.config.HTTP.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(s.config.HTTP.ReadTimeout) * time.Second,
			IdleTimeout:  time.Duration(s.config.HTTP.ReadTimeout) * time.Second,
		} //exhaustruct:ignore
		slog.Info("listening HTTP on", "net", s.config.HTTP.Net, "addr", s.config.HTTP.Addr)
		go func() {
			if err = server.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("serving HTTP", "error", err)
				os.Exit(1)
			}
		}()
	}

	// IRC bot
	go s.ircBotLoop()

	select {}
}
