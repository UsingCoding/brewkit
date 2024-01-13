package llb

import (
	`github.com/moby/buildkit/client/llb`

	`github.com/ispringtech/brewkit/internal/common/either`
	`github.com/ispringtech/brewkit/internal/common/maybe`
)

type copyDTO struct {
	From maybe.Maybe[either.Either[llb.State, string]]
	Src  string
	Dst  string
}

func (conv *CommonConverter) proceedCopy(copyDTO []copyDTO, st llb.State) (llb.State, error) {
	for _, c := range copyDTO {
		src := conv.buildCtx

		var err error
		if from, ok := maybe.JustValid(c.From); ok {
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

		st = st.File(llb.Copy(
			src,
			c.Src,
			c.Dst,
			&llb.CopyInfo{
				FollowSymlinks:      true,
				CopyDirContentsOnly: true,
				AttemptUnpack:       false,
				CreateDestPath:      true,
				AllowWildcard:       true,
				AllowEmptyWildcard:  true,
			},
		))
	}

	return st, nil
}
