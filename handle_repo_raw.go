package main

import (
	"net/http"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_raw(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	raw_path_spec := r.PathValue("rest")
	ref_name, group_name, repo_name, path_spec := r.PathValue("ref"), r.PathValue("group_name"), r.PathValue("repo_name"), strings.TrimSuffix(raw_path_spec, "/")
	data["ref"], data["group_name"], data["repo_name"], data["path_spec"] = ref_name, group_name, repo_name, path_spec
	repo, err := open_git_repo(group_name, repo_name)
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(ref_name), true)
	if err != nil {
		_, _ = w.Write([]byte("Error getting repo reference: " + err.Error()))
		return
	}
	ref_hash := ref.Hash()
	commit_object, err := repo.CommitObject(ref_hash)
	if err != nil {
		_, _ = w.Write([]byte("Error getting commit object: " + err.Error()))
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		_, _ = w.Write([]byte("Error getting file tree: " + err.Error()))
		return
	}

	var target *object.Tree
	if path_spec == "" {
		target = tree
	} else {
		target, err = tree.Tree(path_spec)
		if err != nil {
			file, err := tree.File(path_spec)
			if err != nil {
				_, _ = w.Write([]byte("Error retrieving path: " + err.Error()))
				return
			}
			if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] == '/' {
				http.Redirect(w, r, "../"+path_spec, http.StatusSeeOther)
				return
			}
			file_contents, err := file.Contents()
			if err != nil {
				_, _ = w.Write([]byte("Error reading file: " + err.Error()))
				return
			}
			_, _ = w.Write([]byte(file_contents))
			return
		}
	}

	if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] != '/' {
		http.Redirect(w, r, path.Base(path_spec)+"/", http.StatusSeeOther)
		return
	}

	data["files"] = build_display_git_tree(target)

	err = templates.ExecuteTemplate(w, "repo_raw_dir", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
