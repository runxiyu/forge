package handlers

import (
	"log"
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type IndexHTTP struct {
	r templates.Renderer
}

func NewIndexHTTP(r templates.Renderer) *IndexHTTP { return &IndexHTTP{r: r} }

func (h *IndexHTTP) Index(w http.ResponseWriter, _ *http.Request, _ wtypes.Vars) {
	err := h.r.Render(w, "index", struct {
		Title string
	}{
		Title: "Home",
	})
	if err != nil {
		log.Println("failed to render index page", "error", err)
	}
}
