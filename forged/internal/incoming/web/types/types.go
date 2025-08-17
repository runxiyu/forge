package types

import (
	"context"
	"net/http"

	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/global"
)

// BaseData is per-request context computed by the router and read by handlers.
// Keep it small and stable; page-specific data should live in view models.
type BaseData struct {
	UserID         string
	Username       string
	URLSegments    []string
	DirMode        bool
	GroupPath      []string
	SeparatorIndex int
	RefType        string
	RefName        string
	Global         *global.GlobalData
	Queries        *queries.Queries
}

type ctxKey struct{}

// WithBaseData attaches BaseData to a context.
func WithBaseData(ctx context.Context, b *BaseData) context.Context {
	return context.WithValue(ctx, ctxKey{}, b)
}

// Base retrieves BaseData from the request (never nil).
func Base(r *http.Request) *BaseData {
	if v, ok := r.Context().Value(ctxKey{}).(*BaseData); ok && v != nil {
		return v
	}
	return &BaseData{}
}

// Vars are route variables captured by the router (e.g., :repo, *rest).
type Vars map[string]string

// HandlerFunc is the router↔handler function contract.
type HandlerFunc func(http.ResponseWriter, *http.Request, Vars)
