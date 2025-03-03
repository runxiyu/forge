// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"

	"github.com/go-git/go-git/v5"
)

// TODO: I probably shouldn't include *all* commits here...
func handle_repo_log(w http.ResponseWriter, r *http.Request, params map[string]any) {
	repo := params["repo"].(*git.Repository)

	ref_hash, err := get_ref_hash_from_type_and_name(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	commits, err := get_recent_commits(repo, ref_hash, -1)
	if err != nil {
		http.Error(w, "Error getting recent commits: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["commits"] = commits

	render_template(w, "repo_log", params)
}
