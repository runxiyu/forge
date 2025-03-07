// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var repo *git.Repository
	var repo_name string
	var group_path []string
	var ref_hash plumbing.Hash
	var err error
	var recent_commits []*object.Commit
	var commit_object *object.Commit
	var tree *object.Tree
	var notes []string

	repo, repo_name, group_path = params["repo"].(*git.Repository), params["repo_name"].(string), params["group_path"].([]string)

	if strings.Contains(repo_name, "\n") || slice_contains_newline(group_path) {
		notes = append(notes, "Path contains newlines; HTTP Git access impossible")
	}

	ref_hash, err = get_ref_hash_from_type_and_name(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		goto no_ref
	}

	if recent_commits, err = get_recent_commits(repo, ref_hash, 3); err != nil {
		goto no_ref
	}
	params["commits"] = recent_commits

	if commit_object, err = repo.CommitObject(ref_hash); err != nil {
		goto no_ref
	}

	if tree, err = commit_object.Tree(); err != nil {
		goto no_ref
	}

	params["files"] = build_display_git_tree(tree)
	params["readme_filename"], params["readme"] = render_readme_at_tree(tree)

no_ref:

	params["http_clone_url"] = generate_http_remote_url(group_path, repo_name)
	params["ssh_clone_url"] = generate_ssh_remote_url(group_path, repo_name)
	params["notes"] = notes

	render_template(w, "repo_index", params)
}
