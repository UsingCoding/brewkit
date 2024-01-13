package buildkit

import (
	"context"

	buildkitclient "github.com/moby/buildkit/client"
	gatewayclient "github.com/moby/buildkit/frontend/gateway/client"
)

type Client interface {
	Build(
		ctx context.Context,
		opt buildkitclient.SolveOpt,
		product string,
		buildFunc gatewayclient.BuildFunc,
		statusChan chan *buildkitclient.SolveStatus,
	) (*buildkitclient.SolveResponse, error)
	Native() *buildkitclient.Client
}
