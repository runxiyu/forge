package main

import (
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
)

type usable_file_patch struct {
	From   diff.File
	To     diff.File
	Chunks []diff.Chunk
}

func handle_repo_commit(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	group_name, repo_name, commit_id_specified_string := r.PathValue("group_name"), r.PathValue("repo_name"), r.PathValue("commit_id")
	data["group_name"], data["repo_name"] = group_name, repo_name
	repo, err := open_git_repo(group_name, repo_name)
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	commit_id_specified_string_without_suffix := strings.TrimSuffix(commit_id_specified_string, ".patch")
	commit_id := plumbing.NewHash(commit_id_specified_string_without_suffix)
	commit_object, err := repo.CommitObject(commit_id)
	if err != nil {
		_, _ = w.Write([]byte("Error getting commit object: " + err.Error()))
		return
	}
	if commit_id_specified_string_without_suffix != commit_id_specified_string {
		patch, err := format_patch_from_commit(commit_object)
		if err != nil {
			_, _ = w.Write([]byte("Error formatting patch: " + err.Error()))
			return
		}
		_, _ = w.Write([]byte(patch))
		return
	}
	commit_id_string := commit_object.Hash.String()

	if commit_id_string != commit_id_specified_string {
		http.Redirect(w, r, commit_id_string, http.StatusSeeOther)
		return
	}

	data["commit_object"] = commit_object
	data["commit_id"] = commit_id_string

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

	// TODO: Remove unnecessary context
	usable_file_patches := make([]usable_file_patch, 0)
	for _, file_patch := range patch.FilePatches() {
		from, to := file_patch.Files()
		usable_file_patch := usable_file_patch{
			Chunks: file_patch.Chunks(),
			From:   from,
			To:     to,
		}
		usable_file_patches = append(usable_file_patches, usable_file_patch)
	}
	data["file_patches"] = usable_file_patches

	err = templates.ExecuteTemplate(w, "repo_commit", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
