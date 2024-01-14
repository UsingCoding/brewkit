package dockerfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/common/maps"
	"github.com/ispringtech/brewkit/internal/common/maybe"
	"github.com/ispringtech/brewkit/internal/dockerfile"
)

type Vars map[string]string

type TargetGenerator interface {
	GenerateDockerfile() (dockerfile.Dockerfile, error)
}

func NewTargetGenerator(v api.Vertex, vars Vars, dockerfileImage string) TargetGenerator {
	return &targetGenerator{
		v:               v,
		vars:            vars,
		dockerfileImage: dockerfileImage,
		generatedStages: maps.Set[string]{},
	}
}

type targetGenerator struct {
	dockerfileImage string
	v               api.Vertex
	vars            Vars
	generatedStages maps.Set[string]
}

func (generator targetGenerator) GenerateDockerfile() (dockerfile.Dockerfile, error) {
	dockerfileStages, err := generator.stagesForTarget(generator.v)
	if err != nil {
		return dockerfile.Dockerfile{}, err
	}

	return dockerfile.Dockerfile{
		SyntaxHeader: dockerfile.Syntax(generator.dockerfileImage),
		Stages:       dockerfileStages,
	}, nil
}

func (generator targetGenerator) stagesForTarget(v api.Vertex) ([]dockerfile.Stage, error) {
	var stages []dockerfile.Stage
	if maybe.Valid(v.From) {
		s, err := generator.stagesForTarget(*maybe.Just(v.From))
		if err != nil {
			return nil, err
		}
		stages = append(stages, s...)
	}

	if maybe.Valid(v.Stage) {
		stage := maybe.Just(v.Stage)
		s, err := generator.stagesForCopy(stage)
		if err != nil {
			return nil, err
		}
		stages = append(stages, s...)

		s, err = generator.stages(v.Name, stage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate instructions for target %s", v.Name)
		}
		stages = append(stages, s...)
	}

	for _, childV := range v.DependsOn {
		s, err := generator.stagesForTarget(childV)
		if err != nil {
			return nil, err
		}
		stages = append(stages, s...)
	}

	return stages, nil
}

func (generator targetGenerator) stagesForCopy(stage api.Stage) ([]dockerfile.Stage, error) {
	var stages []dockerfile.Stage
	for _, c := range stage.Copy {
		if !maybe.Valid(c.From) {
			// Skip images with empty from
			continue
		}

		var (
			s   []dockerfile.Stage
			err error
		)
		maybe.Just(c.From).
			MapLeft(func(l *api.Vertex) {
				s, err = generator.stagesForTarget(*l)
				if err != nil {
					return
				}
			})
		if err != nil {
			return nil, err
		}
		stages = append(stages, s...)
	}
	return stages, nil
}

func (generator targetGenerator) stages(name string, stage api.Stage) ([]dockerfile.Stage, error) {
	// Stage was already generated by another dependency
	if generator.generatedStages.Has(name) {
		return nil, nil
	}

	instructions, err := generator.instructionsForStage(stage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate instructions for stage")
	}

	stages := []dockerfile.Stage{
		{
			From:         stage.From,
			As:           maybe.NewJust(name),
			Instructions: instructions,
		},
	}

	if len(stage.Output) != 0 {
		// legacy build supports only one export
		output := stage.Output[0]

		const pwd = "."
		stages = append(stages, dockerfile.Stage{
			From: dockerfile.Scratch,
			As:   maybe.NewJust(fmt.Sprintf("%s-out", name)),
			Instructions: []dockerfile.Instruction{
				dockerfile.Copy{
					Src:  output.Artifact,
					Dst:  pwd,
					From: maybe.NewJust(name),
				},
			},
		})
	}

	generator.generatedStages[name] = struct{}{}

	return stages, nil
}

func (generator targetGenerator) instructionsForStage(stage api.Stage) ([]dockerfile.Instruction, error) {
	//nolint:prealloc
	var instructions []dockerfile.Instruction

	if w, ok := maybe.JustValid(stage.WorkDir); ok {
		instructions = append(instructions, dockerfile.Workdir(w))
	}

	for k, v := range stage.Env {
		instructions = append(instructions, dockerfile.Env{
			K: k,
			V: v,
		})
	}

	for _, c := range stage.Copy {
		var from maybe.Maybe[string]
		if maybe.Valid(c.From) {
			maybe.Just(c.From).
				MapLeft(func(v *api.Vertex) {
					from = maybe.NewJust(v.Name)
				}).
				MapRight(func(image string) {
					from = maybe.NewJust(image)
				})
		}

		instructions = append(instructions, dockerfile.Copy{
			Src:  c.Src,
			Dst:  c.Dst,
			From: from,
		})
	}

	//nolint:prealloc
	var mounts []dockerfile.Mount

	for _, cache := range stage.Cache {
		mounts = append(mounts, dockerfile.MountCache{
			ID:     maybe.NewJust(cache.ID),
			Target: cache.Path,
		})
	}

	for _, secret := range stage.Secrets {
		mounts = append(mounts, dockerfile.MountSecret{
			ID:       maybe.NewJust(secret.ID),
			Target:   maybe.NewJust(secret.MountPath),
			Required: maybe.NewJust(true), // make error if secret unavailable
		})
	}

	if maybe.Valid(stage.SSH) {
		mounts = append(mounts, dockerfile.MountSSH{
			Required: maybe.NewJust(true), // make error if ssh key unavailable
		})
	}

	if maybe.Valid(stage.Command) {
		var network string
		if maybe.Valid(stage.Network) {
			network = maybe.Just(stage.Network).Network
		}

		cmd := strings.Join(maybe.Just(stage.Command).Cmd, " ")
		command := generator.fillCommandWithVariables(cmd)

		command = generator.transformToHeredoc(command)

		instructions = append(instructions, dockerfile.Run{
			Mounts:  mounts,
			Network: network,
			Command: command,
		})
	}

	return instructions, nil
}

func (generator targetGenerator) fillCommandWithVariables(command string) string {
	return os.Expand(command, func(v string) string {
		return generator.vars[v]
	})
}

func (generator targetGenerator) transformToHeredoc(s string) string {
	const heredocHeader = "EOF"

	return fmt.Sprintf("<<%s\n%s\n%s", heredocHeader, s, heredocHeader)
}
