package llb

import (
	`context`

	`github.com/moby/buildkit/client/llb`
	gatewayclient `github.com/moby/buildkit/frontend/gateway/client`
	`github.com/pkg/errors`

	`github.com/ispringtech/brewkit/internal/backend/api`
	`github.com/ispringtech/brewkit/internal/backend/app/progress/progresscatcher`
	`github.com/ispringtech/brewkit/internal/common/either`
	`github.com/ispringtech/brewkit/internal/common/maybe`
	`github.com/ispringtech/brewkit/internal/common/slices`
)

func NewVarsConverter(commonConverter *CommonConverter) *VarsConverter {
	return &VarsConverter{CommonConverter: commonConverter}
}

type VarsConverter struct {
	*CommonConverter
}

func (conv *VarsConverter) VarsToLLB(
	ctx context.Context,
	vars []api.Var,
	client gatewayclient.Client,
) (llb.State, error) {
	err := conv.resolveImages(ctx, imagesFromVars(vars), client)
	if err != nil {
		return llb.State{}, err
	}

	states, err := slices.MapErr(vars, func(v api.Var) (llb.State, error) {
		return conv.varToState(ctx, v)
	})
	if err != nil {
		return llb.State{}, err
	}

	st := llb.Scratch().WithValue(targetKey, "vars")
	res, err := compose(
		ctx,
		st,
		states...,
	)
	if err != nil {
		return llb.State{}, errors.Wrap(err, "failed to compose var states")
	}

	return res, nil
}

func (conv *VarsConverter) varToState(ctx context.Context, v api.Var) (llb.State, error) {
	st := conv.llbImage(v.From)

	st = st.WithValue(targetKey, v.Name)

	if w, ok := maybe.JustValid(v.WorkDir); ok {
		st = st.Dir(w)
	}

	for k, v := range v.Env {
		st = st.AddEnv(k, v)
	}

	st, err := conv.proceedCopy(slices.Map(v.Copy, func(c api.CopyVar) copyDTO {
		from := maybe.Map(c.From, func(image string) either.Either[llb.State, string] {
			return either.NewRight[llb.State, string](image)
		})

		return copyDTO{
			From: from,
			Src:  c.Src,
			Dst:  c.Dst,
		}
	}), st)
	if err != nil {
		return llb.State{}, err
	}

	label, err := progresscatcher.MakeCatchLabelf(
		v.Name,
		cmdCustomName(v.Command.Cmd),
	)
	if err != nil {
		return llb.State{}, errors.Wrapf(err, "failed to make label for var %s", v.Name)
	}

	st, err = conv.proceedCommand(ctx, cmdDTO{
		command:     maybe.NewJust(v.Command),
		cache:       v.Cache,
		ssh:         v.SSH,
		secrets:     v.Secrets,
		ignoreCache: true, // ignore build cache for variables
		label:       label,
	}, st)
	if err != nil {
		return llb.State{}, err
	}

	return st, nil
}