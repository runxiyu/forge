// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

func httpHandleRepoIndex(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var repo *git.Repository
	var repo_name string
	var group_path []string
	var refHash plumbing.Hash
	var err error
	var recent_commits []*object.Commit
	var commit_object *object.Commit
	var tree *object.Tree
	var notes []string
	var branches []string
	var branches_ storer.ReferenceIter

	repo, repo_name, group_path = params["repo"].(*git.Repository), params["repo_name"].(string), params["group_path"].([]string)

	if strings.Contains(repo_name, "\n") || slice_contains_newline(group_path) {
		notes = append(notes, "Path contains newlines; HTTP Git access impossible")
	}

	refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		goto no_ref
	}

	branches_, err = repo.Branches()
	if err != nil {
	}
	err = branches_.ForEach(func(branch *plumbing.Reference) error {
		branches = append(branches, branch.Name().Short())
		return nil
	})
	if err != nil {
	}
	params["branches"] = branches

	if recent_commits, err = getRecentCommits(repo, refHash, 3); err != nil {
		goto no_ref
	}
	params["commits"] = recent_commits

	if commit_object, err = repo.CommitObject(refHash); err != nil {
		goto no_ref
	}

	if tree, err = commit_object.Tree(); err != nil {
		goto no_ref
	}

	params["files"] = makeDisplayTree(tree)
	params["readme_filename"], params["readme"] = renderReadmeAtTree(tree)

no_ref:

	params["http_clone_url"] = genHTTPRemoteURL(group_path, repo_name)
	params["ssh_clone_url"] = genSSHRemoteURL(group_path, repo_name)
	params["notes"] = notes

	renderTemplate(w, "repo_index", params)
}
