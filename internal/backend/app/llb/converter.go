package llb

import (
	"github.com/moby/buildkit/client/llb"

	"github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel"
)

func NewConverter(
	buildCtx string,
	cacheNamespace string,
	disableProgressGrouping bool,
	forcePull bool,
) *CommonConverter {
	return &CommonConverter{
		buildCtx: llb.Local(
			buildCtx,
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.InternalLabel, "Loading context")),
		),
		cacheNs:                 cacheNamespace,
		disableProgressGrouping: disableProgressGrouping,
		forcePull:               forcePull,
		caches:                  map[string]llb.MountOption{},
		resolvedImages:          map[string]image{},
	}
}

// CommonConverter - contains common modules for llb conversion
type CommonConverter struct {
	buildCtx llb.State
	cacheNs  string

	disableProgressGrouping bool
	forcePull               bool

	caches         map[string]llb.MountOption
	resolvedImages map[string]image
}
