// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/go-git/go-git/v5"
	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

func (s *Server) lmtpHandlePatch(session *lmtpSession, groupPath []string, repoName string, mbox io.Reader) (err error) {
	var diffFiles []*gitdiff.File
	var preamble string
	if diffFiles, preamble, err = gitdiff.Parse(mbox); err != nil {
		return fmt.Errorf("failed to parse patch: %w", err)
	}

	var header *gitdiff.PatchHeader
	if header, err = gitdiff.ParsePatchHeader(preamble); err != nil {
		return fmt.Errorf("failed to parse patch headers: %w", err)
	}

	var repo *git.Repository
	var fsPath string
	repo, _, _, fsPath, err = s.openRepo(session.ctx, groupPath, repoName)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get repo head hash: %w", err)
	}
	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get repo head commit: %w", err)
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get repo head tree: %w", err)
	}

	headTreeHash := headTree.Hash.String()

	blobUpdates := make(map[string][]byte)
	for _, diffFile := range diffFiles {
		sourceFile, err := headTree.File(diffFile.OldName)
		if err != nil {
			return fmt.Errorf("failed to get file at old name %#v: %w", diffFile.OldName, err)
		}
		sourceString, err := sourceFile.Contents()
		if err != nil {
			return fmt.Errorf("failed to get contents: %w", err)
		}

		sourceBuf := bytes.NewReader(misc.StringToBytes(sourceString))
		var patchedBuf bytes.Buffer
		if err := gitdiff.Apply(&patchedBuf, sourceBuf, diffFile); err != nil {
			return fmt.Errorf("failed to apply patch: %w", err)
		}

		var hashBuf bytes.Buffer

		// It's really difficult to do this via go-git so we're just
		// going to use upstream git for now.
		// TODO
		cmd := exec.CommandContext(session.ctx, "git", "hash-object", "-w", "-t", "blob", "--stdin")
		cmd.Env = append(os.Environ(), "GIT_DIR="+fsPath)
		cmd.Stdout = &hashBuf
		cmd.Stdin = &patchedBuf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run git hash-object: %w", err)
		}

		newHashStr := strings.TrimSpace(hashBuf.String())
		newHash, err := hex.DecodeString(newHashStr)
		if err != nil {
			return fmt.Errorf("failed to decode hex string from git: %w", err)
		}

		blobUpdates[diffFile.NewName] = newHash
		if diffFile.NewName != diffFile.OldName {
			blobUpdates[diffFile.OldName] = nil // Mark for deletion.
		}
	}

	newTreeSha, err := buildTreeRecursive(session.ctx, fsPath, headTreeHash, blobUpdates)
	if err != nil {
		return fmt.Errorf("failed to recursively build a tree: %w", err)
	}

	commitMsg := header.Title
	if header.Body != "" {
		commitMsg += "\n\n" + header.Body
	}

	env := append(os.Environ(),
		"GIT_DIR="+fsPath,
		"GIT_AUTHOR_NAME="+header.Author.Name,
		"GIT_AUTHOR_EMAIL="+header.Author.Email,
		"GIT_AUTHOR_DATE="+header.AuthorDate.Format(time.RFC3339),
	)
	commitCmd := exec.CommandContext(session.ctx, "git", "commit-tree", newTreeSha, "-p", headCommit.Hash.String(), "-m", commitMsg)
	commitCmd.Env = env

	var commitOut bytes.Buffer
	commitCmd.Stdout = &commitOut
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit tree: %w", err)
	}
	newCommitSha := strings.TrimSpace(commitOut.String())

	newBranchName := rand.Text()

	refCmd := exec.CommandContext(session.ctx, "git", "update-ref", "refs/heads/contrib/"+newBranchName, newCommitSha) //#nosec G204
	refCmd.Env = append(os.Environ(), "GIT_DIR="+fsPath)
	if err := refCmd.Run(); err != nil {
		return fmt.Errorf("failed to update ref: %w", err)
	}

	return nil
}
