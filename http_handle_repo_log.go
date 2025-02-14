package main

import (
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
)

// TODO: I probably shouldn't include *all* commits here...
func handle_repo_log(w http.ResponseWriter, r *http.Request, params map[string]any) {
	group_name, repo_name, ref_name := params["group_name"].(string), params["repo_name"].(string), params["ref_name"].(string)
	repo, description, err := open_git_repo(r.Context(), group_name, repo_name)
	if err != nil {
		http.Error(w, "Error opening repo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["repo_description"] = description
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(ref_name), true)
	if err != nil {
		http.Error(w, "Error getting repo reference: "+err.Error(), http.StatusInternalServerError)
		return
	}
	ref_hash := ref.Hash()
	commits, err := get_recent_commits(repo, ref_hash, -1)
	if err != nil {
		http.Error(w, "Error getting recent commits: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["commits"] = commits

	render_template(w, "repo_log", params)
	return
}
