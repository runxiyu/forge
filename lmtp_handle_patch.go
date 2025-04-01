// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/rand"
	"fmt"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/emersion/go-message"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func lmtpHandlePatch(session *lmtpSession, groupPath []string, repoName string, email *message.Entity) (err error) {
	var diffFiles []*gitdiff.File
	var preamble string
	if diffFiles, preamble, err = gitdiff.Parse(email.Body); err != nil {
		return
	}

	var repo *git.Repository
	repo, _, _, err = openRepo(session.ctx, groupPath, repoName)
	if err != nil {
		return
	}

	var headRef *plumbing.Reference
	if headRef, err = repo.Head(); err != nil {
		return
	}

	var headCommit *object.Commit
	if headCommit, err = repo.CommitObject(headRef.Hash()); err != nil {
		return
	}

	var headTree *object.Tree
	if headTree, err = headCommit.Tree(); err != nil {
		return
	}

	// What's left to do: apply the patch on a separate branch.
	// I'm not sure how to do this in go-git. I have a few thoughts:
	// Method 1. Create a copy of the tree object; then iterate through
	//           diffFiles, traversing the tree accordingly, then hash
	//           blobs, insert them into the repo, replace the entry in
	//           the tree, then commit the tree onto a new branch.
	//           Perhaps storer can help with this?
	// Method 2. Create an index, run the patch on it, and commit the
	//           index.
	// I think Method 1 technically suffers from a race if the repo is
	// garbage collected between the time that the blob is created and
	// the time it is committed to a branch. We could just prevent
	// external GCs and lock GCs ourselves though, so it's no big deal.
	// Method 2 is a bit annoying and I have this impression that
	// worktrees/indexes are fragile and bloated.

	myTree := *headTree
	// TODO: Check if it's actually safe to modify myTree. We don't
	// own these slices, so this might interfere with go-git's internal
	// state.

	for _, diffFile := range diffFiles {
		blobObject := plumbing.EncodedObject
		var blobHash plumbing.Hash
		if blobHash, err = repo.Storer.SetEncodedObject(blobObject); err != nil {
			return
		}
	}

	/*
		contribBranchName := rand.Text()

		if err = repo.CreateBranch(&config.Branch{ Name: contribBranchName, }); err != nil {
			return
		}

		var contribRef *plumbing.Reference
		if contribRef, err = repo.Reference(plumbing.NewBranchReferenceName(contribBranchName, true); err != nil {
			return
		}
	*/

	fmt.Println(repo, diffFiles, preamble)

	return nil
}
