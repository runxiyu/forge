// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"iter"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

func httpHandleRepoIndex(w http.ResponseWriter, _ *http.Request, params map[string]any) {
	var repo *git.Repository
	var repoName string
	var groupPath []string
	var refHash plumbing.Hash
	var err error
	var commitIter object.CommitIter
	var commitIterSeq iter.Seq[*object.Commit]
	var commitObj *object.Commit
	var tree *object.Tree
	var notes []string
	var branches []string
	var branchesIter storer.ReferenceIter
	var logOptions git.LogOptions

	repo, repoName, groupPath = params["repo"].(*git.Repository), params["repo_name"].(string), params["group_path"].([]string)

	if strings.Contains(repoName, "\n") || sliceContainsNewlines(groupPath) {
		notes = append(notes, "Path contains newlines; HTTP Git access impossible")
	}

	refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		goto no_ref
	}

	branchesIter, err = repo.Branches()
	if err == nil {
		_ = branchesIter.ForEach(func(branch *plumbing.Reference) error {
			branches = append(branches, branch.Name().Short())
			return nil
		})
	}
	params["branches"] = branches

	logOptions = git.LogOptions{From: refHash} //exhaustruct:ignore
	if commitIter, err = repo.Log(&logOptions); err != nil {
		goto no_ref
	}
	commitIterSeq, params["commits_err"] = commitIterSeqErr(commitIter)
	params["commits"] = iterSeqLimit(commitIterSeq, 3)

	if commitObj, err = repo.CommitObject(refHash); err != nil {
		goto no_ref
	}

	if tree, err = commitObj.Tree(); err != nil {
		goto no_ref
	}

	params["files"] = makeDisplayTree(tree)
	params["readme_filename"], params["readme"] = renderReadmeAtTree(tree)

no_ref:

	params["http_clone_url"] = genHTTPRemoteURL(groupPath, repoName)
	params["ssh_clone_url"] = genSSHRemoteURL(groupPath, repoName)
	params["notes"] = notes

	renderTemplate(w, "repo_index", params)
}
