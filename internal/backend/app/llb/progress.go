package llb

import (
	"github.com/google/uuid"
	"github.com/moby/buildkit/client/llb"

	"github.com/ispringtech/brewkit/internal/common/maybe"
)

func (conv *CommonConverter) makeProgressGroup(name string) (maybe.Maybe[llb.ConstraintsOpt], error) {
	if conv.disableProgressGrouping {
		return maybe.Maybe[llb.ConstraintsOpt]{}, nil
	}

	gID, err := uuid.NewUUID()
	if err != nil {
		return maybe.Maybe[llb.ConstraintsOpt]{}, err
	}

	return maybe.NewJust(llb.ProgressGroup(gID.String(), name, false)), nil
}
