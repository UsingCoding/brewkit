package api

import (
	"github.com/ispringtech/brewkit/internal/common/maps"
)

type BuildParams struct {
	Context string

	V    Vertex
	Vars []Var

	Secrets      []SecretSrc
	Entitlements maps.Set[Entitlement]

	ForcePull bool
}

type BootstrapParams struct {
	Image string
	Wait  bool
}

type ClearParams struct {
	All bool
}
