package handlers

import (
	"log"
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type IndexHTTP struct {
	r templates.Renderer
}

func NewIndexHTTP(r templates.Renderer) *IndexHTTP {
	return &IndexHTTP{
		r: r,
	}
}

func (h *IndexHTTP) Index(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	groups, err := types.Base(r).Queries.GetRootGroups(r.Context())
	if err != nil {
		http.Error(w, "failed to get root groups", http.StatusInternalServerError)
		log.Println("failed to get root groups", "error", err)
		return
	}
	err = h.r.Render(w, "index", struct {
		BaseData *types.BaseData
		Groups   []queries.GetRootGroupsRow
	}{
		BaseData: types.Base(r),
		Groups:   groups,
	})
	if err != nil {
		log.Println("failed to render index page", "error", err)
	}
}
