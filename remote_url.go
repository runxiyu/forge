// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"net/url"
	"strings"

	"go.lindenii.runxiyu.org/forge/misc"
)

// We don't use path.Join because it collapses multiple slashes into one.

// genSSHRemoteURL generates SSH remote URLs from a given group path and repo
// name.
func (s *Server) genSSHRemoteURL(groupPath []string, repoName string) string {
	return strings.TrimSuffix(s.Config.SSH.Root, "/") + "/" + misc.SegmentsToURL(groupPath) + "/-/repos/" + url.PathEscape(repoName)
}

// genHTTPRemoteURL generates HTTP remote URLs from a given group path and repo
// name.
func (s *Server) genHTTPRemoteURL(groupPath []string, repoName string) string {
	return strings.TrimSuffix(s.Config.HTTP.Root, "/") + "/" + misc.SegmentsToURL(groupPath) + "/-/repos/" + url.PathEscape(repoName)
}
