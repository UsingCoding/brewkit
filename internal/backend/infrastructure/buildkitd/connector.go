package buildkitd

import (
	"context"
	"net"

	buildkitclient "github.com/moby/buildkit/client"
	"github.com/pkg/errors"

	dockerclient "github.com/docker/docker/client"

	"github.com/ispringtech/brewkit/internal/backend/app/buildkit"
)

func NewConnector() buildkit.Connector {
	return &connector{}
}

type connector struct {
}

func (c *connector) Connect(ctx context.Context, address string) (buildkit.Client, error) {
	bkcl, err := buildkitclient.New(
		ctx,
		address,
		buildkitclient.WithFailFast(),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to buildkit instance at %s", address)
	}

	return &client{
		Client: bkcl,
	}, nil
}

func (c *connector) ConnectToMoby(ctx context.Context) (buildkit.Client, error) {
	dockerCli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the client")
	}

	bkcli, err := buildkitclient.New(
		ctx,
		"moby-worker://brewkit_buildkitd",
		buildkitclient.WithFailFast(),
		buildkitclient.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return dockerCli.DialHijack(ctx, "/grpc", "h2c", nil)
		}),
		buildkitclient.WithSessionDialer(func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
			return dockerCli.DialHijack(ctx, "/session", proto, meta)
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create buildkit client")
	}

	return &client{
		Client: bkcli,
	}, nil
}
