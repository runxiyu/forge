// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"net/http"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.lindenii.runxiyu.org/forge/forged/internal/web"
)

// httpHandleRepoContribOne provides an interface to each merge request of a
// repo.
func (s *Server) httpHandleRepoContribOne(writer http.ResponseWriter, request *http.Request, params map[string]any) {
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
		web.ErrorPage400(s.templates, writer, params, "Merge request ID not an integer")
		return
	}
	mrIDInt = int(mrIDInt64)

	if err = s.database.QueryRow(request.Context(),
		"SELECT COALESCE(title, ''), status, source_ref, COALESCE(destination_branch, '') FROM merge_requests WHERE repo_id = $1 AND repo_local_id = $2",
		params["repo_id"], mrIDInt,
	).Scan(&title, &status, &srcRefStr, &dstBranchStr); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error querying merge request: "+err.Error())
		return
	}

	repo = params["repo"].(*git.Repository)

	if srcRefHash, err = getRefHash(repo, "branch", srcRefStr); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting source ref hash: "+err.Error())
		return
	}
	if srcCommit, err = repo.CommitObject(srcRefHash); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting source commit: "+err.Error())
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
		web.ErrorPage500(s.templates, writer, params, "Error getting destination branch hash: "+err.Error())
		return
	}

	if dstCommit, err = repo.CommitObject(dstBranchHash); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting destination commit: "+err.Error())
		return
	}
	params["destination_commit"] = dstCommit

	if mergeBases, err = srcCommit.MergeBase(dstCommit); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting merge base: "+err.Error())
		return
	}

	if len(mergeBases) < 1 {
		web.ErrorPage500(s.templates, writer, params, "No merge base found for this merge request; these two branches do not share any common history")
		// TODO
		return
	}

	mergeBaseCommit = mergeBases[0]
	params["merge_base"] = mergeBaseCommit

	patch, err := mergeBaseCommit.Patch(srcCommit)
	if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting patch: "+err.Error())
		return
	}
	params["file_patches"] = makeUsableFilePatches(patch)

	params["mr_title"], params["mr_status"], params["mr_source_ref"], params["mr_destination_branch"] = title, status, srcRefStr, dstBranchStr

	s.renderTemplate(writer, "repo_contrib_one", params)
}
