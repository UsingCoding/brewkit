package cache

import (
	"context"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/backend/app/buildkit"
	"github.com/ispringtech/brewkit/internal/common/channel"
)

func NewCacheService(connector buildkit.Connector) api.CacheAPI {
	return &cacheService{
		connector: connector,
	}
}

type cacheService struct {
	connector buildkit.Connector
}

func (service *cacheService) ClearCache(ctx context.Context, params api.ClearParams) (<-chan api.UsageInfo, error) {
	client, err := service.connector.ConnectToMoby(ctx)
	if err != nil {
		return nil, err
	}

	out, err := client.Prune(ctx, buildkit.PruneParams{
		KeepBytes:    params.KeepBytes,
		KeepDuration: params.KeepDuration,
		All:          params.All,
	})
	if err != nil {
		return nil, err
	}

	return channel.ProxyIn(out, func(u buildkit.UsageInfo) api.UsageInfo {
		return api.UsageInfo{
			ID:      u.ID,
			Mutable: u.Mutable,
			InUse:   u.InUse,
			Size:    u.Size,
			Shared:  u.Shared,
			Err:     u.Err,
		}
	}), nil
}
