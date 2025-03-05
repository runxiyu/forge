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

func handle_repo_contrib_one(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var mr_id_string string
	var mr_id int
	var err error
	var title, status, source_ref, destination_branch string
	var repo *git.Repository
	var source_ref_hash plumbing.Hash
	var source_commit, destination_commit, merge_base *object.Commit
	var merge_bases []*object.Commit

	mr_id_string = params["mr_id"].(string)
	mr_id_int64, err := strconv.ParseInt(mr_id_string, 10, strconv.IntSize)
	if err != nil {
		http.Error(w, "Merge request ID not an integer: "+err.Error(), http.StatusBadRequest)
		return
	}
	mr_id = int(mr_id_int64)

	if err = database.QueryRow(r.Context(),
		"SELECT COALESCE(title, ''), status, source_ref, COALESCE(destination_branch, '') FROM merge_requests WHERE id = $1",
		mr_id,
	).Scan(&title, &status, &source_ref, &destination_branch); err != nil {
		http.Error(w, "Error querying merge request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	repo = params["repo"].(*git.Repository)

	if source_ref_hash, err = get_ref_hash_from_type_and_name(repo, "branch", source_ref); err != nil {
		http.Error(w, "Error getting source ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if source_commit, err = repo.CommitObject(source_ref_hash); err != nil {
		http.Error(w, "Error getting source commit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["source_commit"] = source_commit

	var destination_branch_hash plumbing.Hash
	if destination_branch == "" {
		destination_branch = "HEAD"
		destination_branch_hash, err = get_ref_hash_from_type_and_name(repo, "", "")
	} else {
		destination_branch_hash, err = get_ref_hash_from_type_and_name(repo, "branch", destination_branch)
	}
	if err != nil {
		http.Error(w, "Error getting destination branch hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if destination_commit, err = repo.CommitObject(destination_branch_hash); err != nil {
		http.Error(w, "Error getting destination commit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["destination_commit"] = destination_commit

	if merge_bases, err = source_commit.MergeBase(destination_commit); err != nil {
		http.Error(w, "Error getting merge base: "+err.Error(), http.StatusInternalServerError)
		return
	}
	merge_base = merge_bases[0]
	params["merge_base"] = merge_base

	patch, err := merge_base.Patch(source_commit)
	if err != nil {
		http.Error(w, "Error getting patch: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["file_patches"] = make_usable_file_patches(patch)

	params["mr_title"], params["mr_status"], params["mr_source_ref"], params["mr_destination_branch"] = title, status, source_ref, destination_branch

	render_template(w, "repo_contrib_one", params)
}
