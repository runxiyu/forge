package main

import (
	"net/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_index(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	category_name, repo_name := r.PathValue("category_name"), r.PathValue("repo_name")
	data["category_name"], data["repo_name"] = category_name, repo_name
	repo, err := open_git_repo(category_name, repo_name)
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
	commit_iter, err := repo.Log(&git.LogOptions{From: head_hash})
	if err != nil {
		_, _ = w.Write([]byte("Error getting repo commits: " + err.Error()))
		return
	}
	recent_commits := make([]*object.Commit, 0)
	defer commit_iter.Close()
	for range 3 {
		this_recent_commit, err := commit_iter.Next()
		if err != nil {
			_, _ = w.Write([]byte("Error getting a recent commit: " + err.Error()))
			return
		}
		recent_commits = append(recent_commits, this_recent_commit)
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
