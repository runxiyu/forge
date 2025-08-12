package repo

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
	"go.lindenii.runxiyu.org/forge/forged/internal/ipc/git2c"
)

type logAuthor struct {
	Name  string
	Email string
	When  time.Time
}

type logCommit struct {
	Hash    string
	Message string
	Author  logAuthor
}

func (h *HTTP) Log(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
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

	var refspec string
	if base.RefType == "" {
		refspec = ""
	} else {
		hex, rerr := client.ResolveRef(repoPath, base.RefType, base.RefName)
		if rerr != nil {
			slog.Error("resolve ref failed", "error", rerr)
			refspec = ""
		} else {
			refspec = hex
		}
	}

	var rawCommits []git2c.Commit
	rawCommits, err = client.Log(repoPath, refspec, 0)
	var commitsErr error
	if err != nil {
		commitsErr = err
		slog.Error("git2d log failed", "error", err)
	}
	commits := make([]logCommit, 0, len(rawCommits))
	for _, c := range rawCommits {
		when, _ := time.Parse("2006-01-02 15:04:05", c.Date)
		commits = append(commits, logCommit{
			Hash:    c.Hash,
			Message: c.Message,
			Author:  logAuthor{Name: c.Author, Email: c.Email, When: when},
		})
	}

	repoURLRoot := "/" + misc.SegmentsToURL(base.GroupPath) + "/-/repos/" + url.PathEscape(repoRow.Name) + "/"
	data := map[string]any{
		"BaseData":         base,
		"group_path":       base.GroupPath,
		"repo_name":        repoRow.Name,
		"repo_description": repoRow.Description,
		"repo_url_root":    repoURLRoot,
		"ref_name":         base.RefName,
		"commits":          commits,
		"commits_err":      &commitsErr,
		"global": map[string]any{
			"forge_title": base.Global.ForgeTitle,
		},
	}
	if err := h.r.Render(w, "repo_log", data); err != nil {
		slog.Error("render repo log", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
