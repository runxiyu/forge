package main

type display_git_tree_entry_t struct {
	Name       string
	Mode       string
	Size       int64
	Is_file    bool
	Is_subtree bool
}
