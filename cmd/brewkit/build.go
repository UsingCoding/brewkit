package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	buildapp "github.com/ispringtech/brewkit/internal/backend/app/build"
	"github.com/ispringtech/brewkit/internal/backend/app/buildlegacy"
	"github.com/ispringtech/brewkit/internal/backend/infrastructure/buildkitd"
	"github.com/ispringtech/brewkit/internal/backend/infrastructure/docker"
	"github.com/ispringtech/brewkit/internal/backend/infrastructure/ssh"
	"github.com/ispringtech/brewkit/internal/common/maybe"
	"github.com/ispringtech/brewkit/internal/common/slices"
	"github.com/ispringtech/brewkit/internal/frontend/app/buildconfig"
	"github.com/ispringtech/brewkit/internal/frontend/app/builddefinition"
	"github.com/ispringtech/brewkit/internal/frontend/app/service"
	infrabuilddefinition "github.com/ispringtech/brewkit/internal/frontend/infrastructure/builddefinition"
)

const (
	targetPrefix = "+"
)

func build(workdir string) *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build definition manipulation",
		BashComplete: func(ctx *cli.Context) {
			// print default completion
			cli.DefaultCompleteWithFlags(ctx.Command)(ctx)

			buildService, opts, err := makeBuildServiceCtx(ctx)
			if err != nil {
				return
			}

			targets, err := buildService.ListTargets(opts.BuildDefinition)
			if err != nil {
				return
			}

			writer := ctx.App.Writer
			zsh := strings.HasSuffix(os.Getenv("SHELL"), "zsh")

			for _, target := range targets {
				// print target name to stdout
				name := targetPrefix + target

				if zsh {
					// add grouping information for zsh
					name += ":Build target"
				}

				_, _ = fmt.Fprintf(writer, "%s\n", name)
			}
		},
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
			&cli.StringFlag{
				Name:        "progress",
				Usage:       "Progress modes: auto, plain, rawjson, quiet",
				DefaultText: "auto",
				EnvVars:     []string{"BREWKIT_BUILD__PROGRESS"},
			},
			&cli.BoolFlag{
				Name:    "force-pull",
				Usage:   "Always pull a newer version of images",
				Aliases: []string{"p"},
				EnvVars: []string{"BREWKIT_FORCE_PULL"},
			},
			&cli.BoolFlag{
				Name:    "disable-progress-grouping",
				Usage:   "Disable progress grouping for llb solving",
				EnvVars: []string{"BREWKIT_DISABLE_PROGRESS_GROUPING"},
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

type buildOpt struct {
	commonOpt
	BuildDefinition string
	Context         string
	ForcePull       bool

	Progress                string
	DisableProgressGrouping bool
}

func (o *buildOpt) scan(ctx *cli.Context) {
	o.commonOpt.scan(ctx)
	o.BuildDefinition = ctx.String("definition")
	o.Context = ctx.String("context")
	o.ForcePull = ctx.Bool("force-pull")
	o.Progress = ctx.String("progress")
	o.DisableProgressGrouping = ctx.Bool("disable-progress-grouping")
}

func executeBuild(ctx *cli.Context) error {
	var opts buildOpt
	opts.scan(ctx)

	targets, err := normalizeTargets(ctx.Args().Slice())
	if err != nil {
		return err
	}

	buildService, err := makeBuildService(opts)
	if err != nil {
		return err
	}

	return buildService.Build(ctx.Context, service.BuildParams{
		Targets:         targets,
		BuildDefinition: opts.BuildDefinition,
		Context:         opts.Context,
		ForcePull:       opts.ForcePull,
	})
}

func executeBuildDefinition(ctx *cli.Context) error {
	var opts buildOpt
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
	var opts buildOpt
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

func makeBuildServiceCtx(ctx *cli.Context) (service.BuildService, buildOpt, error) {
	var opts buildOpt
	opts.scan(ctx)

	buildService, err := makeBuildService(opts)
	return buildService, opts, err
}

func makeBuildService(options buildOpt) (service.BuildService, error) {
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
		buildapp.ServiceParams{
			DisableProgressGrouping: options.DisableProgressGrouping,
			ProgressMode:            options.Progress,
		},
		options.verbose,
	)

	return service.NewBuildService(
		infrabuilddefinition.Parser{},
		builddefinition.NewBuilder(),
		buildLegacyService,
		buildService,
		config,
	), nil
}

func normalizeTargets(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}

	// if first target starts with targetPrefix so each target *must* starts with targetPrefix
	if !strings.HasPrefix(args[0], targetPrefix) {
		// return targets as it is
		return args, nil
	}

	argWithoutPrefix := slices.Find(args, func(a string) bool {
		return !strings.HasPrefix(a, targetPrefix)
	})

	if maybe.Valid(argWithoutPrefix) {
		return nil, errors.Errorf(
			"arg without %s prefix used as target name: %s",
			targetPrefix,
			maybe.Just(argWithoutPrefix),
		)
	}

	// clear targetPrefix
	return slices.Map(args, func(a string) string {
		return strings.TrimLeft(a, targetPrefix)
	}), nil
}
