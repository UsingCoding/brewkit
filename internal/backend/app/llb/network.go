package llb

import (
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	"github.com/pkg/errors"

	"github.com/ispringtech/brewkit/internal/backend/api"
)

func (conv *CommonConverter) network(n api.Network) llb.StateOption {
	var netMode pb.NetMode

	switch n {
	case api.HostNetwork:
		netMode = pb.NetMode_HOST
	case api.NoneNetwork:
		netMode = pb.NetMode_NONE
	default:
		panic(errors.Errorf("unknown network %s", n))
	}

	return llb.Network(netMode)
}
