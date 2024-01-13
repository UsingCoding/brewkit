package buildkit

import (
	"context"
)

type Connector interface {
	Connect(ctx context.Context, address string) (Client, error)
	ConnectToMoby(ctx context.Context) (Client, error)
}
