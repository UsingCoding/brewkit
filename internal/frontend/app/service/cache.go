package service

import (
	"context"
	"time"

	"github.com/ispringtech/brewkit/internal/backend/api"
)

type ClearCacheParam struct {
	KeepBytes    int64
	KeepDuration time.Duration
	All          bool
}

type Cache interface {
	ClearCache(ctx context.Context, param ClearCacheParam) (<-chan api.UsageInfo, error)
}

func NewCacheService(cacheAPI api.CacheAPI) Cache {
	return &cacheService{cacheAPI: cacheAPI}
}

type cacheService struct {
	cacheAPI api.CacheAPI
}

func (service *cacheService) ClearCache(ctx context.Context, param ClearCacheParam) (<-chan api.UsageInfo, error) {
	return service.cacheAPI.ClearCache(ctx, api.ClearParams{
		KeepBytes:    param.KeepBytes,
		KeepDuration: param.KeepDuration,
		All:          param.All,
	})
}
