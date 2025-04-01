// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"net/url"
	"strings"
)

// We don't use path.Join because it collapses multiple slashes into one.

// genSSHRemoteURL generates SSH remote URLs from a given group path and repo
// name.
func genSSHRemoteURL(groupPath []string, repoName string) string {
	return strings.TrimSuffix(config.SSH.Root, "/") + "/" + segmentsToURL(groupPath) + "/-/repos/" + url.PathEscape(repoName)
}

// genHTTPRemoteURL generates HTTP remote URLs from a given group path and repo
// name.
func genHTTPRemoteURL(groupPath []string, repoName string) string {
	return strings.TrimSuffix(config.HTTP.Root, "/") + "/" + segmentsToURL(groupPath) + "/-/repos/" + url.PathEscape(repoName)
}
