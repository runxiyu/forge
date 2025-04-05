// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"context"
	"errors"
	"io"
	"iter"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jackc/pgx/v5/pgtype"
)

// openRepo opens a git repository by group and repo name.
//
// TODO: This should be deprecated in favor of doing it in the relevant
// request/router context in the future, as it cannot cover the nuance of
// fields needed.
func (s *Server) openRepo(ctx context.Context, groupPath []string, repoName string) (repo *git.Repository, description string, repoID int, fsPath string, err error) {
	err = s.database.QueryRow(ctx, `
WITH RECURSIVE group_path_cte AS (
	-- Start: match the first name in the path where parent_group IS NULL
	SELECT
		id,
		parent_group,
		name,
		1 AS depth
	FROM groups
	WHERE name = ($1::text[])[1]
		AND parent_group IS NULL

	UNION ALL

	-- Recurse: join next segment of the path
	SELECT
		g.id,
		g.parent_group,
		g.name,
		group_path_cte.depth + 1
	FROM groups g
	JOIN group_path_cte ON g.parent_group = group_path_cte.id
	WHERE g.name = ($1::text[])[group_path_cte.depth + 1]
		AND group_path_cte.depth + 1 <= cardinality($1::text[])
)
SELECT
	r.filesystem_path,
	COALESCE(r.description, ''),
	r.id
FROM group_path_cte g
JOIN repos r ON r.group_id = g.id
WHERE g.depth = cardinality($1::text[])
	AND r.name = $2
	`, pgtype.FlatArray[string](groupPath), repoName).Scan(&fsPath, &description, &repoID)
	if err != nil {
		return
	}

	repo, err = git.PlainOpen(fsPath)
	return
}

// commitIterSeqErr creates an [iter.Seq[*object.Commit]] from an
// [object.CommitIter], and additionally returns a pointer to error.
// The pointer to error is guaranteed to be populated with either nil or the
// error returned by the commit iterator after the returned iterator is
// finished.
func commitIterSeqErr(commitIter object.CommitIter) (iter.Seq[*object.Commit], *error) {
	var err error
	return func(yield func(*object.Commit) bool) {
		for {
			commit, err2 := commitIter.Next()
			if err2 != nil {
				if errors.Is(err2, io.EOF) {
					return
				}
				err = err2
				return
			}
			if !yield(commit) {
				return
			}
		}
	}, &err
}

// commitToPatch creates an [object.Patch] from the first parent of a given
// [object.Commit].
//
// TODO: This function should be deprecated as it only diffs with the first
// parent and does not correctly handle merge commits.
func commitToPatch(commit *object.Commit) (parentCommitHash plumbing.Hash, patch *object.Patch, err error) {
	var parentCommit *object.Commit
	var commitTree *object.Tree

	parentCommit, err = commit.Parent(0)
	switch {
	case errors.Is(err, object.ErrParentNotFound):
		if commitTree, err = commit.Tree(); err != nil {
			return
		}
		if patch, err = nullTree.Patch(commitTree); err != nil {
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

var nullTree object.Tree //nolint:gochecknoglobals
