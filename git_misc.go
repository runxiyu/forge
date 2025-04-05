// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"context"
	"errors"
	"io"
	"iter"
	"os"
	"strings"

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
func (s *server) openRepo(ctx context.Context, groupPath []string, repoName string) (repo *git.Repository, description string, repoID int, fsPath string, err error) {
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

// go-git's tree entries are not friendly for use in HTML templates.
// This struct is a wrapper that is friendlier for use in templating.
type displayTreeEntry struct {
	Name      string
	Mode      string
	Size      uint64
	IsFile    bool
	IsSubtree bool
}

// makeDisplayTree takes git trees of form [object.Tree] and creates a slice of
// [displayTreeEntry] for easier templating.
func makeDisplayTree(tree *object.Tree) (displayTree []displayTreeEntry) {
	for _, entry := range tree.Entries {
		displayEntry := displayTreeEntry{} //exhaustruct:ignore
		var err error
		var osMode os.FileMode

		if osMode, err = entry.Mode.ToOSFileMode(); err != nil {
			displayEntry.Mode = "x---------"
		} else {
			displayEntry.Mode = osMode.String()
		}

		displayEntry.IsFile = entry.Mode.IsFile()

		size, _ := tree.Size(entry.Name)
		displayEntry.Size = uint64(size) //#nosec G115

		displayEntry.Name = strings.TrimPrefix(entry.Name, "/")

		displayTree = append(displayTree, displayEntry)
	}
	return displayTree
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

// getRecentCommits fetches numCommits commits, starting from the headHash in a
// repo.
func getRecentCommits(repo *git.Repository, headHash plumbing.Hash, numCommits int) (recentCommits []*object.Commit, err error) {
	var commitIter object.CommitIter
	var thisCommit *object.Commit

	commitIter, err = repo.Log(&git.LogOptions{From: headHash}) //exhaustruct:ignore
	if err != nil {
		return nil, err
	}
	recentCommits = make([]*object.Commit, 0)
	defer commitIter.Close()
	if numCommits < 0 {
		for {
			thisCommit, err = commitIter.Next()
			if errors.Is(err, io.EOF) {
				return recentCommits, nil
			} else if err != nil {
				return nil, err
			}
			recentCommits = append(recentCommits, thisCommit)
		}
	} else {
		for range numCommits {
			thisCommit, err = commitIter.Next()
			if errors.Is(err, io.EOF) {
				return recentCommits, nil
			} else if err != nil {
				return nil, err
			}
			recentCommits = append(recentCommits, thisCommit)
		}
	}
	return recentCommits, err
}

// getRecentCommitsDisplay generates a slice of [commitDisplay] friendly for
// use in HTML templates, consisting of numCommits commits from headhash in the
// repo.
func getRecentCommitsDisplay(repo *git.Repository, headHash plumbing.Hash, numCommits int) (recentCommits []commitDisplayOld, err error) {
	var commitIter object.CommitIter
	var thisCommit *object.Commit

	commitIter, err = repo.Log(&git.LogOptions{From: headHash}) //exhaustruct:ignore
	if err != nil {
		return nil, err
	}
	recentCommits = make([]commitDisplayOld, 0)
	defer commitIter.Close()
	if numCommits < 0 {
		for {
			thisCommit, err = commitIter.Next()
			if errors.Is(err, io.EOF) {
				return recentCommits, nil
			} else if err != nil {
				return nil, err
			}
			recentCommits = append(recentCommits, commitDisplayOld{
				thisCommit.Hash,
				thisCommit.Author,
				thisCommit.Committer,
				thisCommit.Message,
				thisCommit.TreeHash,
			})
		}
	} else {
		for range numCommits {
			thisCommit, err = commitIter.Next()
			if errors.Is(err, io.EOF) {
				return recentCommits, nil
			} else if err != nil {
				return nil, err
			}
			recentCommits = append(recentCommits, commitDisplayOld{
				thisCommit.Hash,
				thisCommit.Author,
				thisCommit.Committer,
				thisCommit.Message,
				thisCommit.TreeHash,
			})
		}
	}
	return recentCommits, err
}

type commitDisplayOld struct {
	Hash      plumbing.Hash
	Author    object.Signature
	Committer object.Signature
	Message   string
	TreeHash  plumbing.Hash
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
