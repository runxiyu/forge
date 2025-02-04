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

	err = load_templates()
	if err != nil {
		clog.Fatal(1, "Loading templates: "+err.Error())
	}

	http.HandleFunc("/{$}", handle_index)

	listener, err := net.Listen(config.HTTP.Net, config.HTTP.Addr)
	if err != nil {
		clog.Fatal(1, "Listening: "+err.Error())
	}

	http.Serve(listener, nil)
}
