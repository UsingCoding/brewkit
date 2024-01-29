package api

import (
	"context"
)

type API interface {
	BuilderAPI
	CacheAPI
}

type BuilderAPI interface {
	Build(ctx context.Context, params BuildParams) error
}

type CacheAPI interface {
	ClearCache(ctx context.Context, params ClearParams) error
}
