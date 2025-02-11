package main

import (
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
)

// TODO: I probably shouldn't include *all* commitsh here...
func handle_repo_log(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	group_name, repo_name, ref_name := r.PathValue("group_name"), r.PathValue("repo_name"), r.PathValue("ref")
	data["group_name"], data["repo_name"], data["ref"] = group_name, repo_name, ref_name
	repo, err := open_git_repo(group_name, repo_name)
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(ref_name), true)
	if err != nil {
		_, _ = w.Write([]byte("Error getting repo reference: " + err.Error()))
		return
	}
	ref_hash := ref.Hash()
	recent_commits, err := get_recent_commits(repo, ref_hash, -1)
	if err != nil {
		_, _ = w.Write([]byte("Error getting recent commits: " + err.Error()))
		return
	}
	data["commits"] = recent_commits
	commit_object, err := repo.CommitObject(ref_hash)
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

	err = templates.ExecuteTemplate(w, "repo_log", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
