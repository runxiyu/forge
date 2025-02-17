package main

import (
	"flag"
	"net"
	"net/http"

	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

func main() {
	config_path := flag.String(
		"config",
		"/etc/lindenii/forge.scfg",
		"path to configuration file",
	)
	flag.Parse()

	err := load_config(*config_path)
	if err != nil {
		clog.Fatal(1, "Loading configuration: "+err.Error())
	}

	err = deploy_hooks_to_filesystem()
	if err != nil {
		clog.Fatal(1, "Deploying hooks to filesystem: "+err.Error())
	}

	err = load_templates()
	if err != nil {
		clog.Fatal(1, "Loading templates: "+err.Error())
	}

	ssh_listener, err := net.Listen(config.SSH.Net, config.SSH.Addr)
	if err != nil {
		clog.Fatal(1, "Listening SSH: "+err.Error())
	}
	clog.Info("Listening SSH on " + config.SSH.Net + " " + config.SSH.Addr)
	go func() {
		err = serve_ssh(ssh_listener)
		if err != nil {
			clog.Fatal(1, "Serving SSH: "+err.Error())
		}
	}()

	hooks_listener, err := net.Listen("unix", config.Hooks.Socket)
	if err != nil {
		clog.Fatal(1, "Listening hooks: "+err.Error())
	}
	clog.Info("Listening hooks at " + config.Hooks.Socket)
	go func() {
		err = serve_git_hooks(hooks_listener)
		if err != nil {
			clog.Fatal(1, "Serving hooks: "+err.Error())
		}
	}()

	listener, err := net.Listen(config.HTTP.Net, config.HTTP.Addr)
	if err != nil {
		clog.Fatal(1, "Listening HTTP: "+err.Error())
	}
	clog.Info("Listening HTTP on " + config.HTTP.Net + " " + config.HTTP.Addr)

	go func() {
		err = http.Serve(listener, &http_router_t{})
		if err != nil {
			clog.Fatal(1, "Serving HTTP: "+err.Error())
		}
	}()

	select{}
}
