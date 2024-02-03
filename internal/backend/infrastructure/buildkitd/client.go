package buildkitd

import (
	"context"
	"io"

	controlapi "github.com/moby/buildkit/api/services/control"
	buildkitclient "github.com/moby/buildkit/client"
	"github.com/pkg/errors"

	"github.com/ispringtech/brewkit/internal/backend/app/buildkit"
)

type client struct {
	*buildkitclient.Client
}

func (c client) Prune(ctx context.Context, params buildkit.PruneParams) (<-chan buildkit.UsageInfo, error) {
	req := &controlapi.PruneRequest{
		All:          params.All,
		KeepDuration: int64(params.KeepDuration),
		KeepBytes:    params.KeepBytes,
	}

	s, err := c.ControlClient().Prune(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call prune")
	}

	out := make(chan buildkit.UsageInfo)

	go func() {
		for {
			d, err2 := s.Recv()
			if err2 != nil {
				if err2 == io.EOF {
					close(out)
					return
				}
				out <- buildkit.UsageInfo{
					Err: err2,
				}
				close(out)
				return
			}

			out <- buildkit.UsageInfo{
				ID:      d.ID,
				Mutable: d.Mutable,
				InUse:   d.InUse,
				Size:    d.Size_,
				Shared:  d.Shared,
				Err:     nil,
			}
		}
	}()

	return out, nil
}
