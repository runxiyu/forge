// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_raw(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var raw_path_spec, path_spec string
	var repo *git.Repository
	var ref_hash plumbing.Hash
	var commit_object *object.Commit
	var tree *object.Tree
	var err error

	raw_path_spec = params["rest"].(string)
	repo, path_spec = params["repo"].(*git.Repository), strings.TrimSuffix(raw_path_spec, "/")
	params["path_spec"] = path_spec

	if ref_hash, err = get_ref_hash_from_type_and_name(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if commit_object, err = repo.CommitObject(ref_hash); err != nil {
		http.Error(w, "Error getting commit object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tree, err = commit_object.Tree(); err != nil {
		http.Error(w, "Error getting file tree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var target *object.Tree
	if path_spec == "" {
		target = tree
	} else {
		if target, err = tree.Tree(path_spec); err != nil {
			var file *object.File
			var file_contents string
			if file, err = tree.File(path_spec); err != nil {
				http.Error(w, "Error retrieving path: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] == '/' {
				http.Redirect(w, r, "../"+path_spec, http.StatusSeeOther)
				return
			}
			if file_contents, err = file.Contents(); err != nil {
				http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, file_contents)
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
