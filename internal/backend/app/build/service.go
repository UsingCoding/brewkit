package build

import (
	`context`
	`os`

	buildkitclient `github.com/moby/buildkit/client`
	`github.com/moby/buildkit/client/llb`
	gatewayclient `github.com/moby/buildkit/frontend/gateway/client`
	`github.com/moby/buildkit/session`
	`github.com/moby/buildkit/util/progress/progresswriter`
	`github.com/pkg/errors`

	`github.com/ispringtech/brewkit/internal/backend/api`
	`github.com/ispringtech/brewkit/internal/backend/app/buildkit`
	llbconv `github.com/ispringtech/brewkit/internal/backend/app/llb`
	`github.com/ispringtech/brewkit/internal/backend/app/progress`
	`github.com/ispringtech/brewkit/internal/backend/app/progress/progresscatcher`
	`github.com/ispringtech/brewkit/internal/backend/app/progress/progressui`
	`github.com/ispringtech/brewkit/internal/backend/app/ssh`
)

const (
	buildCtxKey = "build-context"

	product        = "brewkit"
	cacheNamespace = "brewkit"
)

type Service interface {
	api.BuilderAPI
}

func NewService(
	connector buildkit.Connector,
	sshAgentProvider ssh.AgentProvider,
	verbose bool,
) Service {
	return &service{
		connector:        connector,
		sshAgentProvider: sshAgentProvider,
		verbose:          verbose,
	}
}

type service struct {
	connector        buildkit.Connector
	sshAgentProvider ssh.AgentProvider

	verbose bool
}

func (s *service) Build(
	ctx context.Context,
	v api.Vertex,
	vars []api.Var,
	secrets []api.SecretSrc,
	params api.BuildParams,
) error {
	client, err := s.connector.ConnectToMoby(ctx)
	if err != nil {
		return err
	}

	conv := llbconv.NewConverter(
		buildCtxKey,
		cacheNamespace,
	)

	solver := buildSolver{
		client:  client,
		context: params.Context,
	}

	varsData, err := s.solveVars(ctx, solver, conv, vars, secrets)
	if err != nil {
		return errors.Wrap(err, "failed to solve vars")
	}

	err = s.solveVertex(ctx, solver, conv, v, secrets, varsData)
	if err != nil {
		return errors.Wrap(err, "failed to solve build vertex")
	}

	return nil
}

func (s *service) solveVars(
	ctx context.Context,
	solver buildSolver,
	conv *llbconv.CommonConverter,
	vars []api.Var,
	secrets []api.SecretSrc,
) (map[string]string, error) {
	var catcher progresscatcher.OutputCatcher

	err := solver.solve(
		ctx,
		func() (progresswriter.Writer, error) {
			pw, err := makeVarsProgressWriter(ctx)
			if err != nil {
				return nil, err
			}

			pw, catcher = progresscatcher.New(pw)
			return pw, nil
		},
		func() ([]session.Attachable, error) {
			return s.makeVarsAttachable(vars, secrets)
		},
		nil,
		func(ctx context.Context, client gatewayclient.Client) (llb.State, error) {
			return llbconv.NewVarsConverter(conv).VarsToLLB(ctx, vars, client)
		},
	)
	logs := catcher.Logs()
	if len(logs) == 0 {
		return nil, err
	}

	res := make(map[string]string, len(logs))
	for varName, log := range logs {
		res[varName] = string(log)
	}

	return res, err
}

func (s *service) solveVertex(
	ctx context.Context,
	solver buildSolver,
	conv *llbconv.CommonConverter,
	v api.Vertex,
	secrets []api.SecretSrc,
	vars map[string]string,
) error {
	return solver.solve(
		ctx,
		func() (progresswriter.Writer, error) {
			return makeVertexProgressWriter(ctx)
		},
		func() ([]session.Attachable, error) {
			return s.makeVertexAttachable(v, secrets)
		},
		[]buildkitclient.ExportEntry{{
			Type:      buildkitclient.ExporterLocal,
			OutputDir: ".",
		}},
		func(ctx context.Context, client gatewayclient.Client) (llb.State, error) {
			return llbconv.NewVertexConverter(conv, vars).VertexToLLB(ctx, &v, client)
		},
	)
}

func makeVarsProgressWriter(ctx context.Context) (progresswriter.Writer, error) {
	return progress.NewPrinter(
		ctx,
		os.Stderr,
		progress.AUTO,
		progressui.WithPhase("Solving variables"),
	)
}

func makeVertexProgressWriter(ctx context.Context) (progresswriter.Writer, error) {
	return progress.NewPrinter(
		ctx,
		os.Stderr,
		progress.AUTO,
		progressui.WithPhase("Building vertex"),
	)
}
