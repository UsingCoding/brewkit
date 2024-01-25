package llb

import (
	"github.com/moby/buildkit/client/llb"

	"github.com/ispringtech/brewkit/internal/common/either"
	"github.com/ispringtech/brewkit/internal/common/maybe"
)

type copyDTO struct {
	from maybe.Maybe[either.Either[llb.State, string]]
	src  string
	dst  string

	name          string
	progressGroup maybe.Maybe[llb.ConstraintsOpt]
}

func (conv *CommonConverter) proceedCopy(copyDTO []copyDTO, st llb.State) (llb.State, error) {
	for _, c := range copyDTO {
		src := conv.buildCtx

		var err error
		if from, ok := maybe.JustValid(c.from); ok {
			from.
				MapLeft(func(st llb.State) {
					src = st
				}).
				MapRight(func(image string) {
					src = conv.llbImage(image)
				})
			if err != nil {
				return llb.State{}, err
			}
		}

		opts := []llb.ConstraintsOpt{llb.WithCustomName(c.name)}

		if g, ok := maybe.JustValid(c.progressGroup); ok {
			opts = append(opts, g)
		}

		st = st.File(llb.Copy(
			src,
			c.src,
			c.dst,
			&llb.CopyInfo{
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
				AttemptUnpack:       false,
				CreateDestPath:      true,
				AllowWildcard:       true,
				AllowEmptyWildcard:  true,
			},
		), opts...)
	}

	return st, nil
}
