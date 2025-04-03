// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

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

// usableFilePatch is a [diff.FilePatch] that is structured in a way more
// friendly for use in HTML templates.
type usableFilePatch struct {
	From   diff.File
	To     diff.File
	Chunks []usableChunk
}

// usableChunk is a [diff.Chunk] that is structured in a way more friendly for
// use in HTML templates.
type usableChunk struct {
	Operation diff.Operation
	Content   string
}

func httpHandleRepoCommit(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var repo *git.Repository
	var commitIDStrSpec, commitIDStrSpecNoSuffix string
	var commitID plumbing.Hash
	var parentCommitHash plumbing.Hash
	var commitObj *object.Commit
	var commitIDStr string
	var err error
	var patch *object.Patch

	repo, commitIDStrSpec = params["repo"].(*git.Repository), params["commit_id"].(string)

	commitIDStrSpecNoSuffix = strings.TrimSuffix(commitIDStrSpec, ".patch")
	commitID = plumbing.NewHash(commitIDStrSpecNoSuffix)
	if commitObj, err = repo.CommitObject(commitID); err != nil {
		errorPage500(writer, params, "Error getting commit object: "+err.Error())
		return
	}
	if commitIDStrSpecNoSuffix != commitIDStrSpec {
		var patchStr string
		if patchStr, err = fmtCommitPatch(commitObj); err != nil {
			errorPage500(writer, params, "Error formatting patch: "+err.Error())
			return
		}
		fmt.Fprintln(writer, patchStr)
		return
	}
	commitIDStr = commitObj.Hash.String()

	if commitIDStr != commitIDStrSpec {
		http.Redirect(writer, request, commitIDStr, http.StatusSeeOther)
		return
	}

	params["commit_object"] = commitObj
	params["commit_id"] = commitIDStr

	parentCommitHash, patch, err = commitToPatch(commitObj)
	if err != nil {
		errorPage500(writer, params, "Error getting patch from commit: "+err.Error())
		return
	}
	params["parent_commit_hash"] = parentCommitHash.String()
	params["patch"] = patch

	params["file_patches"] = makeUsableFilePatches(patch)

	renderTemplate(writer, "repo_commit", params)
}

type fakeDiffFile struct {
	hash plumbing.Hash
	mode filemode.FileMode
	path string
}

func (f fakeDiffFile) Hash() plumbing.Hash {
	return f.hash
}

func (f fakeDiffFile) Mode() filemode.FileMode {
	return f.mode
}

func (f fakeDiffFile) Path() string {
	return f.path
}

var nullFakeDiffFile = fakeDiffFile{
	hash: plumbing.NewHash("0000000000000000000000000000000000000000"),
	mode: misc.FirstOrPanic(filemode.New("100644")),
	path: "",
}

func makeUsableFilePatches(patch diff.Patch) (usableFilePatches []usableFilePatch) {
	// TODO: Remove unnecessary context
	// TODO: Prepend "+"/"-"/" " instead of solely distinguishing based on color

	for _, filePatch := range patch.FilePatches() {
		var fromFile, toFile diff.File
		var ufp usableFilePatch
		chunks := []usableChunk{}

		fromFile, toFile = filePatch.Files()
		if fromFile == nil {
			fromFile = nullFakeDiffFile
		}
		if toFile == nil {
			toFile = nullFakeDiffFile
		}
		for _, chunk := range filePatch.Chunks() {
			var content string

			content = chunk.Content()
			if len(content) > 0 && content[0] == '\n' {
				content = "\n" + content
			} // Horrible hack to fix how browsers newlines that immediately proceed <pre>
			chunks = append(chunks, usableChunk{
				Operation: chunk.Type(),
				Content:   content,
			})
		}
		ufp = usableFilePatch{
			Chunks: chunks,
			From:   fromFile,
			To:     toFile,
		}
		usableFilePatches = append(usableFilePatches, ufp)
	}
	return
}
