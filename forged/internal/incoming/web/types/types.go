package types

import (
	"context"
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/global"
)

type BaseData struct {
	UserID         string
	Username       string
	URLSegments    []string
	DirMode        bool
	GroupPath      []string
	SeparatorIndex int
	RefType        string
	RefName        string
	Global         *global.Global
}

type ctxKey struct{}

func WithBaseData(ctx context.Context, b *BaseData) context.Context {
	return context.WithValue(ctx, ctxKey{}, b)
}

func Base(r *http.Request) *BaseData {
	if v, ok := r.Context().Value(ctxKey{}).(*BaseData); ok && v != nil {
		return v
	}
	return &BaseData{}
}

type Vars map[string]string

type HandlerFunc func(http.ResponseWriter, *http.Request, Vars)
