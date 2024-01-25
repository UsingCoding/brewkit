package llb

import (
	"github.com/google/uuid"
	"github.com/moby/buildkit/client/llb"
)

func (conv *CommonConverter) makeProgressGroup(name string) (llb.ConstraintsOpt, error) {
	gID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	return llb.ProgressGroup(gID.String(), name, false), nil
}
