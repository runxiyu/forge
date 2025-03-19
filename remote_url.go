// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/url"
	"strings"
)

// We don't use path.Join because it collapses multiple slashes into one.

func genSSHRemoteURL(group_path []string, repo_name string) string {
	return strings.TrimSuffix(config.SSH.Root, "/") + "/" + segmentsToURL(group_path) + "/:/repos/" + url.PathEscape(repo_name)
}

func genHTTPRemoteURL(group_path []string, repo_name string) string {
	return strings.TrimSuffix(config.HTTP.Root, "/") + "/" + segmentsToURL(group_path) + "/:/repos/" + url.PathEscape(repo_name)
}
