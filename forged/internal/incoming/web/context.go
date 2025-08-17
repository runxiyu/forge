package web

import (
    "context"
    "net/http"
)

type BaseData struct {
    Global         any
    UserID         int
    Username       string
    URLSegments    []string
    DirMode        bool
    GroupPath      []string
    SeparatorIndex int
}

type ctxKey int

const baseDataKey ctxKey = iota

func WithBaseData(ctx context.Context, b *BaseData) context.Context { return context.WithValue(ctx, baseDataKey, b) }

func Base(r *http.Request) *BaseData {
    if v, ok := r.Context().Value(baseDataKey).(*BaseData); ok && v != nil {
        return v
    }
    return &BaseData{}
}

