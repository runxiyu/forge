// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package ssh

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"

	"go.lindenii.runxiyu.org/forge/forged/internal/ansiec"
	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/repos"
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
		_1, _2, _3, _4, _5, _6, _7 := repos.GetInfo(ctx, s.database, groupPath, moduleName, sshPubkey)
		return groupPath, repoName, _1, _2, _3, _4, _5, _6, _7
	default:
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	}
}

// WriteRedError is a helper function that basically does a Fprintf but makes
// the entire thing red, in terms of ANSI escape sequences. It's useful when
// producing error messages on SSH connections.
// WriteRedError writes a formatted error in red ANSI color.
func WriteRedError(w io.Writer, format string, args ...any) {
	fmt.Fprintln(w, ansiec.Red+fmt.Sprintf(format, args...)+ansiec.Reset)
}

// randomUrlsafeStr generates a random string of the given entropic size
// using the URL-safe base64 encoding. The actual size of the string returned
// will be 4*sz.
func randomUrlsafeStr(sz int) (string, error) {
	r := make([]byte, 3*sz)
	if _, err := rand.Read(r); err != nil {
		return "", fmt.Errorf("error generating random string: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(r), nil
}
