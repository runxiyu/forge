// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String(
		"config",
		"/etc/lindenii/forge.scfg",
		"path to configuration file",
	)
	flag.Parse()

	if err := loadConfig(*configPath); err != nil {
		slog.Error("loading configuration", "error", err)
		os.Exit(1)
	}
	if err := deployHooks(); err != nil {
		slog.Error("deploying hooks", "error", err)
		os.Exit(1)
	}
	if err := loadTemplates(); err != nil {
		slog.Error("loading templates", "error", err)
		os.Exit(1)
	}
	if err := deployGit2D(); err != nil {
		slog.Error("deploying git2d", "error", err)
		os.Exit(1)
	}

	// Launch Git2D
	go func() {
		cmd := exec.Command(config.Git.DaemonPath, config.Git.Socket) //#nosec G204
		cmd.Stderr = log.Writer()
		cmd.Stdout = log.Writer()
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}()

	// UNIX socket listener for hooks
	{
		hooksListener, err := net.Listen("unix", config.Hooks.Socket)
		if errors.Is(err, syscall.EADDRINUSE) {
			slog.Warn("removing existing socket", "path", config.Hooks.Socket)
			if err = syscall.Unlink(config.Hooks.Socket); err != nil {
				slog.Error("removing existing socket", "path", config.Hooks.Socket, "error", err)
				os.Exit(1)
			}
			if hooksListener, err = net.Listen("unix", config.Hooks.Socket); err != nil {
				slog.Error("listening hooks", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening hooks", "error", err)
			os.Exit(1)
		}
		slog.Info("listening hooks on unix", "path", config.Hooks.Socket)
		go func() {
			if err = serveGitHooks(hooksListener); err != nil {
				slog.Error("serving hooks", "error", err)
				os.Exit(1)
			}
		}()
	}

	// UNIX socket listener for LMTP
	{
		lmtpListener, err := net.Listen("unix", config.LMTP.Socket)
		if errors.Is(err, syscall.EADDRINUSE) {
			slog.Warn("removing existing socket", "path", config.LMTP.Socket)
			if err = syscall.Unlink(config.LMTP.Socket); err != nil {
				slog.Error("removing existing socket", "path", config.LMTP.Socket, "error", err)
				os.Exit(1)
			}
			if lmtpListener, err = net.Listen("unix", config.LMTP.Socket); err != nil {
				slog.Error("listening LMTP", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening LMTP", "error", err)
			os.Exit(1)
		}
		slog.Info("listening LMTP on unix", "path", config.LMTP.Socket)
		go func() {
			if err = serveLMTP(lmtpListener); err != nil {
				slog.Error("serving LMTP", "error", err)
				os.Exit(1)
			}
		}()
	}

	// SSH listener
	{
		sshListener, err := net.Listen(config.SSH.Net, config.SSH.Addr)
		if errors.Is(err, syscall.EADDRINUSE) && config.SSH.Net == "unix" {
			slog.Warn("removing existing socket", "path", config.SSH.Addr)
			if err = syscall.Unlink(config.SSH.Addr); err != nil {
				slog.Error("removing existing socket", "path", config.SSH.Addr, "error", err)
				os.Exit(1)
			}
			if sshListener, err = net.Listen(config.SSH.Net, config.SSH.Addr); err != nil {
				slog.Error("listening SSH", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening SSH", "error", err)
			os.Exit(1)
		}
		slog.Info("listening SSH on", "net", config.SSH.Net, "addr", config.SSH.Addr)
		go func() {
			if err = serveSSH(sshListener); err != nil {
				slog.Error("serving SSH", "error", err)
				os.Exit(1)
			}
		}()
	}

	// HTTP listener
	{
		httpListener, err := net.Listen(config.HTTP.Net, config.HTTP.Addr)
		if errors.Is(err, syscall.EADDRINUSE) && config.HTTP.Net == "unix" {
			slog.Warn("removing existing socket", "path", config.HTTP.Addr)
			if err = syscall.Unlink(config.HTTP.Addr); err != nil {
				slog.Error("removing existing socket", "path", config.HTTP.Addr, "error", err)
				os.Exit(1)
			}
			if httpListener, err = net.Listen(config.HTTP.Net, config.HTTP.Addr); err != nil {
				slog.Error("listening HTTP", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening HTTP", "error", err)
			os.Exit(1)
		}
		server := http.Server{
			Handler:      &forgeHTTPRouter{},
			ReadTimeout:  time.Duration(config.HTTP.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(config.HTTP.ReadTimeout) * time.Second,
			IdleTimeout:  time.Duration(config.HTTP.ReadTimeout) * time.Second,
		} //exhaustruct:ignore
		slog.Info("listening HTTP on", "net", config.HTTP.Net, "addr", config.HTTP.Addr)
		go func() {
			if err = server.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("serving HTTP", "error", err)
				os.Exit(1)
			}
		}()
	}

	// IRC bot
	go ircBotLoop()

	select {}
}
