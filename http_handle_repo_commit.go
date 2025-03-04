// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.lindenii.runxiyu.org/lindenii-common/misc"
)

// The file patch type from go-git isn't really usable in HTML templates
// either.
type usable_file_patch_t struct {
	From   diff.File
	To     diff.File
	Chunks []usable_chunk
}

type usable_chunk struct {
	Operation diff.Operation
	Content   string
}

func handle_repo_commit(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var repo *git.Repository
	var commit_id_specified_string, commit_id_specified_string_without_suffix string
	var commit_id plumbing.Hash
	var parent_commit_hash plumbing.Hash
	var commit_object *object.Commit
	var commit_id_string string
	var err error
	var patch *object.Patch

	repo, commit_id_specified_string = params["repo"].(*git.Repository), params["commit_id"].(string)

	commit_id_specified_string_without_suffix = strings.TrimSuffix(commit_id_specified_string, ".patch")
	commit_id = plumbing.NewHash(commit_id_specified_string_without_suffix)
	if commit_object, err = repo.CommitObject(commit_id); err != nil {
		http.Error(w, "Error getting commit object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if commit_id_specified_string_without_suffix != commit_id_specified_string {
		var formatted_patch string
		if formatted_patch, err = format_patch_from_commit(commit_object); err != nil {
			http.Error(w, "Error formatting patch: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, formatted_patch)
		return
	}
	commit_id_string = commit_object.Hash.String()

	if commit_id_string != commit_id_specified_string {
		http.Redirect(w, r, commit_id_string, http.StatusSeeOther)
		return
	}

	params["commit_object"] = commit_object
	params["commit_id"] = commit_id_string

	parent_commit_hash, patch, err = get_patch_from_commit(commit_object)
	if err != nil {
		http.Error(w, "Error getting patch from commit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	params["parent_commit_hash"] = parent_commit_hash.String()
	params["patch"] = patch

	params["file_patches"] = make_usable_file_patches(patch)

	render_template(w, "repo_commit", params)
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
	hash: plumbing.NewHash("0000000000000000000000000000000000000000"),
	mode: misc.First_or_panic(filemode.New("100644")),
	path: "",
}

func make_usable_file_patches(patch diff.Patch) (usable_file_patches []usable_file_patch_t) {
	// TODO: Remove unnecessary context
	// TODO: Prepend "+"/"-"/" " instead of solely distinguishing based on color

	for _, file_patch := range patch.FilePatches() {
		var from, to diff.File
		var usable_file_patch usable_file_patch_t
		chunks := []usable_chunk{}

		from, to = file_patch.Files()
		if from == nil {
			from = fake_diff_file_null
		}
		if to == nil {
			to = fake_diff_file_null
		}
		for _, chunk := range file_patch.Chunks() {
			var content string

			content = chunk.Content()
			if len(content) > 0 && content[0] == '\n' {
				content = "\n" + content
			} // Horrible hack to fix how browsers newlines that immediately proceed <pre>
			chunks = append(chunks, usable_chunk{
				Operation: chunk.Type(),
				Content:   content,
			})
		}
		usable_file_patch = usable_file_patch_t{
			Chunks: chunks,
			From:   from,
			To:     to,
		}
		usable_file_patches = append(usable_file_patches, usable_file_patch)
	}
	return
}
