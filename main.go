// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"flag"
	"net"
	"net/http"
	"syscall"
	"time"

	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

func main() {
	configPath := flag.String(
		"config",
		"/etc/lindenii/forge.scfg",
		"path to configuration file",
	)
	flag.Parse()

	if err := loadConfig(*configPath); err != nil {
		clog.Fatal(1, "Loading configuration: "+err.Error())
	}
	if err := deployHooks(); err != nil {
		clog.Fatal(1, "Deploying hooks to filesystem: "+err.Error())
	}
	if err := loadTemplates(); err != nil {
		clog.Fatal(1, "Loading templates: "+err.Error())
	}

	// UNIX socket listener for hooks
	var hooksListener net.Listener
	var err error
	hooksListener, err = net.Listen("unix", config.Hooks.Socket)
	if errors.Is(err, syscall.EADDRINUSE) {
		clog.Warn("Removing existing socket " + config.Hooks.Socket)
		if err = syscall.Unlink(config.Hooks.Socket); err != nil {
			clog.Fatal(1, "Removing existing socket: "+err.Error())
		}
		if hooksListener, err = net.Listen("unix", config.Hooks.Socket); err != nil {
			clog.Fatal(1, "Listening hooks: "+err.Error())
		}
	} else if err != nil {
		clog.Fatal(1, "Listening hooks: "+err.Error())
	}
	clog.Info("Listening hooks on unix " + config.Hooks.Socket)
	go func() {
		if err = serveGitHooks(hooksListener); err != nil {
			clog.Fatal(1, "Serving hooks: "+err.Error())
		}
	}()

	// SSH listener
	sshListener, err := net.Listen(config.SSH.Net, config.SSH.Addr)
	if errors.Is(err, syscall.EADDRINUSE) && config.SSH.Net == "unix" {
		clog.Warn("Removing existing socket " + config.SSH.Addr)
		if err = syscall.Unlink(config.SSH.Addr); err != nil {
			clog.Fatal(1, "Removing existing socket: "+err.Error())
		}
		if sshListener, err = net.Listen(config.SSH.Net, config.SSH.Addr); err != nil {
			clog.Fatal(1, "Listening SSH: "+err.Error())
		}
	} else if err != nil {
		clog.Fatal(1, "Listening SSH: "+err.Error())
	}
	clog.Info("Listening SSH on " + config.SSH.Net + " " + config.SSH.Addr)
	go func() {
		if err = serveSSH(sshListener); err != nil {
			clog.Fatal(1, "Serving SSH: "+err.Error())
		}
	}()

	// HTTP listener
	httpListener, err := net.Listen(config.HTTP.Net, config.HTTP.Addr)
	if errors.Is(err, syscall.EADDRINUSE) && config.HTTP.Net == "unix" {
		clog.Warn("Removing existing socket " + config.HTTP.Addr)
		if err = syscall.Unlink(config.HTTP.Addr); err != nil {
			clog.Fatal(1, "Removing existing socket: "+err.Error())
		}
		if httpListener, err = net.Listen(config.HTTP.Net, config.HTTP.Addr); err != nil {
			clog.Fatal(1, "Listening HTTP: "+err.Error())
		}
	} else if err != nil {
		clog.Fatal(1, "Listening HTTP: "+err.Error())
	}
	server := http.Server{
		Handler:      &forgeHTTPRouter{},
		ReadTimeout:  time.Duration(config.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.HTTP.ReadTimeout) * time.Second,
		IdleTimeout:  time.Duration(config.HTTP.ReadTimeout) * time.Second,
	} //exhaustruct:ignore
	clog.Info("Listening HTTP on " + config.HTTP.Net + " " + config.HTTP.Addr)
	go func() {
		if err = server.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			clog.Fatal(1, "Serving HTTP: "+err.Error())
		}
	}()

	// IRC bot
	go ircBotLoop()

	select {}
}
