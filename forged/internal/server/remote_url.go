// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package server

import (
	"net/url"
	"strings"

	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

// genHTTPRemoteURL generates HTTP remote URLs from a given group path and repo name.
func (s *Server) genHTTPRemoteURL(groupPath []string, repoName string) string {
	return strings.TrimSuffix(s.config.HTTP.Root, "/") + "/" + misc.SegmentsToURL(groupPath) + "/-/repos/" + url.PathEscape(repoName)
}
