package buildkit

import (
	"context"
	"time"

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
	Prune(ctx context.Context, params PruneParams) (<-chan UsageInfo, error)
}

type PruneParams struct {
	KeepBytes    int64
	KeepDuration time.Duration
	All          bool
}

type UsageInfo struct {
	ID      string
	Mutable bool
	InUse   bool
	Size    int64
	Shared  bool

	Err error
}
