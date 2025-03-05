// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var repo *git.Repository
	var repo_name, group_name string
	var ref_hash plumbing.Hash
	var err error
	var recent_commits []*object.Commit
	var commit_object *object.Commit
	var tree *object.Tree

	repo, repo_name, group_name = params["repo"].(*git.Repository), params["repo_name"].(string), params["group_name"].(string)

	if ref_hash, err = get_ref_hash_from_type_and_name(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if recent_commits, err = get_recent_commits(repo, ref_hash, 3); err != nil {
		http.Error(w, "Error getting recent commits: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["commits"] = recent_commits
	commit_object, err = repo.CommitObject(ref_hash)
	if err != nil {
		http.Error(w, "Error getting commit object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tree, err = commit_object.Tree()
	if err != nil {
		http.Error(w, "Error getting file tree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	params["readme_filename"], params["readme"] = render_readme_at_tree(tree)
	params["files"] = build_display_git_tree(tree)

	params["http_clone_url"] = generate_http_remote_url(group_name, repo_name)
	params["ssh_clone_url"] = generate_ssh_remote_url(group_name, repo_name)

	render_template(w, "repo_index", params)
}
