// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	"go.lindenii.runxiyu.org/forge/internal/ansiec"
	"go.lindenii.runxiyu.org/forge/internal/misc"
)

var errIllegalSSHRepoPath = errors.New("illegal SSH repo path")

// getRepoInfo2 also fetches repo information... it should be deprecated and
// implemented in individual handlers.
func (s *Server) getRepoInfo2(ctx context.Context, sshPath, sshPubkey string) (groupPath []string, repoName string, repoID int, repoPath string, directAccess bool, contribReq, userType string, userID int, err error) {
	var segments []string
	var sepIndex int
	var moduleType, moduleName string

	segments, err = misc.PathToSegments(sshPath)
	if err != nil {
		return
	}

	for i, segment := range segments {
		var err error
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return []string{}, "", 0, "", false, "", "", 0, err
		}
	}

	if segments[0] == "-" {
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	}

	sepIndex = -1
	for i, part := range segments {
		if part == "-" {
			sepIndex = i
			break
		}
	}
	if segments[len(segments)-1] == "" {
		segments = segments[:len(segments)-1]
	}

	switch {
	case sepIndex == -1:
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	case len(segments) <= sepIndex+2:
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	}

	groupPath = segments[:sepIndex]
	moduleType = segments[sepIndex+1]
	moduleName = segments[sepIndex+2]
	repoName = moduleName
	switch moduleType {
	case "repos":
		_1, _2, _3, _4, _5, _6, _7 := s.getRepoInfo(ctx, groupPath, moduleName, sshPubkey)
		return groupPath, repoName, _1, _2, _3, _4, _5, _6, _7
	default:
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	}
}

// writeRedError is a helper function that basically does a Fprintf but makes
// the entire thing red, in terms of ANSI escape sequences. It's useful when
// producing error messages on SSH connections.
func writeRedError(w io.Writer, format string, args ...any) {
	fmt.Fprintln(w, ansiec.Red+fmt.Sprintf(format, args...)+ansiec.Reset)
}
