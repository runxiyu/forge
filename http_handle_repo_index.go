// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"git.sr.ht/~sircmpwn/go-bare"
)

type commitDisplay struct {
	Hash    string
	Author  string
	Email   string
	Date    string
	Message string
}

// httpHandleRepoIndex provides the front page of a repo using git2d.
func httpHandleRepoIndex(w http.ResponseWriter, req *http.Request, params map[string]any) {
	repoName := params["repo_name"].(string)
	groupPath := params["group_path"].([]string)

	_, repoPath, _, _, _, _, _ := getRepoInfo(req.Context(), groupPath, repoName, "") // TODO: Don't use getRepoInfo

	var notes []string
	if strings.Contains(repoName, "\n") || sliceContainsNewlines(groupPath) {
		notes = append(notes, "Path contains newlines; HTTP Git access impossible")
	}

	conn, err := net.Dial("unix", config.Git.Socket)
	if err != nil {
		errorPage500(w, params, "git2d connection failed: "+err.Error())
		return
	}
	defer conn.Close()

	writer := bare.NewWriter(conn)
	reader := bare.NewReader(conn)

	if err := writer.WriteData(stringToBytes(repoPath)); err != nil {
		errorPage500(w, params, "sending repo path failed: "+err.Error())
		return
	}

	if err := writer.WriteUint(1); err != nil {
		errorPage500(w, params, "sending command failed: "+err.Error())
		return
	}

	status, err := reader.ReadUint()
	if err != nil {
		errorPage500(w, params, "reading status failed: "+err.Error())
		return
	}
	if status != 0 {
		errorPage500(w, params, fmt.Sprintf("git2d error: %d", status))
		return
	}

	// README
	readmeRaw, err := reader.ReadData()
	if err != nil {
		readmeRaw = nil
	}
	readmeFilename, readmeRendered := renderReadme(readmeRaw, "README.md")

	// Commits
	var commits []commitDisplay
	for {
		id, err := reader.ReadData()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			errorPage500(w, params, "error reading commit ID: "+err.Error())
			return
		}

		title, _ := reader.ReadData()
		authorName, _ := reader.ReadData()
		authorEmail, _ := reader.ReadData()
		authorDate, _ := reader.ReadData()

		commits = append(commits, commitDisplay{
			Hash:    hex.EncodeToString(id),
			Author:  bytesToString(authorName),
			Email:   bytesToString(authorEmail),
			Date:    bytesToString(authorDate),
			Message: bytesToString(title),
		})
	}

	params["commits"] = commits
	params["readme_filename"] = readmeFilename
	params["readme"] = readmeRendered
	params["notes"] = notes

	renderTemplate(w, "repo_index", params)

	// TODO: Caching
}
