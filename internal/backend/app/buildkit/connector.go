package buildkit

import (
	"context"
)

type Connector interface {
	// Connect - connects to specific buildkitd
	Connect(ctx context.Context, address string) (Client, error)
	// ConnectToMoby - connects to embedded moby (docker) buildkit
	ConnectToMoby(ctx context.Context) (Client, error)
}
