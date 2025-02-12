package main

import (
	"net/http"
	"net/url"
)

func handle_repo_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	group_name, repo_name := params["group_name"].(string), params["repo_name"].(string)
	repo, err := open_git_repo(r.Context(), group_name, repo_name)
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	head, err := repo.Head()
	if err != nil {
		_, _ = w.Write([]byte("Error getting repo HEAD: " + err.Error()))
		return
	}
	params["ref"] = head.Name().Short()
	head_hash := head.Hash()
	recent_commits, err := get_recent_commits(repo, head_hash, 3)
	if err != nil {
		_, _ = w.Write([]byte("Error getting recent commits: " + err.Error()))
		return
	}
	params["commits"] = recent_commits
	commit_object, err := repo.CommitObject(head_hash)
	if err != nil {
		_, _ = w.Write([]byte("Error getting commit object: " + err.Error()))
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		_, _ = w.Write([]byte("Error getting file tree: " + err.Error()))
		return
	}

	params["readme_filename"], params["readme"] = render_readme_at_tree(tree)
	params["files"] = build_display_git_tree(tree)

	params["clone_url"] = "ssh://" + r.Host + "/" + url.PathEscape(group_name) + "/:/repos/" + url.PathEscape(repo_name)

	err = templates.ExecuteTemplate(w, "repo_index", params)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
