package llb

import (
	"fmt"

	"github.com/moby/buildkit/client/llb"

	"github.com/ispringtech/brewkit/internal/backend/api"
)

func (conv *CommonConverter) proceedCache(caches []api.Cache) (opts []llb.RunOption) {
	for _, cache := range caches {
		mount, ok := conv.caches[cache.ID]
		if !ok {
			mount = llb.AsPersistentCacheDir(
				fmt.Sprintf("%s/%s", conv.cacheNs, cache.ID),
				llb.CacheMountShared,
			)
			conv.caches[cache.ID] = mount
		}

		opts = append(opts, llb.AddMount(
			cache.Path,
			llb.Scratch(),
			mount,
		))
	}
	return
}

type cacheExports struct {
	state llb.State
	src   string
	dst   string
}
