package main

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_raw(w http.ResponseWriter, r *http.Request, params map[string]any) {
	raw_path_spec := params["rest"].(string)
	group_name, repo_name, path_spec := params["group_name"].(string), params["repo_name"].(string), strings.TrimSuffix(raw_path_spec, "/")

	params["path_spec"] = path_spec

	repo, description, err := open_git_repo(r.Context(), group_name, repo_name)
	if err != nil {
		http.Error(w, "Error opening repo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["repo_description"] = description

	ref_hash, err := get_ref_hash_from_type_and_name(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	commit_object, err := repo.CommitObject(ref_hash)
	if err != nil {
		http.Error(w, "Error getting commit object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		http.Error(w, "Error getting file tree: "+err.Error(), http.StatusInternalServerError)
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
				http.Error(w, "Error retrieving path: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] == '/' {
				http.Redirect(w, r, "../"+path_spec, http.StatusSeeOther)
				return
			}
			file_contents, err := file.Contents()
			if err != nil {
				http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintln(w, file_contents)
			return
		}
	}

	if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] != '/' {
		http.Redirect(w, r, path.Base(path_spec)+"/", http.StatusSeeOther)
		return
	}

	params["files"] = build_display_git_tree(target)

	render_template(w, "repo_raw_dir", params)
}
