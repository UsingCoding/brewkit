package api

import (
	"time"

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
	KeepBytes    int64
	KeepDuration time.Duration
	All          bool
}

type UsageInfo struct {
	ID      string
	Mutable bool
	InUse   bool
	Size    int64
	Shared  bool

	Err error
}
