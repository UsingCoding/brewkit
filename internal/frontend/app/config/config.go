package config

import (
	"github.com/ispringtech/brewkit/internal/common/maps"
)

type Config struct {
	Secrets     []Secret
	Entitlement maps.Set[Entitlement]
}

type Secret struct {
	ID   string
	Path string
}

type Entitlement string

const (
	EntitlementNetworkHost = "network.host"
)
