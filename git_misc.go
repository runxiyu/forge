package main

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.lindenii.runxiyu.org/lindenii-common/misc"
)

func open_git_repo(group_name, repo_name string) (*git.Repository, error) {
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

func get_recent_commits(repo *git.Repository, head_hash plumbing.Hash) (recent_commits []*object.Commit, err error) {
	commit_iter, err := repo.Log(&git.LogOptions{From: head_hash})
	if err != nil {
		err = misc.Wrap_one_error(err_get_recent_commits, err)
		return nil, err
	}
	recent_commits = make([]*object.Commit, 0)
	defer commit_iter.Close()
	for range 3 {
		this_recent_commit, err := commit_iter.Next()
		if err != nil {
			err = misc.Wrap_one_error(err_get_recent_commits, err)
			return nil, err
		}
		recent_commits = append(recent_commits, this_recent_commit)
	}
	return
}
