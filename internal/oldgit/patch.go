// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package oldgit

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitToPatch creates an [object.Patch] from the first parent of a given
// [object.Commit].
//
// TODO: This function should be deprecated as it only diffs with the first
// parent and does not correctly handle merge commits.
func CommitToPatch(commit *object.Commit) (parentCommitHash plumbing.Hash, patch *object.Patch, err error) {
	var parentCommit *object.Commit
	var commitTree *object.Tree

	parentCommit, err = commit.Parent(0)
	switch {
	case errors.Is(err, object.ErrParentNotFound):
		if commitTree, err = commit.Tree(); err != nil {
			return
		}
		if patch, err = NullTree.Patch(commitTree); err != nil {
			return
		}
	case err != nil:
		return
	default:
		parentCommitHash = parentCommit.Hash
		if patch, err = parentCommit.Patch(commit); err != nil {
			return
		}
	}
	return
}

var NullTree object.Tree //nolint:gochecknoglobals
