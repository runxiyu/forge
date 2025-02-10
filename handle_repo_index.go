package main

import (
	"net/http"
)

func handle_repo_index(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	group_name, repo_name := r.PathValue("group_name"), r.PathValue("repo_name")
	data["group_name"], data["repo_name"] = group_name, repo_name
	repo, err := open_git_repo(group_name, repo_name)
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	head, err := repo.Head()
	if err != nil {
		_, _ = w.Write([]byte("Error getting repo HEAD: " + err.Error()))
		return
	}
	data["ref"] = head.Name().Short()
	head_hash := head.Hash()
	recent_commits, err := get_recent_commits(repo, head_hash)
	if err != nil {
		_, _ = w.Write([]byte("Error getting recent commits: " + err.Error()))
		return
	}
	data["commits"] = recent_commits
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

	data["readme"] = render_readme_at_tree(tree)
	data["files"] = build_display_git_tree(tree)

	err = templates.ExecuteTemplate(w, "repo_index", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
