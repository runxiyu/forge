package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"go.lindenii.runxiyu.org/lindenii-common/misc"
)

var err_get_patch = errors.New("Failed to get patch from commit")

func format_patch_from_commit(commit *object.Commit) (string, error) {
	parent, err := commit.Parent(0)
	if err != nil {
		return "", err
	}

	var patch *object.Patch
	patch, err = parent.Patch(commit)
	if err != nil {
		return "", misc.Wrap_one_error(err_get_patch, err)
	}

	var buf bytes.Buffer

	author := commit.Author
	date := author.When.Format(time.RFC1123Z)

	commit_msg_title, commit_msg_details, _ := strings.Cut(commit.Message, "\n")

	fmt.Fprintf(&buf, "From %s Mon Sep 17 00:00:00 2001\n", commit.Hash)
	fmt.Fprintf(&buf, "From: %s <%s>\n", author.Name, author.Email)
	fmt.Fprintf(&buf, "Date: %s\n", date)
	fmt.Fprintf(&buf, "Subject: [PATCH] %s\n\n", commit_msg_title)

	if commit_msg_details != "" {
		fmt.Println("fdsafsad")
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
