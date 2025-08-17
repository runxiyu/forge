package handlers

import (
	"log"
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
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
	p, err := base.Queries.GetGroupIDDescByPath(r.Context(), base.URLSegments)
	if err != nil {
		log.Println("failed to get group ID by path", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	subgroups, err := base.Queries.GetSubgroups(r.Context(), &p.ID)
	if err != nil {
		log.Println("failed to get subgroups", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// TODO
	}
	repos, err := base.Queries.GetReposInGroup(r.Context(), p.ID)
	if err != nil {
		log.Println("failed to get repos in group", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		// TODO
	}
	err = h.r.Render(w, "group", struct {
		BaseData    *wtypes.BaseData
		Subgroups   []queries.GetSubgroupsRow
		Repos       []queries.GetReposInGroupRow
		Description string
	}{
		BaseData:    base,
		Subgroups:   subgroups,
		Repos:       repos,
		Description: p.Description,
	})
	if err != nil {
		log.Println("failed to render index page", "error", err)
	}
}
