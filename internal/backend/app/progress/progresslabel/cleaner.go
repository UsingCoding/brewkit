package progresslabel

import (
	buildkitclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progresswriter"

	"github.com/ispringtech/brewkit/internal/backend/app/progress/progressinterceptor"
)

// NewLabelsCleaner removes payload labels from vertexes names for rendering
func NewLabelsCleaner(pw progresswriter.Writer) progresswriter.Writer {
	return progressinterceptor.Intercept(pw, func(s *buildkitclient.SolveStatus) error {
		for _, vertex := range s.Vertexes {
			_, text, err := ParsePayloadLabel(vertex.Name)
			if err != nil {
				continue
			}

			if text == "" {
				continue
			}

			vertex.Name = text
		}

		return nil
	})
}
