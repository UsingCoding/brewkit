package builddefinition

import (
	"encoding/json"
	"os"
	"path"

	"github.com/google/go-jsonnet"
	"github.com/pkg/errors"

	"github.com/ispringtech/brewkit/internal/common/either"
	"github.com/ispringtech/brewkit/internal/common/maybe"
	"github.com/ispringtech/brewkit/internal/common/slices"
	"github.com/ispringtech/brewkit/internal/frontend/app/buildconfig"
)

type Parser struct{}

func (parser Parser) Parse(configPath string) (buildconfig.Config, error) {
	data, err := parser.compileConfig(configPath)
	if err != nil {
		return buildconfig.Config{}, err
	}

	var c Config

	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return buildconfig.Config{}, errors.Wrap(err, "failed to parse json config")
	}

	return mapConfig(c)
}

func (parser Parser) CompileConfig(configPath string) (string, error) {
	return parser.compileConfig(configPath)
}

func (parser Parser) compileConfig(configPath string) (string, error) {
	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read build config file")
	}

	vm := jsonnet.MakeVM()

	for _, f := range funcs {
		vm.NativeFunction(f.nativeFunc())
	}

	data, err := vm.EvaluateAnonymousSnippet(path.Base(configPath), string(fileBytes))
	return data, errors.Wrap(err, "failed to compile jsonnet for build definition")
}

func mapConfig(c Config) (buildconfig.Config, error) {
	targets := mapTargets(c.Targets)

	vars, err := mapVars(c.Vars)
	if err != nil {
		return buildconfig.Config{}, err
	}

	return buildconfig.Config{
		APIVersion: c.APIVersion,
		Targets:    targets,
		Vars:       vars,
	}, nil
}

func mapTargets(targets map[string]either.Either[[]string, Target]) []buildconfig.TargetData {
	result := make([]buildconfig.TargetData, 0, len(targets))
	for name, target := range targets {
		target.
			MapLeft(func(dependsOn []string) {
				result = append(result, buildconfig.TargetData{
					Name:      name,
					DependsOn: dependsOn,
					Stage:     maybe.NewNone[buildconfig.StageData](),
				})
			}).
			MapRight(func(t Target) {
				s := maybe.FromPtr(t.Stage)

				result = append(result, buildconfig.TargetData{
					Name:      name,
					DependsOn: t.DependsOn,
					Stage:     maybe.Map(s, mapStage),
				})
			})
	}

	return result
}

func mapStage(stage Stage) buildconfig.StageData {
	return buildconfig.StageData{
		From:     stage.From,
		Platform: stage.Platform,
		WorkDir:  stage.WorkDir,
		Env:      stage.Env,
		Command:  parseCmd(stage.Command),
		SSH:      mapSSH(stage.SSH),
		Cache:    slices.Map(stage.Cache, mapCache),
		Copy:     parseCopy(stage.Copy),
		Secrets:  parseSecret(stage.Secrets),
		Network: maybe.Map(stage.Network, func(n string) string {
			return n
		}),
		Output: maybe.Map(stage.Output, func(o Output) buildconfig.Output {
			return buildconfig.Output{
				Artifact: o.Artifact,
				Local:    o.Local,
			}
		}),
	}
}

func mapVars(vars map[string]Var) ([]buildconfig.VarData, error) {
	result := make([]buildconfig.VarData, 0, len(vars))
	for name, v := range vars {
		data, err := mapVar(name, v)
		if err != nil {
			return nil, errors.Wrapf(err, "var %s", name)
		}
		result = append(result, data)
	}
	return result, nil
}

func mapVar(name string, v Var) (buildconfig.VarData, error) {
	cmd := parseCmd(v.Command)
	if !maybe.Valid(cmd) {
		return buildconfig.VarData{}, errors.Errorf("command is required for var section")
	}

	return buildconfig.VarData{
		Name:     name,
		From:     v.From,
		Platform: v.Platform,
		WorkDir:  v.WorkDir,
		Env:      v.Env,
		SSH:      mapSSH(v.SSH),
		Cache:    slices.Map(v.Cache, mapCache),
		Copy:     parseCopy(v.Copy),
		Secrets:  parseSecret(v.Secrets),
		Network: maybe.Map(v.Network, func(n string) string {
			return n
		}),
		Command: maybe.Just(cmd),
	}, nil
}

func mapSSH(ssh maybe.Maybe[SSH]) maybe.Maybe[buildconfig.SSH] {
	return maybe.Map(ssh, func(s SSH) buildconfig.SSH {
		return buildconfig.SSH{}
	})
}

func mapCache(cache Cache) buildconfig.Cache {
	return buildconfig.Cache{
		ID:   cache.ID,
		Path: cache.Path,
	}
}

func parseCopy(c either.Either[[]Copy, Copy]) (result []buildconfig.Copy) {
	c.
		MapLeft(func(l []Copy) {
			result = slices.Map(l, mapCopy)
		}).
		MapRight(func(r Copy) {
			result = append(result, mapCopy(r))
		})
	return result
}

func parseCmd(cmd either.Either[[]string, string]) maybe.Maybe[buildconfig.Command] {
	var (
		res buildconfig.Command
	)
	cmd.
		MapLeft(func(l []string) {
			res.Cmd = l
			res.Shell = false
		}).
		MapRight(func(r string) {
			res.Cmd = append(res.Cmd, r)
			res.Shell = true
		})

	if len(res.Cmd) == 0 {
		return maybe.Maybe[buildconfig.Command]{}
	}

	return maybe.NewJust(res)
}

func mapCopy(c Copy) buildconfig.Copy {
	return buildconfig.Copy{
		Src:  c.Src,
		Dst:  c.Dst,
		From: c.From,
	}
}

func parseSecret(s either.Either[[]Secret, Secret]) (result []buildconfig.Secret) {
	s.
		MapLeft(func(l []Secret) {
			result = slices.Map(l, mapSecret)
		}).
		MapRight(func(r Secret) {
			result = append(result, mapSecret(r))
		})
	return result
}

func mapSecret(secret Secret) buildconfig.Secret {
	return buildconfig.Secret{
		ID:   secret.ID,
		Path: secret.Path,
	}
}
