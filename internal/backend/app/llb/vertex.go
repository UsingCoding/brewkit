package llb

import (
	`context`

	`github.com/moby/buildkit/client/llb`
	gatewayclient `github.com/moby/buildkit/frontend/gateway/client`

	`github.com/ispringtech/brewkit/internal/backend/api`
	`github.com/ispringtech/brewkit/internal/backend/app/progress/progresslabel`
	`github.com/ispringtech/brewkit/internal/common/either`
	`github.com/ispringtech/brewkit/internal/common/maybe`
	`github.com/ispringtech/brewkit/internal/common/slices`
)

type (
	Vars map[string]string
)

func NewVertexConverter(commonConverter *CommonConverter, vars Vars) *VertexConverter {
	return &VertexConverter{
		CommonConverter: commonConverter,
		vars:            vars,
		visitedVertexes: map[string]llb.State{},
		exports:         map[string]api.Output{},
		cacheExports:    map[string]cacheExports{},
	}
}

type VertexConverter struct {
	*CommonConverter

	vars Vars

	visitedVertexes map[string]llb.State
	exports         map[string]api.Output
	cacheExports    map[string]cacheExports
}

func (conv *VertexConverter) VertexToLLB(
	ctx context.Context,
	v *api.Vertex,
	client gatewayclient.Client,
) (llb.State, error) {
	err := conv.resolveImages(ctx, imagesFromVertex(*v), client)
	if err != nil {
		return llb.State{}, err
	}

	res, err := conv.vertexToState(ctx, v)
	if err != nil {
		return llb.State{}, err
	}

	if res.Output() != nil {
		res = res.File(
			// use wildcard to clear filesystem since llbsolver ensures that path is not absolute
			llb.Rm("/*", &llb.RmInfo{
				AllowNotFound: true,
				AllowWildcard: true,
			}),
			llb.WithCustomName(progresslabel.MakeLabelf(progresslabel.HiddenLabel, "Clear result state")),
		)
	}

	res, err = conv.proceedExport(res)
	if err != nil {
		return llb.State{}, err
	}

	return res, nil
}

func (conv *VertexConverter) vertexToState(ctx context.Context, v *api.Vertex) (llb.State, error) {
	if s, ok := conv.visitedVertexes[v.Name]; ok {
		return s, nil
	}

	st := llb.Scratch()

	if from, ok := maybe.JustValid(v.From); ok {
		var err error
		st, err = conv.vertexToState(ctx, from)
		if err != nil {
			return llb.State{}, err
		}
	}

	st = st.WithValue(targetKey, v.Name)

	if len(v.DependsOn) != 0 {
		deps, err := slices.MapErr(v.DependsOn, func(dep api.Vertex) (llb.State, error) {
			return conv.vertexToState(ctx, &dep)
		})
		if err != nil {
			return llb.State{}, err
		}

		st, err = compose(ctx, st, deps...)
		if err != nil {
			return llb.State{}, err
		}
	}

	if stage, ok := maybe.JustValid(v.Stage); ok {
		// TODO: raise above DependsOn composing
		if st.Output() == nil {
			st = conv.llbImage(stage.From)
		}

		var err error
		st, err = conv.populateState(ctx, stage, st)
		if err != nil {
			return llb.State{}, err
		}

		if o, ok := maybe.JustValid(stage.Output); ok {
			conv.exports[v.Name] = o
		}
	}

	conv.visitedVertexes[v.Name] = st

	return st, nil
}

func (conv *VertexConverter) populateState(ctx context.Context, s api.Stage, st llb.State) (llb.State, error) {
	if w, ok := maybe.JustValid(s.WorkDir); ok {
		st = st.Dir(w)
	}

	for k, v := range s.Env {
		st = st.AddEnv(k, v)
	}

	dtos, err := slices.MapErr(s.Copy, func(c api.Copy) (copyDTO, error) {
		from, err := conv.convertFromForCopy(ctx, c.From)
		if err != nil {
			return copyDTO{}, err
		}

		return copyDTO{
			From: from,
			Src:  c.Src,
			Dst:  c.Dst,
		}, nil
	})
	if err != nil {
		return llb.State{}, err
	}

	st, err = conv.proceedCopy(dtos, st)
	if err != nil {
		return llb.State{}, err
	}

	st, err = conv.proceedCommand(ctx, cmdDTO{
		command: s.Command,
		cache:   s.Cache,
		ssh:     s.SSH,
		secrets: s.Secrets,
		params:  conv.vars,
	}, st)
	if err != nil {
		return llb.State{}, err
	}

	return st, nil
}

func (conv *VertexConverter) convertFromForCopy(
	ctx context.Context,
	from maybe.Maybe[either.Either[*api.Vertex, string]],
) (
	res maybe.Maybe[either.Either[llb.State, string]],
	err error,
) {
	f, ok := maybe.JustValid(from)
	if !ok {
		return
	}

	f.
		MapLeft(func(v *api.Vertex) {
			var state llb.State
			state, err = conv.vertexToState(ctx, v)
			if err != nil {
				return
			}
			res = maybe.NewJust(either.NewLeft[llb.State, string](state))
		}).
		MapRight(func(image string) {
			res = maybe.NewJust(either.NewRight[llb.State, string](image))
		})

	return res, err
}
