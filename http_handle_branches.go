// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// httpHandleRepoBranches provides the branches page in repos.
func httpHandleRepoBranches(writer http.ResponseWriter, _ *http.Request, params map[string]any) {
	var repo *git.Repository
	var repoName string
	var groupPath []string
	var err error
	var notes []string
	var branches []string
	var branchesIter storer.ReferenceIter

	repo, repoName, groupPath = params["repo"].(*git.Repository), params["repo_name"].(string), params["group_path"].([]string)

	if strings.Contains(repoName, "\n") || sliceContainsNewlines(groupPath) {
		notes = append(notes, "Path contains newlines; HTTP Git access impossible")
	}

	branchesIter, err = repo.Branches()
	if err == nil {
		_ = branchesIter.ForEach(func(branch *plumbing.Reference) error {
			branches = append(branches, branch.Name().Short())
			return nil
		})
	}
	params["branches"] = branches

	params["http_clone_url"] = genHTTPRemoteURL(groupPath, repoName)
	params["ssh_clone_url"] = genSSHRemoteURL(groupPath, repoName)
	params["notes"] = notes

	renderTemplate(writer, "repo_branches", params)
}
