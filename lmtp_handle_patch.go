// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	// "crypto/rand"
	// "fmt"
	"os"
	"os/exec"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/emersion/go-message"
	"github.com/go-git/go-git/v5"
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
	var fsPath string
	repo, _, _, fsPath, err = openRepo(session.ctx, groupPath, repoName)
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

	// TODO: Try to not shell out

	for _, diffFile := range diffFiles {
		var sourceFile *object.File
		if sourceFile, err = headTree.File(diffFile.OldName); err != nil {
			return err
		}
		var sourceString string
		if sourceString, err = sourceFile.Contents(); err != nil {
			return err
		}
		hashBuf := bytes.Buffer{}
		patchedBuf := bytes.Buffer{}
		sourceBuf := bytes.NewReader(stringToBytes(sourceString))
		if err = gitdiff.Apply(&patchedBuf, sourceBuf, diffFile); err != nil {
			return err
		}
		proc := exec.CommandContext(session.ctx, "git", "hash-object", "-w", "-t", "blob", "--stdin")
		proc.Env = append(os.Environ(), "GIT_DIR="+fsPath)
		proc.Stdout = &hashBuf
		proc.Stdin = &patchedBuf
		if err = proc.Start(); err != nil {
			return err
		}
		if err = proc.Wait(); err != nil {
			return err
		}
		newHash := hashBuf.Bytes()
		if len(newHash) != 20*2+1 { // TODO: Hardcoded from the size of plumbing.Hash
			panic("unexpected hash size")
		}
		// TODO: Add to tree
	}

	// contribBranchName := rand.Text()

	// TODO: Store the branch

	// fmt.Println(repo, diffFiles, preamble)
	_ = preamble

	return nil
}
