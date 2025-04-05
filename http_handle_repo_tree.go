// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"html/template"
	"net/http"
	"strings"

	"go.lindenii.runxiyu.org/forge/internal/git2c"
	"go.lindenii.runxiyu.org/forge/internal/render"
)

// httpHandleRepoTree provides a friendly, syntax-highlighted view of
// individual files, and provides directory views that link to these files.
//
// TODO: Do not highlight files that are too large.
func (s *Server) httpHandleRepoTree(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	repoName := params["repo_name"].(string)
	groupPath := params["group_path"].([]string)
	rawPathSpec := params["rest"].(string)
	pathSpec := strings.TrimSuffix(rawPathSpec, "/")
	params["path_spec"] = pathSpec

	_, repoPath, _, _, _, _, _ := s.getRepoInfo(request.Context(), groupPath, repoName, "")

	client, err := git2c.NewClient(s.Config.Git.Socket)
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
		renderTemplate(writer, "repo_tree_dir", params)
	case content != "":
		rendered := render.Highlight(pathSpec, content)
		params["file_contents"] = rendered
		renderTemplate(writer, "repo_tree_file", params)
	default:
		errorPage500(writer, params, "Unknown object type, something is seriously wrong")
	}
}
