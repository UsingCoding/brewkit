package main

import (
	"path"

	"github.com/urfave/cli/v2"

	buildapp "github.com/ispringtech/brewkit/internal/backend/app/build"
	"github.com/ispringtech/brewkit/internal/backend/app/buildlegacy"
	"github.com/ispringtech/brewkit/internal/backend/infrastructure/buildkitd"
	"github.com/ispringtech/brewkit/internal/backend/infrastructure/docker"
	"github.com/ispringtech/brewkit/internal/backend/infrastructure/ssh"
	"github.com/ispringtech/brewkit/internal/frontend/app/buildconfig"
	"github.com/ispringtech/brewkit/internal/frontend/app/builddefinition"
	"github.com/ispringtech/brewkit/internal/frontend/app/service"
	infrabuilddefinition "github.com/ispringtech/brewkit/internal/frontend/infrastructure/builddefinition"
)

func build(workdir string) *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build project from build definition",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "definition",
				Usage:   "Config with build definition",
				Aliases: []string{"d"},
				Value:   path.Join(workdir, buildconfig.DefaultName),
				EnvVars: []string{"BREWKIT_BUILD_CONFIG"},
			},
			&cli.StringFlag{
				Name:    "context",
				Usage:   "Local build context",
				Aliases: []string{"c"},
				Value:   workdir,
				EnvVars: []string{"BREWKIT_CONTEXT"},
			},
			&cli.BoolFlag{
				Name:    "force-pull",
				Usage:   "Always pull a newer version of images for targets",
				Aliases: []string{"p"},
				EnvVars: []string{"BREWKIT_FORCE_PULL"},
			},
		},
		Action: executeBuild,
		Subcommands: []*cli.Command{
			{
				Name:   "definition",
				Usage:  "Print full parsed and verified build definition",
				Action: executeBuildDefinition,
			},
			{
				Name:   "definition-debug",
				Usage:  "Print compiled build definition in raw JSON, useful for debugging complex build definitions",
				Action: executeCompileBuildDefinition,
			},
		},
	}
}

type buildOps struct {
	commonOpt
	BuildDefinition string
	Context         string
	ForcePull       bool
}

func (o *buildOps) scan(ctx *cli.Context) {
	o.commonOpt.scan(ctx)
	o.BuildDefinition = ctx.String("definition")
	o.Context = ctx.String("context")
	o.ForcePull = ctx.Bool("force-pull")
}

func executeBuild(ctx *cli.Context) error {
	var opts buildOps
	opts.scan(ctx)

	buildService, err := makeBuildService(opts)
	if err != nil {
		return err
	}

	return buildService.Build(ctx.Context, service.BuildParams{
		Targets:         ctx.Args().Slice(),
		BuildDefinition: opts.BuildDefinition,
		Context:         opts.Context,
		ForcePull:       opts.ForcePull,
	})
}

func executeBuildDefinition(ctx *cli.Context) error {
	var opts buildOps
	opts.scan(ctx)

	logger := makeLogger(opts.verbose)

	buildService, err := makeBuildService(opts)
	if err != nil {
		return err
	}

	buildDefinition, err := buildService.DumpBuildDefinition(ctx.Context, opts.BuildDefinition)
	if err != nil {
		return err
	}

	logger.Outputf(buildDefinition)

	return nil
}

func executeCompileBuildDefinition(ctx *cli.Context) error {
	var opts buildOps
	opts.scan(ctx)

	logger := makeLogger(opts.verbose)

	buildService, err := makeBuildService(opts)
	if err != nil {
		return err
	}

	buildDefinition, err := buildService.DumpCompiledBuildDefinition(ctx.Context, opts.BuildDefinition)
	if err != nil {
		return err
	}

	logger.Outputf(buildDefinition)

	return nil
}

func makeBuildService(options buildOps) (service.BuildService, error) {
	parser := infrabuilddefinition.Parser{}

	logger := makeLogger(options.verbose)

	config, err := parseConfig(options.configPath, logger)
	if err != nil {
		return nil, err
	}

	dockerClient, err := docker.NewClient(options.dockerClientConfigPath, logger)
	if err != nil {
		return nil, err
	}

	agentProvider := ssh.NewAgentProvider()

	buildLegacyService := buildlegacy.NewBuildService(
		dockerClient,
		DockerfileImage,
		agentProvider,
		logger,
	)

	buildService := buildapp.NewService(
		buildkitd.NewConnector(),
		agentProvider,
		options.verbose,
	)

	return service.NewBuildService(
		parser,
		builddefinition.NewBuilder(),
		buildLegacyService,
		buildService,
		config,
	), nil
}
