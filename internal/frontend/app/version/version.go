package version

const (
	// APIVersionV2 - actual version
	APIVersionV2 = "brewkit/v2"
	// APIVersionV1 - soon will be deprecated
	APIVersionV1 = "brewkit/v1"
)

func Supports(v string) bool {
	switch v {
	case APIVersionV2,
		APIVersionV1:
		return true
	default:
		return false
	}
}
