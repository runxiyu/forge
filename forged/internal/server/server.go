// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package server

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/cmap"
	"go.lindenii.runxiyu.org/forge/forged/internal/config"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/irc"
	"go.lindenii.runxiyu.org/forge/forged/internal/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/web"
)

type Server struct {
	config config.Config

	database database.Database

	serverPubkeyString string
	serverPubkeyFP     string

	// packPasses contains hook cookies mapped to their packPass.
	packPasses cmap.Map[string, ssh.PackPass]

	web *web.Server
	ssh *ssh.Server

	ircBot *irc.Bot

	ready bool
}

func NewServer(configPath string) (*Server, error) {
	s := &Server{} //exhaustruct:ignore
	s.packPasses = cmap.Map[string, ssh.PackPass]{}

	cfg, err := config.Load(configPath)
	if err != nil {
		return s, err
	}
	if cfg.DB.Type != "postgres" {
		return s, errors.New("unsupported database type")
	}
	s.config = cfg
	if s.database, err = database.Open(s.config.DB.Conn); err != nil {
		return s, err
	}
	s.web, err = web.New(s.config, s.database, &s.serverPubkeyString, &s.serverPubkeyFP, version)
	if err != nil {
		return s, err
	}
	s.ssh = ssh.New(s.config, s.database, &s.serverPubkeyString, &s.serverPubkeyFP, &s.packPasses, version)

	s.ready = true

	return s, nil
}

func (s *Server) Run() error {
	if !s.ready {
		return errors.New("not ready")
	}

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
			if err = s.ssh.Serve(sshListener); err != nil {
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
			Handler:      s.web,
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

	// Pprof listener
	{
		pprofListener, err := net.Listen(s.config.Pprof.Net, s.config.Pprof.Addr)
		if err != nil {
			slog.Error("listening pprof", "error", err)
			os.Exit(1)
		}

		slog.Info("listening pprof on", "net", s.config.Pprof.Net, "addr", s.config.Pprof.Addr)
		go func() {
			if err := http.Serve(pprofListener, nil); err != nil {
				slog.Error("serving pprof", "error", err)
				os.Exit(1)
			}
		}()
	}

	s.ircBot = irc.NewBot(&s.config.IRC)
	// IRC bot
	go s.ircBot.ConnectLoop()

	select {}
}
