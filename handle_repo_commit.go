package main

import (
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"go.lindenii.runxiyu.org/lindenii-common/misc"
)

type usable_file_patch struct {
	From   diff.File
	To     diff.File
	Chunks []diff.Chunk
}

func handle_repo_commit(w http.ResponseWriter, r *http.Request, params map[string]string) {
	data := make(map[string]any)
	data["global"] = global_data
	group_name, repo_name, commit_id_specified_string := params["group_name"], params["repo_name"], params["commit_id"]
	data["group_name"], data["repo_name"] = group_name, repo_name
	repo, err := open_git_repo(r.Context(), group_name, repo_name)
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

	parent_commit_hash, patch, err := get_patch_from_commit(commit_object)
	if err != nil {
		_, _ = w.Write([]byte("Error getting patch from commit: " + err.Error()))
		return
	}
	data["parent_commit_hash"] = parent_commit_hash.String()
	data["patch"] = patch

	// TODO: Remove unnecessary context
	// TODO: Prepend "+"/"-"/" " instead of solely distinguishing based on color
	usable_file_patches := make([]usable_file_patch, 0)
	for _, file_patch := range patch.FilePatches() {
		from, to := file_patch.Files()
		if from == nil {
			from = fake_diff_file_null
		}
		if to == nil {
			to = fake_diff_file_null
		}
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

type fake_diff_file struct {
	hash plumbing.Hash
	mode filemode.FileMode
	path string
}

func (f fake_diff_file) Hash() plumbing.Hash {
	return f.hash
}

func (f fake_diff_file) Mode() filemode.FileMode {
	return f.mode
}

func (f fake_diff_file) Path() string {
	return f.path
}

var fake_diff_file_null = fake_diff_file{
	hash: plumbing.NewHash("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"),
	mode: misc.First_or_panic(filemode.New("100644")),
	path: "",
}
