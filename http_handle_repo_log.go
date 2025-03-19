// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TODO: I probably shouldn't include *all* commits here...
func httpHandleRepoLog(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var repo *git.Repository
	var ref_hash plumbing.Hash
	var err error
	var commits []*object.Commit

	repo = params["repo"].(*git.Repository)

	if ref_hash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if commits, err = getRecentCommits(repo, ref_hash, -1); err != nil {
		http.Error(w, "Error getting recent commits: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["commits"] = commits

	renderTemplate(w, "repo_log", params)
}
