package main

import (
	"fmt"
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_raw(w http.ResponseWriter, r *http.Request, params map[string]any) {
	raw_path_spec := params["rest"].(string)
	group_name, repo_name, path_spec := params["group_name"].(string), params["repo_name"].(string), strings.TrimSuffix(raw_path_spec, "/")

	ref_type, ref_name, err := get_param_ref_and_type(r)
	if err != nil {
		if errors.Is(err, err_no_ref_spec) {
			ref_type = "head"
		} else {
			fmt.Fprintln(w, "Error querying ref type:", err.Error())
			return
		}
	}

	params["ref_type"], params["ref"], params["path_spec"] = ref_type, ref_name, path_spec

	repo, err := open_git_repo(r.Context(), group_name, repo_name)
	if err != nil {
		fmt.Fprintln(w, "Error opening repo:", err.Error())
		return
	}

	ref_hash, err := get_ref_hash_from_type_and_name(repo, ref_type, ref_name)
	if err != nil {
		fmt.Fprintln(w, "Error getting ref hash:", err.Error())
		return
	}

	commit_object, err := repo.CommitObject(ref_hash)
	if err != nil {
		fmt.Fprintln(w, "Error getting commit object:", err.Error())
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		fmt.Fprintln(w, "Error getting file tree:", err.Error())
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
				fmt.Fprintln(w, "Error retrieving path:", err.Error())
				return
			}
			if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] == '/' {
				http.Redirect(w, r, "../"+path_spec, http.StatusSeeOther)
				return
			}
			file_contents, err := file.Contents()
			if err != nil {
				fmt.Fprintln(w, "Error reading file:", err.Error())
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

	err = templates.ExecuteTemplate(w, "repo_raw_dir", params)
	if err != nil {
		fmt.Fprintln(w, "Error rendering template:", err.Error())
		return
	}
}
