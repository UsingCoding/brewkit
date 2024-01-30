package main

import (
	"encoding/json"

	"github.com/urfave/cli/v2"

	appversion "github.com/ispringtech/brewkit/internal/frontend/app/version"
)

func version() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Show brewkit version info",
		Action: executeVersion,
	}
}

func executeVersion(ctx *cli.Context) error {
	var opt commonOpt
	opt.scan(ctx)

	logger := makeLogger(opt.verbose)

	v := struct {
		APIVersions []string `json:"apiVersions"`
		Commit      string   `json:"commit"`
		Dockerfile  string   `json:"dockerfile"`
	}{
		APIVersions: appversion.SupportedVersions(),
		Commit:      Commit,
		Dockerfile:  DockerfileImage,
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	logger.Outputf(string(bytes) + "\n")

	return nil
}
