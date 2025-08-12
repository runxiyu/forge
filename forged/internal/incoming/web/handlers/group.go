package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
	"go.lindenii.runxiyu.org/forge/forged/internal/ipc/git2c"
)

type GroupHTTP struct {
	r templates.Renderer
}

func NewGroupHTTP(r templates.Renderer) *GroupHTTP {
	return &GroupHTTP{
		r: r,
	}
}

func (h *GroupHTTP) Index(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	base := wtypes.Base(r)
	userID, err := strconv.ParseInt(base.UserID, 10, 64)
	if err != nil {
		userID = 0
	}

	queryParams := queries.GetGroupByPathParams{
		Column1: base.URLSegments,
		UserID:  userID,
	}
	p, err := base.Global.Queries.GetGroupByPath(r.Context(), queryParams)
	if err != nil {
		slog.Error("failed to get group ID by path", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	subgroups, err := base.Global.Queries.GetSubgroups(r.Context(), &p.ID)
	if err != nil {
		slog.Error("failed to get subgroups", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// TODO: gracefully fail this part of the page
	}
	repos, err := base.Global.Queries.GetReposInGroup(r.Context(), p.ID)
	if err != nil {
		slog.Error("failed to get repos in group", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// TODO: gracefully fail this part of the page
	}
	err = h.r.Render(w, "group", struct {
		BaseData     *wtypes.BaseData
		Subgroups    []queries.GetSubgroupsRow
		Repos        []queries.GetReposInGroupRow
		Description  string
		DirectAccess bool
	}{
		BaseData:     base,
		Subgroups:    subgroups,
		Repos:        repos,
		Description:  p.Description,
		DirectAccess: p.HasRole,
	})
	if err != nil {
		slog.Error("failed to render index page", "error", err)
	}
}

func (h *GroupHTTP) Post(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	base := wtypes.Base(r)
	userID, err := strconv.ParseInt(base.UserID, 10, 64)
	if err != nil {
		userID = 0
	}

	queryParams := queries.GetGroupByPathParams{
		Column1: base.URLSegments,
		UserID:  userID,
	}
	p, err := base.Global.Queries.GetGroupByPath(r.Context(), queryParams)
	if err != nil {
		slog.Error("failed to get group ID by path", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !p.HasRole {
		http.Error(w, "You do not have the necessary permissions to create repositories in this group.", http.StatusForbidden)
		return
	}

	name := r.PostFormValue("repo_name")
	desc := r.PostFormValue("repo_desc")
	contrib := r.PostFormValue("repo_contrib")
	if name == "" {
		http.Error(w, "Repo name is required", http.StatusBadRequest)
		return
	}

	if contrib == "" || contrib == "public" {
		contrib = "open"
	}

	tx, err := base.Global.DB.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		slog.Error("begin tx failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(r.Context()) }()

	txq := base.Global.Queries.WithTx(tx)
	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}
	repoID, err := txq.InsertRepo(r.Context(), queries.InsertRepoParams{
		GroupID:             p.ID,
		Name:                name,
		Description:         descPtr,
		ContribRequirements: contrib,
	})
	if err != nil {
		slog.Error("insert repo failed", "error", err)
		http.Error(w, "Failed to create repository", http.StatusInternalServerError)
		return
	}

	repoPath := filepath.Join(base.Global.Config.Git.RepoDir, fmt.Sprintf("%d.git", repoID))

	gitc, err := git2c.NewClient(r.Context(), base.Global.Config.Git.Socket)
	if err != nil {
		slog.Error("git2d connect failed", "error", err)
		http.Error(w, "Failed to initialize repository (backend)", http.StatusInternalServerError)
		return
	}
	defer func() { _ = gitc.Close() }()
	if err = gitc.InitRepo(repoPath, base.Global.Config.Hooks.Execs); err != nil {
		slog.Error("git2d init failed", "error", err)
		http.Error(w, "Failed to initialize repository", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(r.Context()); err != nil {
		slog.Error("commit tx failed", "error", err)
		http.Error(w, "Failed to finalize repository creation", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}
