package progresscatcher

import (
	buildkitclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/progress/progresswriter"
	"github.com/opencontainers/go-digest"

	"github.com/ispringtech/brewkit/internal/backend/app/progress"
	"github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel"
)

type OutputCatcher interface {
	Logs() map[string][]byte
}

func New(w progresswriter.Writer) (progresswriter.Writer, OutputCatcher) {
	c := &outputCatcher{
		logs:         map[string][]byte{},
		catchDigests: map[digest.Digest]string{},
	}
	w = progress.Intercept(w, c.intercept)

	return w, c
}

const (
	logKey = "key"
)

type outputCatcher struct {
	logs map[string][]byte

	catchDigests map[digest.Digest]string
}

func (c *outputCatcher) Logs() map[string][]byte {
	return c.logs
}

func (c *outputCatcher) intercept(s *buildkitclient.SolveStatus) error {
	for _, vertex := range s.Vertexes {
		payload, _, err := progresslabel.ParsePayloadLabel(vertex.Name)
		if err != nil {
			return err
		}

		if payload == nil {
			continue
		}

		_, ok := payload[progresslabel.CatchOutputLabel]
		if !ok {
			continue
		}

		k := payload[logKey]
		if k == "" {
			continue
		}

		c.catchDigests[vertex.Digest] = k
	}

	for _, log := range s.Logs {
		d := log.Vertex

		k, ok := c.catchDigests[d]
		if !ok {
			continue
		}

		data := make([]byte, len(log.Data))
		_ = copy(data, log.Data)
		_ = k

		c.logs[k] = append(c.logs[k], data...)
	}

	return nil
}
