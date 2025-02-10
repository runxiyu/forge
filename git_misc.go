package main

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func open_git_repo(category_name, repo_name string) (*git.Repository, error) {
	return git.PlainOpen(filepath.Join(config.Git.Root, category_name, repo_name+".git"))
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
		display_git_tree_entry.Name = entry.Name
		display_git_tree = append(display_git_tree, display_git_tree_entry)
	}
	return display_git_tree
}
