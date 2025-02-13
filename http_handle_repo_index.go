package main

import (
	"net/http"
	"net/url"
)

func handle_repo_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	group_name, repo_name := params["group_name"].(string), params["repo_name"].(string)
	repo, description, err := open_git_repo(r.Context(), group_name, repo_name)
	if err != nil {
		http.Error(w, "Error opening repo:: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["repo_description"] = description
	head, err := repo.Head()
	if err != nil {
		http.Error(w, "Error getting repo HEAD:: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["ref"] = head.Name().Short()
	head_hash := head.Hash()
	recent_commits, err := get_recent_commits(repo, head_hash, 3)
	if err != nil {
		http.Error(w, "Error getting recent commits:: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["commits"] = recent_commits
	commit_object, err := repo.CommitObject(head_hash)
	if err != nil {
		http.Error(w, "Error getting commit object:: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		http.Error(w, "Error getting file tree:: "+err.Error(), http.StatusInternalServerError)
		return
	}

	params["readme_filename"], params["readme"] = render_readme_at_tree(tree)
	params["files"] = build_display_git_tree(tree)

	params["clone_url"] = "ssh://" + r.Host + "/" + url.PathEscape(group_name) + "/:/repos/" + url.PathEscape(repo_name)

	err = templates.ExecuteTemplate(w, "repo_index", params)
	if err != nil {
		http.Error(w, "Error rendering template:: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
