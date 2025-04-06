// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"net/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"go.lindenii.runxiyu.org/forge/forged/internal/web"
)

// httpHandleRepoLog provides a page with a complete Git log.
//
// TODO: This currently provides all commits in the branch. It should be
// paginated and cached instead.
func (s *Server) httpHandleRepoLog(writer http.ResponseWriter, req *http.Request, params map[string]any) {
	var repo *git.Repository
	var refHash plumbing.Hash
	var err error

	repo = params["repo"].(*git.Repository)

	if refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting ref hash: "+err.Error())
		return
	}

	logOptions := git.LogOptions{From: refHash} //exhaustruct:ignore
	commitIter, err := repo.Log(&logOptions)
	if err != nil {
		web.ErrorPage500(s.templates, writer, params, "Error getting recent commits: "+err.Error())
		return
	}
	params["commits"], params["commits_err"] = commitIterSeqErr(req.Context(), commitIter)

	s.renderTemplate(writer, "repo_log", params)
}
