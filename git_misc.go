// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jackc/pgx/v5/pgtype"
)

// openRepo opens a git repository by group and repo name.
func openRepo(ctx context.Context, groupPath []string, repoName string) (repo *git.Repository, description string, repoID int, err error) {
	var fsPath string

	err = database.QueryRow(ctx, `
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
type displayTreeEntry struct {
	Name      string
	Mode      string
	Size      int64
	IsFile    bool
	IsSubtree bool
}

func makeDisplayTree(tree *object.Tree) (displayTree []displayTreeEntry) {
	for _, entry := range tree.Entries {
		displayEntry := displayTreeEntry{}
		var err error
		var osMode os.FileMode

		if osMode, err = entry.Mode.ToOSFileMode(); err != nil {
			displayEntry.Mode = "x---------"
		} else {
			displayEntry.Mode = osMode.String()
		}

		displayEntry.IsFile = entry.Mode.IsFile()

		if displayEntry.Size, err = tree.Size(entry.Name); err != nil {
			displayEntry.Size = 0
		}

		displayEntry.Name = strings.TrimPrefix(entry.Name, "/")

		displayTree = append(displayTree, displayEntry)
	}
	return displayTree
}

func getRecentCommits(repo *git.Repository, headHash plumbing.Hash, numCommits int) (recentCommits []*object.Commit, err error) {
	var commitIter object.CommitIter
	var thisCommit *object.Commit

	commitIter, err = repo.Log(&git.LogOptions{From: headHash})
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

func fmtCommitAsPatch(commit *object.Commit) (parentCommitHash plumbing.Hash, patch *object.Patch, err error) {
	var parentCommit *object.Commit
	var commitTree *object.Tree

	parentCommit, err = commit.Parent(0)
	if errors.Is(err, object.ErrParentNotFound) {
		if commitTree, err = commit.Tree(); err != nil {
			return
		}
		if patch, err = (&object.Tree{}).Patch(commitTree); err != nil {
			return
		}
	} else if err != nil {
		return
	} else {
		parentCommitHash = parentCommit.Hash
		if patch, err = parentCommit.Patch(commit); err != nil {
			return
		}
	}
	return
}
