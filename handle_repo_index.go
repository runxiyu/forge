package main

import (
	"bytes"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"html/template"
	"net/http"
	"path/filepath"
)

func handle_repo_index(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	project_name, repo_name := r.PathValue("project_name"), r.PathValue("repo_name")
	data["project_name"], data["repo_name"] = project_name, repo_name
	repo, err := git.PlainOpen(filepath.Join(config.Git.Root, project_name, repo_name+".git"))
	if err != nil {
		w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	head, err := repo.Head()
	if err != nil {
		w.Write([]byte("Error getting repo HEAD: " + err.Error()))
		return
	}
	head_hash := head.Hash()
	commit_iter, err := repo.Log(&git.LogOptions{From: head_hash})
	if err != nil {
		w.Write([]byte("Error getting repo commits: " + err.Error()))
		return
	}
	recent_commits := make([]*object.Commit, 0)
	defer commit_iter.Close()
	for range 3 {
		this_recent_commit, err := commit_iter.Next()
		if err != nil {
			w.Write([]byte("Error getting a recent commit: " + err.Error()))
			return
		}
		recent_commits = append(recent_commits, this_recent_commit)
	}
	data["commits"] = recent_commits
	commit_object, err := repo.CommitObject(head_hash)
	if err != nil {
		w.Write([]byte("Error getting commit object: " + err.Error()))
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		w.Write([]byte("Error getting file tree: " + err.Error()))
		return
	}
	readme_file, err := tree.File("README.md")
	if err != nil {
		w.Write([]byte("Error getting file: " + err.Error()))
		return
	}
	readme_file_contents, err := readme_file.Contents()
	var readme_rendered_unsafe bytes.Buffer
	err = goldmark.Convert([]byte(readme_file_contents), &readme_rendered_unsafe)
	if err != nil {
		readme_rendered_unsafe.WriteString("Unable to render README: " + err.Error())
		return
	}
	readme_rendered_safe := template.HTML(bluemonday.UGCPolicy().SanitizeBytes(readme_rendered_unsafe.Bytes()))
	data["readme"] = readme_rendered_safe

	err = templates.ExecuteTemplate(w, "repo_index", data)
	if err != nil {
		w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
