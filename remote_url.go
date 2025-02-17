package main

import (
	"net/url"
	"path"
)

func generate_ssh_remote_url(group_name, repo_name string) string {
	return path.Join(config.SSH.Root, url.PathEscape(group_name), "/:/repos/", url.PathEscape(repo_name))
}

func generate_http_remote_url(group_name, repo_name string) string {
	return path.Join(config.HTTP.Root, url.PathEscape(group_name), "/:/repos/", url.PathEscape(repo_name))
}
