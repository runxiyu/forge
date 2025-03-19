// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func httpHandleRepoContribOne(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var mrIDStr string
	var mrIDInt int
	var err error
	var title, status, srcRefStr, dstBranchStr string
	var repo *git.Repository
	var srcRefHash plumbing.Hash
	var dstBranchHash plumbing.Hash
	var srcCommit, dstCommit, mergeBaseCommit *object.Commit
	var mergeBases []*object.Commit

	mrIDStr = params["mr_id"].(string)
	mrIDInt64, err := strconv.ParseInt(mrIDStr, 10, strconv.IntSize)
	if err != nil {
		http.Error(w, "Merge request ID not an integer: "+err.Error(), http.StatusBadRequest)
		return
	}
	mrIDInt = int(mrIDInt64)

	if err = database.QueryRow(r.Context(),
		"SELECT COALESCE(title, ''), status, source_ref, COALESCE(destination_branch, '') FROM merge_requests WHERE id = $1",
		mrIDInt,
	).Scan(&title, &status, &srcRefStr, &dstBranchStr); err != nil {
		http.Error(w, "Error querying merge request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	repo = params["repo"].(*git.Repository)

	if srcRefHash, err = getRefHash(repo, "branch", srcRefStr); err != nil {
		http.Error(w, "Error getting source ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if srcCommit, err = repo.CommitObject(srcRefHash); err != nil {
		http.Error(w, "Error getting source commit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["source_commit"] = srcCommit

	if dstBranchStr == "" {
		dstBranchStr = "HEAD"
		dstBranchHash, err = getRefHash(repo, "", "")
	} else {
		dstBranchHash, err = getRefHash(repo, "branch", dstBranchStr)
	}
	if err != nil {
		http.Error(w, "Error getting destination branch hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if dstCommit, err = repo.CommitObject(dstBranchHash); err != nil {
		http.Error(w, "Error getting destination commit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["destination_commit"] = dstCommit

	if mergeBases, err = srcCommit.MergeBase(dstCommit); err != nil {
		http.Error(w, "Error getting merge base: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mergeBaseCommit = mergeBases[0]
	params["merge_base"] = mergeBaseCommit

	patch, err := mergeBaseCommit.Patch(srcCommit)
	if err != nil {
		http.Error(w, "Error getting patch: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["file_patches"] = makeUsableFilePatches(patch)

	params["mr_title"], params["mr_status"], params["mr_source_ref"], params["mr_destination_branch"] = title, status, srcRefStr, dstBranchStr

	renderTemplate(w, "repo_contrib_one", params)
}
