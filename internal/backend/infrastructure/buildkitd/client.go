package buildkitd

import (
	buildkitclient "github.com/moby/buildkit/client"
)

type client struct {
	*buildkitclient.Client
}

func (c client) Native() *buildkitclient.Client {
	return c.Client
}
