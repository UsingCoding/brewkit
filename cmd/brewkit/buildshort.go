package main

import (
	"github.com/urfave/cli/v2"
)

func buildShort(workdir string) *cli.Command {
	buildCmd := build(workdir)

	return &cli.Command{
		Name:         "b",
		Usage:        "Build project shortcut",
		UsageText:    "Build project from definition: brewkit b +gobuild",
		HideHelp:     true,
		BashComplete: buildCmd.BashComplete,
		Action:       buildCmd.Action,
		Flags:        buildCmd.Flags,
	}
}
