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
func httpHandleRepoLog(writer http.ResponseWriter, _ *http.Request, params map[string]any) {
	var repo *git.Repository
	var refHash plumbing.Hash
	var err error
	var commits []*object.Commit

	repo = params["repo"].(*git.Repository)

	if refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		errorPage500(writer, params, "Error getting ref hash: "+err.Error())
		return
	}

	if commits, err = getRecentCommits(repo, refHash, -1); err != nil {
		errorPage500(writer, params, "Error getting recent commits: "+err.Error())
		return
	}
	params["commits"] = commits

	renderTemplate(writer, "repo_log", params)
}
