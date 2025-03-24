// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

func httpHandleRepoIndex(writer http.ResponseWriter, _ *http.Request, params map[string]any) {
	var repo *git.Repository
	var repoName string
	var groupPath []string
	var refHash plumbing.Hash
	var refHashSlice []byte
	var err error
	var commitObj *object.Commit
	var tree *object.Tree
	var notes []string
	var branches []string
	var branchesIter storer.ReferenceIter
	var commits []commitDisplay

	repo, repoName, groupPath = params["repo"].(*git.Repository), params["repo_name"].(string), params["group_path"].([]string)

	if strings.Contains(repoName, "\n") || sliceContainsNewlines(groupPath) {
		notes = append(notes, "Path contains newlines; HTTP Git access impossible")
	}

	refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		goto no_ref
	}
	refHashSlice = refHash[:]

	branchesIter, err = repo.Branches()
	if err == nil {
		_ = branchesIter.ForEach(func(branch *plumbing.Reference) error {
			branches = append(branches, branch.Name().Short())
			return nil
		})
	}
	params["branches"] = branches

	if value, found := indexCommitsDisplayCache.Get(refHashSlice); found {
		if value != nil {
			commits = value
		} else {
			goto readme
		}
	} else {
		start := time.Now()
		commits, err = getRecentCommitsDisplay(repo, refHash, 5)
		if err != nil {
			commits = nil
		}
		cost := time.Since(start).Nanoseconds()
		indexCommitsDisplayCache.Set(refHashSlice, commits, cost)
		if err != nil {
			goto readme
		}
	}

	params["commits"] = commits

readme:

	if value, found := treeReadmeCache.Get(refHashSlice); found {
		params["files"] = value.DisplayTree
		params["readme_filename"] = value.ReadmeFilename
		params["readme"] = value.ReadmeRendered
	} else {
		start := time.Now()
		if commitObj, err = repo.CommitObject(refHash); err != nil {
			goto no_ref
		}

		if tree, err = commitObj.Tree(); err != nil {
			goto no_ref
		}
		displayTree := makeDisplayTree(tree)
		readmeFilename, readmeRendered := renderReadmeAtTree(tree)
		cost := time.Since(start).Nanoseconds()

		params["files"] = displayTree
		params["readme_filename"] = readmeFilename
		params["readme"] = readmeRendered

		entry := treeReadmeCacheEntry{
			DisplayTree:    displayTree,
			ReadmeFilename: readmeFilename,
			ReadmeRendered: readmeRendered,
		}
		treeReadmeCache.Set(refHashSlice, entry, cost)
	}

no_ref:

	params["http_clone_url"] = genHTTPRemoteURL(groupPath, repoName)
	params["ssh_clone_url"] = genSSHRemoteURL(groupPath, repoName)
	params["notes"] = notes

	renderTemplate(writer, "repo_index", params)
}
