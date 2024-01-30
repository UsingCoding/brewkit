package version

import (
	"golang.org/x/exp/slices"
)

const (
	// APIVersionV2 - actual version
	APIVersionV2 = "brewkit/v2"
	// APIVersionV1 - deprecated, soon will be removed
	APIVersionV1 = "brewkit/v1"
)

func SupportedVersions() []string {
	return []string{
		APIVersionV2,
		APIVersionV1,
	}
}

func Supports(v string) bool {
	return slices.Contains(SupportedVersions(), v)
}
