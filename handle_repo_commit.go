package main

import (
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
)

func handle_repo_commit(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	group_name, repo_name, commit_id_string := r.PathValue("group_name"), r.PathValue("repo_name"), r.PathValue("commit_id")
	data["group_name"], data["repo_name"], data["commit_id"] = group_name, repo_name, commit_id_string
	repo, err := open_git_repo(group_name, repo_name)
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	commit_id := plumbing.NewHash(commit_id_string)
	commit_object, err := repo.CommitObject(commit_id)
	if err != nil {
		_, _ = w.Write([]byte("Error getting commit object: " + err.Error()))
		return
	}
	data["commit_object"] = commit_object

	parent_commit_object, err := commit_object.Parent(0)
	if err != nil {
		_, _ = w.Write([]byte("Error getting parent commit object: " + err.Error()))
		return
	}
	data["parent_commit_object"] = parent_commit_object

	patch, err := parent_commit_object.Patch(commit_object)
	if err != nil {
		_, _ = w.Write([]byte("Error getting patch of commit: " + err.Error()))
		return
	}
	data["patch"] = patch

	/*
		for _, file_patch := range patch.FilePatches() {
			for _, chunk := range file_patch.Chunks() {
			}
		}
	*/

	err = templates.ExecuteTemplate(w, "repo_commit", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
