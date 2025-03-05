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
)

// open_git_repo opens a git repository by group and repo name.
func open_git_repo(ctx context.Context, group_name, repo_name string) (repo *git.Repository, description string, repo_id int, err error) {
	var fs_path string
	err = database.QueryRow(ctx,
		"SELECT r.filesystem_path, COALESCE(r.description, ''), r.id FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1 AND r.name = $2;",
		group_name, repo_name,
	).Scan(&fs_path, &description, &repo_id)
	if err != nil {
		return
	}
	repo, err = git.PlainOpen(fs_path)
	return
}

// go-git's tree entries are not friendly for use in HTML templates.
type display_git_tree_entry_t struct {
	Name       string
	Mode       string
	Size       int64
	Is_file    bool
	Is_subtree bool
}

func build_display_git_tree(tree *object.Tree) (display_git_tree []display_git_tree_entry_t) {
	for _, entry := range tree.Entries {
		display_git_tree_entry := display_git_tree_entry_t{}
		var err error
		var os_mode os.FileMode

		if os_mode, err = entry.Mode.ToOSFileMode(); err != nil {
			display_git_tree_entry.Mode = "x---------"
		} else {
			display_git_tree_entry.Mode = os_mode.String()
		}

		display_git_tree_entry.Is_file = entry.Mode.IsFile()

		if display_git_tree_entry.Size, err = tree.Size(entry.Name); err != nil {
			display_git_tree_entry.Size = 0
		}

		display_git_tree_entry.Name = strings.TrimPrefix(entry.Name, "/")

		display_git_tree = append(display_git_tree, display_git_tree_entry)
	}
	return display_git_tree
}

func get_recent_commits(repo *git.Repository, head_hash plumbing.Hash, number_of_commits int) (recent_commits []*object.Commit, err error) {
	var commit_iter object.CommitIter
	var this_recent_commit *object.Commit

	commit_iter, err = repo.Log(&git.LogOptions{From: head_hash})
	if err != nil {
		return nil, err
	}
	recent_commits = make([]*object.Commit, 0)
	defer commit_iter.Close()
	if number_of_commits < 0 {
		for {
			this_recent_commit, err = commit_iter.Next()
			if errors.Is(err, io.EOF) {
				return recent_commits, nil
			} else if err != nil {
				return nil, err
			}
			recent_commits = append(recent_commits, this_recent_commit)
		}
	} else {
		for range number_of_commits {
			this_recent_commit, err = commit_iter.Next()
			if errors.Is(err, io.EOF) {
				return recent_commits, nil
			} else if err != nil {
				return nil, err
			}
			recent_commits = append(recent_commits, this_recent_commit)
		}
	}
	return recent_commits, err
}

func get_patch_from_commit(commit_object *object.Commit) (parent_commit_hash plumbing.Hash, patch *object.Patch, err error) {
	var parent_commit_object *object.Commit
	var commit_tree *object.Tree

	parent_commit_object, err = commit_object.Parent(0)
	if errors.Is(err, object.ErrParentNotFound) {
		if commit_tree, err = commit_object.Tree(); err != nil {
			return
		}
		if patch, err = (&object.Tree{}).Patch(commit_tree); err != nil {
			return
		}
	} else if err != nil {
		return
	} else {
		parent_commit_hash = parent_commit_object.Hash
		if patch, err = parent_commit_object.Patch(commit_object); err != nil {
			return
		}
	}
	return
}
