package repo

import (
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
)

type HTTP struct {
	r templates.Renderer
}

func NewHTTP(r templates.Renderer) *HTTP {
	return &HTTP{
		r: r,
	}
}
