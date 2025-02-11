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

	err = serve_static()
	if err != nil {
		clog.Fatal(1, "Serving static: "+err.Error())
	}

	serve_source()

	http.HandleFunc("/{$}", handle_index)
	http.HandleFunc("/g/{group_name}/repos/{$}", handle_group_repos)
	http.HandleFunc("/g/{group_name}/repos/{repo_name}/{$}", handle_repo_index)
	http.HandleFunc("/g/{group_name}/repos/{repo_name}/tree/{rest...}", handle_repo_tree)
	http.HandleFunc("/g/{group_name}/repos/{repo_name}/raw/{rest...}", handle_repo_raw)
	http.HandleFunc("/g/{group_name}/repos/{repo_name}/log/{ref}/", handle_repo_log)
	http.HandleFunc("/g/{group_name}/repos/{repo_name}/commit/{commit_id}", handle_repo_commit)

	listener, err := net.Listen(config.HTTP.Net, config.HTTP.Addr)
	if err != nil {
		clog.Fatal(1, "Listening: "+err.Error())
	}

	err = http.Serve(listener, nil)
	if err != nil {
		clog.Fatal(1, "Serving: "+err.Error())
	}
}
