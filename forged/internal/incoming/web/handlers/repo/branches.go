package repo

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
	"go.lindenii.runxiyu.org/forge/forged/internal/ipc/git2c"
)

func (h *HTTP) Branches(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repoName := v["repo"]

	var userID int64
	if base.UserID != "" {
		_, _ = fmt.Sscan(base.UserID, &userID)
	}
	grp, err := base.Global.Queries.GetGroupByPath(r.Context(), queries.GetGroupByPathParams{Column1: base.GroupPath, UserID: userID})
	if err != nil {
		slog.Error("get group by path", "error", err)
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	repoRow, err := base.Global.Queries.GetRepoByGroupAndName(r.Context(), queries.GetRepoByGroupAndNameParams{GroupID: grp.ID, Name: repoName})
	if err != nil {
		slog.Error("get repo by name", "error", err)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	repoPath := filepath.Join(base.Global.Config.Git.RepoDir, fmt.Sprintf("%d.git", repoRow.ID))
	client, err := git2c.NewClient(r.Context(), base.Global.Config.Git.Socket)
	if err != nil {
		slog.Error("git2d connect failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = client.Close() }()

	branches, err := client.ListBranches(repoPath)
	if err != nil {
		slog.Error("list branches failed", "error", err)
		branches = nil
	}

	repoURLRoot := "/" + misc.SegmentsToURL(base.GroupPath) + "/-/repos/" + url.PathEscape(repoRow.Name) + "/"
	data := map[string]any{
		"BaseData":         base,
		"group_path":       base.GroupPath,
		"repo_name":        repoRow.Name,
		"repo_description": repoRow.Description,
		"repo_url_root":    repoURLRoot,
		"branches":         branches,
		"global": map[string]any{
			"forge_title": base.Global.ForgeTitle,
		},
	}
	if err := h.r.Render(w, "repo_branches", data); err != nil {
		slog.Error("render repo branches", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
