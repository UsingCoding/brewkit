package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ispringtech/brewkit/internal/backend/api"
	"github.com/ispringtech/brewkit/internal/common/maps"
	"github.com/ispringtech/brewkit/internal/common/maybe"
	"github.com/ispringtech/brewkit/internal/common/slices"
	"github.com/ispringtech/brewkit/internal/frontend/app/buildconfig"
	"github.com/ispringtech/brewkit/internal/frontend/app/builddefinition"
	appconfig "github.com/ispringtech/brewkit/internal/frontend/app/config"
	"github.com/ispringtech/brewkit/internal/frontend/app/version"
)

const (
	allTargetKeyword = "all"
)

type BuildService interface {
	Build(ctx context.Context, p BuildParams) error
	ListTargets(spec buildconfig.Spec) ([]string, error)

	DumpBuildDefinition(spec buildconfig.Spec) (string, error)
	DumpCompiledBuildDefinition(spec buildconfig.Spec) (string, error)
}

type BuildParams struct {
	Targets []string // Target names to run
	Spec    buildconfig.Spec

	Context   string
	ForcePull bool
}

func NewBuildService(
	configParser buildconfig.Parser,
	definitionBuilder builddefinition.Builder,
	builderv1 api.BuilderAPI,
	builderv2 api.BuilderAPI,
	config appconfig.Config,
) BuildService {
	return &buildService{
		configParser:      configParser,
		definitionBuilder: definitionBuilder,
		builderv1:         builderv1,
		builderv2:         builderv2,
		config:            config,
	}
}

type buildService struct {
	configParser      buildconfig.Parser
	definitionBuilder builddefinition.Builder
	builderv1         api.BuilderAPI
	builderv2         api.BuilderAPI
	config            appconfig.Config
}

func (service *buildService) Build(ctx context.Context, p BuildParams) error {
	c, err := service.configParser.Parse(p.Spec)
	if err != nil {
		return err
	}

	definition, err := service.definitionBuilder.Build(c, service.config.Secrets)
	if err != nil {
		return err
	}

	vertex, err := service.buildVertex(p.Targets, definition)
	if err != nil {
		return err
	}

	secrets := slices.Map(service.config.Secrets, func(s appconfig.Secret) api.SecretSrc {
		return api.SecretSrc{
			ID:         s.ID,
			SourcePath: s.Path,
		}
	})

	entitlement := maps.Map(service.config.Entitlement, func(e appconfig.Entitlement, s struct{}) (api.Entitlement, struct{}) {
		switch e {
		case appconfig.EntitlementNetworkHost:
			return api.EntitlementNetworkHost, struct{}{}
		default:
			panic(fmt.Sprintf("unknown entitlement %s", e))
		}
	})

	return service.builderAPI(c.APIVersion).Build(
		ctx,
		api.BuildParams{
			Context:      p.Context,
			V:            vertex,
			Vars:         definition.Vars,
			Secrets:      secrets,
			Entitlements: entitlement,
			ForcePull:    p.ForcePull,
		},
	)
}

func (service *buildService) ListTargets(spec buildconfig.Spec) ([]string, error) {
	c, err := service.configParser.Parse(spec)
	if err != nil {
		return nil, err
	}

	definition, err := service.definitionBuilder.Build(c, service.config.Secrets)
	if err != nil {
		return nil, err
	}

	return definition.ListPublicTargetNames(), nil
}

func (service *buildService) DumpBuildDefinition(spec buildconfig.Spec) (string, error) {
	c, err := service.configParser.Parse(spec)
	if err != nil {
		return "", err
	}

	definition, err := service.definitionBuilder.Build(c, service.config.Secrets)
	if err != nil {
		return "", err
	}

	d, err := json.Marshal(definition)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(d), nil
}

func (service *buildService) DumpCompiledBuildDefinition(spec buildconfig.Spec) (string, error) {
	return service.configParser.CompileConfig(spec)
}

func (service *buildService) buildVertex(targets []string, definition builddefinition.Definition) (api.Vertex, error) {
	if len(targets) == 0 {
		v, err := service.findTarget(allTargetKeyword, definition)
		return v, errors.Wrap(err, "failed to find default target")
	}

	if len(targets) == 1 {
		return service.findTarget(targets[0], definition)
	}

	vertexes, err := slices.MapErr(targets, func(t string) (api.Vertex, error) {
		return service.findTarget(t, definition)
	})
	if err != nil {
		return api.Vertex{}, err
	}

	return api.Vertex{
		Name:      allTargetKeyword,
		DependsOn: vertexes,
	}, nil
}

func (service *buildService) findTarget(target string, definition builddefinition.Definition) (api.Vertex, error) {
	vertex := definition.Vertex(target)
	if !maybe.Valid(vertex) {
		return api.Vertex{}, errors.Errorf("target %s not found", target)
	}
	return maybe.Just(vertex), nil
}

func (service *buildService) builderAPI(v string) api.BuilderAPI {
	switch v {
	case version.APIVersionV1:
		return service.builderv1
	case version.APIVersionV2:
		return service.builderv2
	default:
		panic(fmt.Sprintf("unknown apiVersion %s", v))
	}
}
