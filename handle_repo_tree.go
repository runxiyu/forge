package main

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

func handle_repo_tree(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	project_name, repo_name, path_spec := r.PathValue("project_name"), r.PathValue("repo_name"), strings.TrimSuffix(r.PathValue("rest"), "/")
	data["project_name"], data["repo_name"], data["path_spec"] = project_name, repo_name, path_spec
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

	target, err := tree.Tree(path_spec)
	if err != nil {
		file, err := tree.File(path_spec)
		if err != nil {
			w.Write([]byte("Error retrieving path: " + err.Error()))
			return
		}
		data["file_contents"], err = file.Contents()
		if err != nil {
			w.Write([]byte("Error reading file: " + err.Error()))
			return
		}
		err = templates.ExecuteTemplate(w, "repo_tree_file", data)
		if err != nil {
			w.Write([]byte("Error rendering template: " + err.Error()))
			return
		}
		return
	}

	readme_file, err := target.File("README.md")
	if err != nil {
		data["readme"] = "There is no README for this directory."
	} else {
		readme_file_contents, err := readme_file.Contents()
		var readme_rendered_unsafe bytes.Buffer
		err = goldmark.Convert([]byte(readme_file_contents), &readme_rendered_unsafe)
		if err != nil {
			readme_rendered_unsafe.WriteString("Unable to render README: " + err.Error())
		}
		readme_rendered_safe := template.HTML(bluemonday.UGCPolicy().SanitizeBytes(readme_rendered_unsafe.Bytes()))
		data["readme"] = readme_rendered_safe
	}

	display_git_tree := make([]display_git_tree_entry_t, 0)
	for _, entry := range target.Entries {
		display_git_tree_entry := display_git_tree_entry_t{}
		os_mode, err := entry.Mode.ToOSFileMode()
		if err != nil {
			display_git_tree_entry.Mode = "----"
		} else {
			display_git_tree_entry.Mode = os_mode.String()[:4]
		}
		display_git_tree_entry.Is_file = entry.Mode.IsFile()
		display_git_tree_entry.Size, err = target.Size(entry.Name)
		if err != nil {
			display_git_tree_entry.Size = 0
		}
		display_git_tree_entry.Name = entry.Name
		display_git_tree = append(display_git_tree, display_git_tree_entry)
	}
	data["files"] = display_git_tree

	err = templates.ExecuteTemplate(w, "repo_tree_dir", data)
	if err != nil {
		w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
