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
	segments, err = misc.PathToSegments(sshPath)
	if err != nil {
		return
	}
	for i, segment := range segments {
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return []string{}, "", 0, "", false, "", "", 0, err
		}
	}
	var repoIndex int
	if groupPath, repoName, repoIndex, err = repos.ExtractRepo(segments); err != nil {
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	}
	_ = repoIndex // index not used but kept for consistency
	repoID, repoPath, directAccess, contribReq, userType, userID, err = repos.GetInfo(ctx, s.database, groupPath, repoName, sshPubkey)
	if err != nil {
		return []string{}, "", 0, "", false, "", "", 0, err
	}
	return
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
