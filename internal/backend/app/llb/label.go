package llb

import (
	"fmt"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/common/maybe"
)

const (
	copyLabelTpl = "%s: copy%s %s -> %s"
)

func makeCopyLabelVertex(vertexName string, c api.Copy) string {
	var from string
	if f, ok := maybe.JustValid(c.From); ok {
		const fromTpl = " --from=%s"
		f.
			MapLeft(func(v *api.Vertex) {
				from = fmt.Sprintf(fromTpl, v.Name)
			}).
			MapRight(func(imgRef string) {
				from = fmt.Sprintf(fromTpl, imgRef)
			})
	}

	return fmt.Sprintf(copyLabelTpl, vertexName, from, c.Src, c.Dst)
}

func makeCopyLabelVar(varName string, c api.CopyVar) string {
	var from string
	if f, ok := maybe.JustValid(c.From); ok {
		const fromTpl = " --from=%s"
		from = fmt.Sprintf(fromTpl, f)
	}

	return fmt.Sprintf(copyLabelTpl, varName, from, c.Src, c.Dst)
}
