package build

import (
	`context`

	buildkitclient `github.com/moby/buildkit/client`
	`github.com/moby/buildkit/client/llb`
	gatewayclient `github.com/moby/buildkit/frontend/gateway/client`
	`github.com/moby/buildkit/session`
	`github.com/moby/buildkit/util/progress/progresswriter`
	`github.com/pkg/errors`
	`golang.org/x/sync/errgroup`

	`github.com/ispringtech/brewkit/internal/backend/app/buildkit`
)

type buildSolver struct {
	client  buildkit.Client
	context string // path to context
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

	eg, ctx := errgroup.WithContext(ctx)

	opt := buildkitclient.SolveOpt{
		Exports: exports,
		LocalDirs: map[string]string{
			buildCtxKey: s.context,
		},
		Session: attachable,
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
		return errors.Wrap(err, "failed to solve")
	})

	eg.Go(func() error {
		<-pw.Done()
		return pw.Err()
	})

	return eg.Wait()
}
