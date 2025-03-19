// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"go.lindenii.runxiyu.org/lindenii-common/ansiec"
)

var errIllegalSSHRepoPath = errors.New("illegal SSH repo path")

func getRepoInfo2(ctx context.Context, sshPath, sshPubkey string) (groupPath []string, repoName string, repoID int, repoPath string, directAccess bool, contribReq, userType string, userID int, err error) {
	var segments []string
	var sepIndex int
	var moduleType, moduleName string

	segments = strings.Split(strings.TrimPrefix(sshPath, "/"), "/")

	for i, segment := range segments {
		var err error
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return []string{}, "", 0, "", false, "", "", 0, err
		}
	}

	if segments[0] == ":" {
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	}

	sepIndex = -1
	for i, part := range segments {
		if part == ":" {
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
		_1, _2, _3, _4, _5, _6, _7 := getRepoInfo(ctx, groupPath, moduleName, sshPubkey)
		return groupPath, repoName, _1, _2, _3, _4, _5, _6, _7
	default:
		return []string{}, "", 0, "", false, "", "", 0, errIllegalSSHRepoPath
	}
}

func writeRedError(w io.Writer, format string, args ...any) {
	fmt.Fprintln(w, ansiec.Red+fmt.Sprintf(format, args...)+ansiec.Reset)
}
