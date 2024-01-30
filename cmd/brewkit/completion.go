package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	data "github.com/ispringtech/brewkit/data/autocompletion"
)

func completion() *cli.Command {
	return &cli.Command{
		Name:  "completion",
		Usage: "Generate autocompletion",
		Subcommands: []*cli.Command{
			{
				Name:   "bash",
				Usage:  "Generate autocompletion for bash",
				Action: executeCompletionBash,
			},
			{
				Name:   "zsh",
				Usage:  "Generate autocompletion for zsh",
				Action: executeCompletionZsh,
			},
		},
	}
}

func executeCompletionBash(*cli.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, data.Bash(appID))
	return nil
}

func executeCompletionZsh(*cli.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, data.Zsh(appID))
	return nil
}
