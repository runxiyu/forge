// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package oldgit

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

// FmtCommitPatch formats a commit object as if it was returned by
// git-format-patch.
func FmtCommitPatch(commit *object.Commit) (final string, err error) {
	var patch *object.Patch
	var buf bytes.Buffer
	var author object.Signature
	var date string
	var commitTitle, commitDetails string

	if _, patch, err = CommitToPatch(commit); err != nil {
		return "", err
	}

	author = commit.Author
	date = author.When.Format(time.RFC1123Z)

	commitTitle, commitDetails, _ = strings.Cut(commit.Message, "\n")

	// This date is hardcoded in Git.
	fmt.Fprintf(&buf, "From %s Mon Sep 17 00:00:00 2001\n", commit.Hash)
	fmt.Fprintf(&buf, "From: %s <%s>\n", author.Name, author.Email)
	fmt.Fprintf(&buf, "Date: %s\n", date)
	fmt.Fprintf(&buf, "Subject: [PATCH] %s\n\n", commitTitle)

	if commitDetails != "" {
		commitDetails1, commitDetails2, _ := strings.Cut(commitDetails, "\n")
		if strings.TrimSpace(commitDetails1) == "" {
			commitDetails = commitDetails2
		}
		buf.WriteString(commitDetails)
		buf.WriteString("\n")
	}
	buf.WriteString("---\n")
	fmt.Fprint(&buf, patch.Stats().String())
	fmt.Fprintln(&buf)

	buf.WriteString(patch.String())

	fmt.Fprintf(&buf, "\n-- \n2.48.1\n")

	return buf.String(), nil
}
