package main

import (
	"net/http"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func handle_repo_contrib_one(w http.ResponseWriter, r *http.Request, params map[string]any) {
	mr_id_string := params["mr_id"].(string)
	mr_id, err := strconv.ParseInt(mr_id_string, 10, strconv.IntSize)
	if err != nil {
		http.Error(w, "Merge request ID not an integer: "+err.Error(), http.StatusBadRequest)
		return
	}

	var title, status, source_ref, destination_branch string
	err = database.QueryRow(r.Context(), "SELECT COALESCE(title, ''), status, source_ref, COALESCE(destination_branch, '') FROM merge_requests WHERE id = $1", mr_id).Scan(&title, &status, &source_ref, &destination_branch)
	if err != nil {
		http.Error(w, "Error querying merge request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	repo := params["repo"].(*git.Repository)

	source_ref_hash, err := get_ref_hash_from_type_and_name(repo, "branch", source_ref)
	if err != nil {
		http.Error(w, "Error getting source ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}
	source_commit, err := repo.CommitObject(source_ref_hash)
	if err != nil {
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
		if err != nil {
			http.Error(w, "Error getting destination branch hash: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	destination_commit, err := repo.CommitObject(destination_branch_hash)
	if err != nil {
		http.Error(w, "Error getting destination commit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["destination_commit"] = destination_commit

	patch, err := destination_commit.Patch(source_commit)
	if err != nil {
		http.Error(w, "Error getting patch: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["file_patches"] = make_usable_file_patches(patch)

	params["mr_title"], params["mr_status"], params["mr_source_ref"], params["mr_destination_branch"] = title, status, source_ref, destination_branch

	render_template(w, "repo_contrib_one", params)
}
