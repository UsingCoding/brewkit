package llb

import (
	`github.com/moby/buildkit/client/llb`
	`github.com/pkg/errors`

	`github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel`
)

func (conv *VertexConverter) proceedExport(src llb.State) (llb.State, error) {
	for name, output := range conv.exports {
		st, ok := conv.visitedVertexes[name]
		if !ok {
			return llb.State{}, errors.Errorf("logic error: state for target %s not proceeded", name)
		}

		src = src.File(
			llb.Copy(
				st,
				output.Artifact,
				output.Local,
				&llb.CopyInfo{
					AllowWildcard:  true,
					CreateDestPath: true,
				},
			),
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.HiddenLabel, "exporting artifact for %s", name)),
		)
	}

	for id, exports := range conv.cacheExports {
		src = src.File(
			llb.Copy(
				exports.state,
				"/",
				exports.dst,
				&llb.CopyInfo{
					AllowWildcard:  true,
					CreateDestPath: true,
				},
			),
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.HiddenLabel, "exporting cache from %s", id)),
		)
	}

	return src, nil
}
