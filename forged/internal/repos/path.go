// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package repos

import "errors"

// ErrInvalidRepoPath is returned when a path does not point to a repository.
var ErrInvalidRepoPath = errors.New("invalid repo path")

// ExtractRepo parses path segments of the form group/.../-/repos/<repoName>/...
// and returns the group path, repo name, and index of the repo name segment.
func ExtractRepo(segments []string) (groupPath []string, repoName string, repoIndex int, err error) {
	sepIndex := -1
	for i, part := range segments {
		if part == "-" {
			sepIndex = i
			break
		}
	}
	if sepIndex == -1 || len(segments) <= sepIndex+2 || segments[sepIndex+1] != "repos" {
		return nil, "", 0, ErrInvalidRepoPath
	}
	groupPath = segments[:sepIndex]
	repoName = segments[sepIndex+2]
	repoIndex = sepIndex + 2
	return
}
