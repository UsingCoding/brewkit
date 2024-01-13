package llb

import (
	`context`
	`fmt`

	`github.com/moby/buildkit/client/llb`

	`github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel`
)

const (
	tmpFakeDep = "/tmp/fakedep"
	tmpPath    = "/tmp/"
	fakePath   = "/fake-%s-*"
)

// creates fake dependency between states by trying to copy non-existent file
func compose(ctx context.Context, from llb.State, to ...llb.State) (llb.State, error) {
	var populatedFrom bool

	if from.Output() == nil {
		from = from.File(
			llb.Mkdir(tmpFakeDep, 0o755, llb.WithParents(true)),
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.HiddenLabel, "Init scratch")),
		)
		populatedFrom = true
	}

	for _, state := range append(to, from) {
		if state.Output() == nil {
			v, err := state.Value(ctx, targetKey)
			if err != nil {
				return llb.State{}, err
			}
			targetName := v.(string)

			panic(fmt.Sprintf("%s state output is nil", targetName))
		}
	}

	v, err := from.Value(ctx, targetKey)
	if err != nil {
		return llb.State{}, err
	}
	fromTargetName := v.(string)

	for _, t := range to {
		v, err := t.Value(ctx, targetKey)
		if err != nil {
			return llb.State{}, err
		}
		targetName := v.(string)
		p := fmt.Sprintf(fakePath, targetName)

		from = from.File(
			llb.Copy(
				t,
				p,
				tmpPath,
				&llb.CopyInfo{
					CreateDestPath:      true,
					AllowWildcard:       true,
					AllowEmptyWildcard:  true,
					CopyDirContentsOnly: true,
				},
			),
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.HiddenLabel, "Depends %s on %s", fromTargetName, targetName)),
		)
	}

	if populatedFrom {
		from.File(
			llb.Rm(
				tmpFakeDep,
				llb.WithAllowNotFound(true),
			),
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.HiddenLabel, "Remove temporary directory from base layer")),
		)
	}

	return from, nil
}
