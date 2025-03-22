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

func httpHandleRepoContribOne(writer http.ResponseWriter, request *http.Request, params map[string]any) {
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
		http.Error(writer, "Merge request ID not an integer: "+err.Error(), http.StatusBadRequest)
		return
	}
	mrIDInt = int(mrIDInt64)

	if err = database.QueryRow(request.Context(),
		"SELECT COALESCE(title, ''), status, source_ref, COALESCE(destination_branch, '') FROM merge_requests WHERE id = $1",
		mrIDInt,
	).Scan(&title, &status, &srcRefStr, &dstBranchStr); err != nil {
		errorPage500(writer, params, "Error querying merge request: "+err.Error())
		return
	}

	repo = params["repo"].(*git.Repository)

	if srcRefHash, err = getRefHash(repo, "branch", srcRefStr); err != nil {
		errorPage500(writer, params, "Error getting source ref hash: "+err.Error())
		return
	}
	if srcCommit, err = repo.CommitObject(srcRefHash); err != nil {
		errorPage500(writer, params, "Error getting source commit: "+err.Error())
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
		errorPage500(writer, params, "Error getting destination branch hash: "+err.Error())
		return
	}

	if dstCommit, err = repo.CommitObject(dstBranchHash); err != nil {
		errorPage500(writer, params, "Error getting destination commit: "+err.Error())
		return
	}
	params["destination_commit"] = dstCommit

	if mergeBases, err = srcCommit.MergeBase(dstCommit); err != nil {
		errorPage500(writer, params, "Error getting merge base: "+err.Error())
		return
	}
	mergeBaseCommit = mergeBases[0]
	params["merge_base"] = mergeBaseCommit

	patch, err := mergeBaseCommit.Patch(srcCommit)
	if err != nil {
		errorPage500(writer, params, "Error getting patch: "+err.Error())
		return
	}
	params["file_patches"] = makeUsableFilePatches(patch)

	params["mr_title"], params["mr_status"], params["mr_source_ref"], params["mr_destination_branch"] = title, status, srcRefStr, dstBranchStr

	renderTemplate(writer, "repo_contrib_one", params)
}
