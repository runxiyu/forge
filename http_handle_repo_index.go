// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"strings"

	"go.lindenii.runxiyu.org/forge/git2c"
	"go.lindenii.runxiyu.org/forge/render"
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

	client, err := git2c.NewClient(config.Git.Socket)
	if err != nil {
		errorPage500(w, params, err.Error())
		return
	}
	defer client.Close()

	commits, readme, err := client.Cmd1(repoPath)
	if err != nil {
		errorPage500(w, params, err.Error())
		return
	}

	params["commits"] = commits
	params["readme_filename"] = readme.Filename
	_, params["readme"] = render.Readme(readme.Content, readme.Filename)
	params["notes"] = notes

	renderTemplate(w, "repo_index", params)

	// TODO: Caching
}
