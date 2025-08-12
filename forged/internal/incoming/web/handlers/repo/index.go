package repo

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
	"go.lindenii.runxiyu.org/forge/forged/internal/ipc/git2c"
)

func (h *HTTP) Index(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repoName := v["repo"]
	slog.Info("repo index", "group_path", base.GroupPath, "repo", repoName)

	var userID int64
	if base.UserID != "" {
		_, _ = fmt.Sscan(base.UserID, &userID)
	}
	grp, err := base.Global.Queries.GetGroupByPath(r.Context(), queries.GetGroupByPathParams{
		Column1: base.GroupPath,
		UserID:  userID,
	})
	if err != nil {
		slog.Error("get group by path", "error", err)
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	repoRow, err := base.Global.Queries.GetRepoByGroupAndName(r.Context(), queries.GetRepoByGroupAndNameParams{
		GroupID: grp.ID,
		Name:    repoName,
	})
	if err != nil {
		slog.Error("get repo by name", "error", err)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	repoPath := filepath.Join(base.Global.Config.Git.RepoDir, fmt.Sprintf("%d.git", repoRow.ID))

	var commits []git2c.Commit
	var readme template.HTML
	var commitsErr error
	var readmeFile *git2c.FilenameContents
	var cerr error
	client, err := git2c.NewClient(r.Context(), base.Global.Config.Git.Socket)
	if err == nil {
		defer func() { _ = client.Close() }()
		commits, readmeFile, cerr = client.CmdIndex(repoPath)
		if cerr != nil {
			commitsErr = cerr
			slog.Error("git2d CmdIndex failed", "error", cerr, "path", repoPath)
		} else if readmeFile != nil {
			nameLower := strings.ToLower(readmeFile.Filename)
			if strings.HasSuffix(nameLower, ".md") || strings.HasSuffix(nameLower, ".markdown") || nameLower == "readme" {
				md := goldmark.New(
					goldmark.WithExtensions(extension.GFM),
				)
				var buf bytes.Buffer
				if err := md.Convert(readmeFile.Content, &buf); err == nil {
					readme = template.HTML(buf.String())
				} else {
					readme = template.HTML(template.HTMLEscapeString(string(readmeFile.Content)))
				}
			} else {
				readme = template.HTML(template.HTMLEscapeString(string(readmeFile.Content)))
			}
		}
	} else {
		commitsErr = err
		slog.Error("git2d connect failed", "error", err)
	}

	sshRoot := strings.TrimSuffix(base.Global.Config.SSH.Root, "/")
	httpRoot := strings.TrimSuffix(base.Global.Config.Web.Root, "/")
	pathPart := misc.SegmentsToURL(base.GroupPath) + "/-/repos/" + url.PathEscape(repoRow.Name)
	sshURL := ""
	httpURL := ""
	if sshRoot != "" {
		sshURL = sshRoot + "/" + pathPart
	}
	if httpRoot != "" {
		httpURL = httpRoot + "/" + pathPart
	}

	var notes []string
	if len(commits) == 0 && commitsErr == nil {
		notes = append(notes, "This repository has no commits yet.")
	}
	if readme == template.HTML("") {
		notes = append(notes, "No README found in the default branch.")
	}
	if sshURL == "" && httpURL == "" {
		notes = append(notes, "Clone URLs not configured (missing SSH root and HTTP root).")
	}

	cloneURL := sshURL
	if cloneURL == "" {
		cloneURL = httpURL
	}

	data := map[string]any{
		"BaseData":         base,
		"group_path":       base.GroupPath,
		"repo_name":        repoRow.Name,
		"repo_description": repoRow.Description,
		"ssh_clone_url":    cloneURL,
		"ref_name":         base.RefName,
		"commits":          commits,
		"commits_err":      &commitsErr,
		"readme":           readme,
		"notes":            notes,
		"global": map[string]any{
			"forge_title": base.Global.ForgeTitle,
		},
	}
	if err := h.r.Render(w, "repo_index", data); err != nil {
		slog.Error("render repo index", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
