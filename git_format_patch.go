// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

// get_patch_from_commit formats a commit object as if it was returned by
// git-format-patch.
func format_patch_from_commit(commit *object.Commit) (string, error) {
	_, patch, err := get_patch_from_commit(commit)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	author := commit.Author
	date := author.When.Format(time.RFC1123Z)

	commit_msg_title, commit_msg_details, _ := strings.Cut(commit.Message, "\n")

	// This date is hardcoded in Git.
	fmt.Fprintf(&buf, "From %s Mon Sep 17 00:00:00 2001\n", commit.Hash)
	fmt.Fprintf(&buf, "From: %s <%s>\n", author.Name, author.Email)
	fmt.Fprintf(&buf, "Date: %s\n", date)
	fmt.Fprintf(&buf, "Subject: [PATCH] %s\n\n", commit_msg_title)

	if commit_msg_details != "" {
		commit_msg_details_first_line, commit_msg_details_rest, _ := strings.Cut(commit_msg_details, "\n")
		if strings.TrimSpace(commit_msg_details_first_line) == "" {
			commit_msg_details = commit_msg_details_rest
		}
		buf.WriteString(commit_msg_details)
		buf.WriteString("\n")
	}
	buf.WriteString("---\n")
	fmt.Fprint(&buf, patch.Stats().String())
	fmt.Fprintln(&buf)

	buf.WriteString(patch.String())

	fmt.Fprintf(&buf, "\n-- \n2.48.1\n")

	return buf.String(), nil
}
