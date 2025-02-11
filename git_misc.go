package main

import (
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.lindenii.runxiyu.org/lindenii-common/misc"
)

var err_unsafe_path = errors.New("Unsafe path")
var err_getting_commit_tree = errors.New("Error getting commit tree")
var err_getting_patch_of_commit = errors.New("Error getting patch of commit")
var err_getting_parent_commit_object = errors.New("Error getting parent commit object")

func open_git_repo(group_name, repo_name string) (*git.Repository, error) {
	group_name, group_name_ok := misc.Sanitize_path(group_name)
	if !group_name_ok {
		return nil, err_unsafe_path
	}
	repo_name, repo_name_ok := misc.Sanitize_path(repo_name)
	if !repo_name_ok {
		return nil, err_unsafe_path
	}
	return git.PlainOpen(filepath.Join(config.Git.Root, group_name, repo_name+".git"))
}

func build_display_git_tree(tree *object.Tree) []display_git_tree_entry_t {
	display_git_tree := make([]display_git_tree_entry_t, 0)
	for _, entry := range tree.Entries {
		display_git_tree_entry := display_git_tree_entry_t{}
		os_mode, err := entry.Mode.ToOSFileMode()
		if err != nil {
			display_git_tree_entry.Mode = "x---"
		} else {
			display_git_tree_entry.Mode = os_mode.String()[:4]
		}
		display_git_tree_entry.Is_file = entry.Mode.IsFile()
		display_git_tree_entry.Size, err = tree.Size(entry.Name)
		if err != nil {
			display_git_tree_entry.Size = 0
		}
		display_git_tree_entry.Name = strings.TrimPrefix(entry.Name, "/")
		display_git_tree = append(display_git_tree, display_git_tree_entry)
	}
	return display_git_tree
}

var err_get_recent_commits = errors.New("Error getting recent commits:")

func get_recent_commits(repo *git.Repository, head_hash plumbing.Hash, number_of_commits int) (recent_commits []*object.Commit, err error) {
	commit_iter, err := repo.Log(&git.LogOptions{From: head_hash})
	if err != nil {
		err = misc.Wrap_one_error(err_get_recent_commits, err)
		return nil, err
	}
	recent_commits = make([]*object.Commit, 0)
	defer commit_iter.Close()
	if number_of_commits < 0 {
		for {
			this_recent_commit, err := commit_iter.Next()
			if errors.Is(err, io.EOF) {
				return recent_commits, nil
			} else if err != nil {
				err = misc.Wrap_one_error(err_get_recent_commits, err)
				return nil, err
			}
			recent_commits = append(recent_commits, this_recent_commit)
		}
	} else {
		for range number_of_commits {
			this_recent_commit, err := commit_iter.Next()
			if errors.Is(err, io.EOF) {
				return recent_commits, nil
			} else if err != nil {
				err = misc.Wrap_one_error(err_get_recent_commits, err)
				return nil, err
			}
			recent_commits = append(recent_commits, this_recent_commit)
		}
	}
	return recent_commits, err
}

func get_patch_from_commit(commit_object *object.Commit) (parent_commit_hash plumbing.Hash, patch *object.Patch, ret_err error) {
	parent_commit_object, err := commit_object.Parent(0)
	if errors.Is(err, object.ErrParentNotFound) {
		commit_tree, err := commit_object.Tree()
		if err != nil {
			ret_err = misc.Wrap_one_error(err_getting_commit_tree, err)
			return
		}
		patch, err = (&object.Tree{}).Patch(commit_tree)
		if err != nil {
			ret_err = misc.Wrap_one_error(err_getting_patch_of_commit, err)
			return
		}
	} else if err != nil {
		ret_err = misc.Wrap_one_error(err_getting_parent_commit_object, err)
		return
	} else {
		parent_commit_hash = parent_commit_object.Hash
		patch, err = parent_commit_object.Patch(commit_object)
		if err != nil {
			ret_err = misc.Wrap_one_error(err_getting_patch_of_commit, err)
			return
		}
	}
	return
}
