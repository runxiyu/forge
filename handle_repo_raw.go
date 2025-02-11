package main

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_raw(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	raw_path_spec := r.PathValue("rest")
	group_name, repo_name, path_spec := r.PathValue("group_name"), r.PathValue("repo_name"), strings.TrimSuffix(raw_path_spec, "/")

	ref_type, ref_name, err := get_param_ref_and_type(r)
	if err != nil {
		if errors.Is(err, err_no_ref_spec) {
			ref_type = "head"
		} else {
			_, _ = w.Write([]byte("Error querying ref type: " + err.Error()))
			return
		}
	}

	data["ref_type"], data["ref"], data["group_name"], data["repo_name"], data["path_spec"] = ref_type, ref_name, group_name, repo_name, path_spec

	repo, err := open_git_repo(group_name, repo_name)
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}

	ref_hash, err := get_ref_hash_from_type_and_name(repo, ref_type, ref_name)
	if err != nil {
		_, _ = w.Write([]byte("Error getting ref hash: " + err.Error()))
		return
	}

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
