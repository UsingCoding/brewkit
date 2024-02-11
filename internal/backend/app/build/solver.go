package build

import (
	"context"
	"fmt"

	buildkitclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	gatewayclient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/entitlements"
	"github.com/moby/buildkit/util/progress/progresswriter"
	"golang.org/x/sync/errgroup"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/backend/app/buildkit"
	"github.com/ispringtech/brewkit/internal/common/maps"
)

type buildSolver struct {
	client  buildkit.Client
	context string // path to context

	entitlements maps.Set[api.Entitlement]
}

func (s buildSolver) solve(
	ctx context.Context,
	progress func() (progresswriter.Writer, error),
	sessionAttachable func() ([]session.Attachable, error),
	exports []buildkitclient.ExportEntry,
	llbProvider func(ctx context.Context, client gatewayclient.Client) (llb.State, error),
) error {
	pw, err := progress()
	if err != nil {
		return err
	}

	attachable, err := sessionAttachable()
	if err != nil {
		return err
	}

	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	opt := buildkitclient.SolveOpt{
		Exports: exports,
		LocalDirs: map[string]string{
			buildCtxKey: s.context,
		},
		Session: attachable,
		AllowedEntitlements: maps.ToSlice(s.entitlements, func(e api.Entitlement, s struct{}) entitlements.Entitlement {
			switch e {
			case api.EntitlementNetworkHost:
				return entitlements.EntitlementNetworkHost
			default:
				panic(fmt.Sprintf("unknown entitlement %s", e))
			}
		}),
	}

	eg.Go(func() error {
		_, err = s.client.Build(
			ctx,
			opt,
			product,
			func(ctx context.Context, client gatewayclient.Client) (*gatewayclient.Result, error) {
				state, err2 := llbProvider(ctx, client)
				if err2 != nil {
					return nil, err2
				}

				def, err2 := state.Marshal(ctx)
				if err2 != nil {
					return nil, err2
				}

				req := gatewayclient.SolveRequest{
					Definition: def.ToPB(),
				}

				return client.Solve(ctx, req)
			},
			pw.Status(),
		)
		return err
	})

	eg.Go(func() error {
		<-pw.Done()
		return pw.Err()
	})

	return eg.Wait()
}
