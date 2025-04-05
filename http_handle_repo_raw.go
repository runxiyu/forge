// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"go.lindenii.runxiyu.org/forge/git2c"
	"go.lindenii.runxiyu.org/forge/misc"
)

// httpHandleRepoRaw serves raw files, or directory listings that point to raw
// files.
func httpHandleRepoRaw(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	repoName := params["repo_name"].(string)
	groupPath := params["group_path"].([]string)
	rawPathSpec := params["rest"].(string)
	pathSpec := strings.TrimSuffix(rawPathSpec, "/")
	params["path_spec"] = pathSpec

	_, repoPath, _, _, _, _, _ := getRepoInfo(request.Context(), groupPath, repoName, "")

	client, err := git2c.NewClient(config.Git.Socket)
	if err != nil {
		errorPage500(writer, params, err.Error())
		return
	}
	defer client.Close()

	files, content, err := client.Cmd2(repoPath, pathSpec)
	if err != nil {
		errorPage500(writer, params, err.Error())
		return
	}

	switch {
	case files != nil:
		params["files"] = files
		params["readme_filename"] = "README.md"
		params["readme"] = template.HTML("<p>README rendering here is WIP again</p>") // TODO
		renderTemplate(writer, "repo_raw_dir", params)
	case content != "":
		if misc.RedirectNoDir(writer, request) {
			return
		}
		writer.Header().Set("Content-Type", "application/octet-stream")
		fmt.Fprint(writer, content)
	default:
		errorPage500(writer, params, "Unknown error fetching repo raw data")
	}
}
