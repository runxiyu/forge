// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"net/http"

	"go.lindenii.runxiyu.org/forge/internal/git2c"
	"go.lindenii.runxiyu.org/forge/internal/render"
	"go.lindenii.runxiyu.org/forge/internal/web"
)

// httpHandleRepoIndex provides the front page of a repo using git2d.
func (s *Server) httpHandleRepoIndex(w http.ResponseWriter, req *http.Request, params map[string]any) {
	repoName := params["repo_name"].(string)
	groupPath := params["group_path"].([]string)

	_, repoPath, _, _, _, _, _ := s.getRepoInfo(req.Context(), groupPath, repoName, "") // TODO: Don't use getRepoInfo

	client, err := git2c.NewClient(s.config.Git.Socket)
	if err != nil {
		web.ErrorPage500(templates, w, params, err.Error())
		return
	}
	defer client.Close()

	commits, readme, err := client.Cmd1(repoPath)
	if err != nil {
		web.ErrorPage500(templates, w, params, err.Error())
		return
	}

	params["commits"] = commits
	params["readme_filename"] = readme.Filename
	_, params["readme"] = render.Readme(readme.Content, readme.Filename)

	renderTemplate(w, "repo_index", params)

	// TODO: Caching
}
