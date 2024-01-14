package llb

import (
	"github.com/moby/buildkit/client/llb"

	"github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel"
)

func NewConverter(
	buildCtx string,
	cacheNamespace string,
) *CommonConverter {
	return &CommonConverter{
		buildCtx: llb.Local(
			buildCtx,
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.InternalLabel, "Loading context")),
		),
		cacheNs:        cacheNamespace,
		caches:         map[string]llb.MountOption{},
		resolvedImages: map[string]image{},
	}
}

// CommonConverter - contains common modules for llb conversion
type CommonConverter struct {
	buildCtx llb.State
	cacheNs  string

	caches         map[string]llb.MountOption
	resolvedImages map[string]image
}
