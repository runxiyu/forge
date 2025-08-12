package repo

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
	"go.lindenii.runxiyu.org/forge/forged/internal/ipc/git2c"
)

func (h *HTTP) Tree(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repoName := v["repo"]
	rawPathSpec := v["rest"]
	pathSpec := strings.TrimSuffix(rawPathSpec, "/")

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

	files, content, err := client.CmdTreeRaw(repoPath, pathSpec)
	if err != nil {
		slog.Error("git2d CmdTreeRaw failed", "error", err, "path", repoPath, "spec", pathSpec)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	repoURLRoot := "/" + misc.SegmentsToURL(base.GroupPath) + "/-/repos/" + url.PathEscape(repoRow.Name) + "/"

	switch {
	case files != nil:
		if !base.DirMode && misc.RedirectDir(w, r) {
			return
		}
		data := map[string]any{
			"BaseData":         base,
			"group_path":       base.GroupPath,
			"repo_name":        repoRow.Name,
			"repo_description": repoRow.Description,
			"repo_url_root":    repoURLRoot,
			"ref_name":         base.RefName,
			"path_spec":        pathSpec,
			"files":            files,
			"readme_filename":  "README.md",
			"readme":           template.HTML("<p>README rendering here is WIP.</p>"),
			"global": map[string]any{
				"forge_title": base.Global.ForgeTitle,
			},
		}
		if err := h.r.Render(w, "repo_tree_dir", data); err != nil {
			slog.Error("render repo tree dir", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	case content != "":
		if base.DirMode && misc.RedirectNoDir(w, r) {
			return
		}
		escaped := template.HTMLEscapeString(content)
		rendered := template.HTML("<pre class=\"chroma\"><code>" + escaped + "</code></pre>")
		data := map[string]any{
			"BaseData":         base,
			"group_path":       base.GroupPath,
			"repo_name":        repoRow.Name,
			"repo_description": repoRow.Description,
			"repo_url_root":    repoURLRoot,
			"ref_name":         base.RefName,
			"path_spec":        pathSpec,
			"file_contents":    rendered,
			"global": map[string]any{
				"forge_title": base.Global.ForgeTitle,
			},
		}
		if err := h.r.Render(w, "repo_tree_file", data); err != nil {
			slog.Error("render repo tree file", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Unknown object type", http.StatusInternalServerError)
	}
}
