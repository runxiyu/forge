// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/url"
	"strings"
)

// We don't use path.Join because it collapses multiple slashes into one.

func generate_ssh_remote_url(group_path []string, repo_name string) string {
	return strings.TrimSuffix(config.SSH.Root, "/") + "/" + path_escape_cat_segments(group_path) + "/:/repos/" + url.PathEscape(repo_name)
}

func generate_http_remote_url(group_path []string, repo_name string) string {
	return strings.TrimSuffix(config.HTTP.Root, "/") + "/" + path_escape_cat_segments(group_path) + "/:/repos/" + url.PathEscape(repo_name)
}
