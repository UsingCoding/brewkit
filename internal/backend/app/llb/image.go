package llb

import (
	"context"
	"encoding/json"

	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/client/llb"
	containerimage "github.com/moby/buildkit/exporter/containerimage/image"
	gatewayclient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/common/maps"
	"github.com/ispringtech/brewkit/internal/common/maybe"
)

const (
	scratch = "scratch"
)

func (conv *CommonConverter) llbImage(imgRef string) llb.State {
	if imgRef == scratch {
		return llb.Scratch()
	}

	i, ok := conv.resolvedImages[imgRef]
	if !ok {
		panic(errors.Errorf("img %s no resolved earlier", imgRef))
	}

	return i.st
}

type image struct {
	meta containerimage.Image
	st   llb.State
}

func (conv *CommonConverter) resolveImages(
	ctx context.Context,
	images maps.Set[string],
	client gatewayclient.Client,
) error {
	newImages := maps.SubtractSet(images, maps.FromMapKeys(conv.resolvedImages))
	if len(newImages) == 0 {
		// no new images
		return nil
	}

	eg, ctx := errgroup.WithContext(ctx)
	for ref := range images {
		imgRef := ref
		eg.Go(func() error {
			st := llb.Image(imgRef)

			named, err := reference.ParseNormalizedNamed(imgRef)
			if err != nil {
				return errors.Wrapf(err, "failwwwed to parse ref for %s", imgRef)
			}

			_, _, data, err := client.ResolveImageConfig(ctx, named.String(), llb.ResolveImageConfigOpt{
				ResolveMode: llb.ResolveModePreferLocal.String(),
			})
			if err != nil {
				return errors.Wrapf(err, "failed to resolve img %s", imgRef)
			}

			// WithImageConfig adds env, platform and workdir to state
			st, err = st.WithImageConfig(data)
			if err != nil {
				return errors.Wrapf(err, "failed to init metadata for %s", imgRef)
			}

			var img containerimage.Image
			err = json.Unmarshal(data, &img)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal image config for %s", imgRef)
			}

			// set user to state manually
			st = st.User(img.Config.User)

			if len(img.Config.Shell) > 0 {
				st = st.WithValue(shellKey, img.Config.Shell)
			}

			conv.resolvedImages[imgRef] = image{
				meta: img,
				st:   st,
			}
			return nil
		})
	}

	err := eg.Wait()
	return errors.Wrapf(err, "failed to resolve images")
}

func imagesFromVertex(v api.Vertex) maps.Set[string] {
	res := maps.Set[string]{}

	var recursive func(v api.Vertex)
	recursive = func(v api.Vertex) {
		if f, ok := maybe.JustValid(v.From); ok {
			recursive(*f)
		}

		if s, ok := maybe.JustValid(v.Stage); ok {
			if !maybe.Valid(v.From) && s.From != scratch {
				res.Add(s.From)
			}

			for _, c := range s.Copy {
				copyFrom, ok := maybe.JustValid(c.From)
				if !ok {
					continue
				}

				copyFrom.MapLeft(func(l *api.Vertex) {
					recursive(*l)
				})
			}
		}

		for _, vertex := range v.DependsOn {
			recursive(vertex)
		}
	}
	recursive(v)

	return res
}

func imagesFromVars(vars []api.Var) maps.Set[string] {
	res := maps.Set[string]{}

	for _, v := range vars {
		res.Add(v.From)
	}

	return res
}
