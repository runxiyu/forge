package main

import (
	"net/url"
	"strings"
)

// We don't use path.Join because it collapses multiple slashes into one.

func generate_ssh_remote_url(group_name, repo_name string) string {
	return strings.TrimSuffix(config.SSH.Root, "/") + "/" + url.PathEscape(group_name) + "/:/repos/" + url.PathEscape(repo_name)
}

func generate_http_remote_url(group_name, repo_name string) string {
	return strings.TrimSuffix(config.HTTP.Root, "/") + "/" + url.PathEscape(group_name) + "/:/repos/" + url.PathEscape(repo_name)
}
