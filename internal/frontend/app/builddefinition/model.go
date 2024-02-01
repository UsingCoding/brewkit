package builddefinition

import (
	stdslices "golang.org/x/exp/slices"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/common/maybe"
	"github.com/ispringtech/brewkit/internal/common/slices"
)

type Definition struct {
	Vertexes []api.Vertex
	Vars     []api.Var
}

func (d Definition) Vertex(name string) maybe.Maybe[api.Vertex] {
	i := stdslices.IndexFunc(d.Vertexes, func(vertex api.Vertex) bool {
		return vertex.Name == name && !vertex.Private
	})
	if i == -1 {
		return maybe.Maybe[api.Vertex]{}
	}

	return maybe.NewJust(d.Vertexes[i])
}

func (d Definition) ListPublicTargetNames() []string {
	return slices.MapMaybe(d.Vertexes, func(v api.Vertex) maybe.Maybe[string] {
		if v.Private {
			return maybe.Maybe[string]{}
		}
		return maybe.NewJust(v.Name)
	})
}
