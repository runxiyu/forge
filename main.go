package main

import (
	"errors"
	"flag"
	"net"
	"net/http"
	"syscall"

	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

func main() {
	config_path := flag.String(
		"config",
		"/etc/lindenii/forge.scfg",
		"path to configuration file",
	)
	flag.Parse()

	if err := load_config(*config_path); err != nil {
		clog.Fatal(1, "Loading configuration: "+err.Error())
	}
	if err := deploy_hooks_to_filesystem(); err != nil {
		clog.Fatal(1, "Deploying hooks to filesystem: "+err.Error())
	}
	if err := load_templates(); err != nil {
		clog.Fatal(1, "Loading templates: "+err.Error())
	}

	// UNIX socket listener for hooks
	var hooks_listener net.Listener
	var err error
	hooks_listener, err = net.Listen("unix", config.Hooks.Socket)
	if errors.Is(err, syscall.EADDRINUSE) {
		clog.Warn("Removing stale socket " + config.Hooks.Socket)
		if err := syscall.Unlink(config.Hooks.Socket); err != nil {
			clog.Fatal(1, "Removing stale socket: "+err.Error())
		}
		hooks_listener, err = net.Listen("unix", config.Hooks.Socket)
		if err != nil {
			clog.Fatal(1, "Listening hooks: "+err.Error())
		}
	} else if err != nil {
		clog.Fatal(1, "Listening hooks: "+err.Error())
	}
	clog.Info("Listening hooks on unix " + config.Hooks.Socket)
	go func() {
		if err := serve_git_hooks(hooks_listener); err != nil {
			clog.Fatal(1, "Serving hooks: "+err.Error())
		}
	}()

	// SSH listener
	ssh_listener, err := net.Listen(config.SSH.Net, config.SSH.Addr)
	if errors.Is(err, syscall.EADDRINUSE) && config.SSH.Net == "unix" {
		clog.Warn("Removing stale socket " + config.SSH.Addr)
		if err := syscall.Unlink(config.SSH.Addr); err != nil {
			clog.Fatal(1, "Removing stale socket: "+err.Error())
		}
		ssh_listener, err = net.Listen(config.SSH.Net, config.SSH.Addr)
		if err != nil {
			clog.Fatal(1, "Listening SSH: "+err.Error())
		}
	} else if err != nil {
		clog.Fatal(1, "Listening SSH: "+err.Error())
	}
	clog.Info("Listening SSH on " + config.SSH.Net + " " + config.SSH.Addr)
	go func() {
		if err := serve_ssh(ssh_listener); err != nil {
			clog.Fatal(1, "Serving SSH: "+err.Error())
		}
	}()

	// HTTP listener
	http_listener, err := net.Listen(config.HTTP.Net, config.HTTP.Addr)
	if errors.Is(err, syscall.EADDRINUSE) && config.HTTP.Net == "unix" {
		clog.Warn("Removing stale socket " + config.HTTP.Addr)
		if err := syscall.Unlink(config.HTTP.Addr); err != nil {
			clog.Fatal(1, "Removing stale socket: "+err.Error())
		}
		http_listener, err = net.Listen(config.HTTP.Net, config.HTTP.Addr)
		if err != nil {
			clog.Fatal(1, "Listening HTTP: "+err.Error())
		}
	} else if err != nil {
		clog.Fatal(1, "Listening HTTP: "+err.Error())
	}
	clog.Info("Listening HTTP on " + config.HTTP.Net + " " + config.HTTP.Addr)
	go func() {
		if err := http.Serve(http_listener, &http_router_t{}); err != nil {
			clog.Fatal(1, "Serving HTTP: "+err.Error())
		}
	}()

	select {}
}
