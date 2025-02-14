package main

import (
	"net/http"
)

// TODO: I probably shouldn't include *all* commits here...
func handle_repo_log(w http.ResponseWriter, r *http.Request, params map[string]any) {
	repo, description, err := open_git_repo(r.Context(), params["group_name"].(string), params["repo_name"].(string))
	if err != nil {
		http.Error(w, "Error opening repo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["repo_description"] = description

	ref_hash, err := get_ref_hash_from_type_and_name(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	commits, err := get_recent_commits(repo, ref_hash, -1)
	if err != nil {
		http.Error(w, "Error getting recent commits: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["commits"] = commits

	render_template(w, "repo_log", params)
}
