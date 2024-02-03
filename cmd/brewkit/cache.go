package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/tonistiigi/units"
	"github.com/urfave/cli/v2"

	"github.com/ispringtech/brewkit/internal/backend/api"
	backendcache "github.com/ispringtech/brewkit/internal/backend/app/cache"
	"github.com/ispringtech/brewkit/internal/backend/infrastructure/buildkitd"
	"github.com/ispringtech/brewkit/internal/frontend/app/service"
)

func cache() *cli.Command {
	return &cli.Command{
		Name:  "cache",
		Usage: "Manipulate brewkit docker cache",
		Subcommands: []*cli.Command{
			cacheClear(),
		},
	}
}

func cacheClear() *cli.Command {
	return &cli.Command{
		Name:  "clear",
		Usage: "Clear docker builder cache",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Clear all cache, not just dangling ones",
			},
			&cli.DurationFlag{
				Name:  "keep-duration",
				Usage: "Keep cache older than",
			},
			&cli.Int64Flag{
				Name:  "keep-bytes",
				Usage: "Keep cache bytes",
			},
		},
		Action: func(ctx *cli.Context) error {
			var opts commonOpt
			opts.scan(ctx)

			cacheAPI := backendcache.NewCacheService(buildkitd.NewConnector())
			cacheService := service.NewCacheService(cacheAPI)

			infos, err := cacheService.ClearCache(ctx.Context, service.ClearCacheParam{
				KeepBytes:    ctx.Int64("keep-bytes"),
				KeepDuration: ctx.Duration("keep-duration"),
				All:          ctx.Bool("all"),
			})
			if err != nil {
				return err
			}

			tw := tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)
			first := true
			total := int64(0)

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case info, ok := <-infos:
					if !ok {
						// Reset tabwriter
						tw = tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)
						fmt.Fprintf(tw, "Total:\t%.2f\n", units.Bytes(total))
						tw.Flush()
						return nil
					}

					if info.Err != nil {
						return err
					}

					total += info.Size
					if first {
						printTableHeader(tw)
						first = false
					}
					printTableRow(tw, info)
					tw.Flush()
				}
			}
		},
	}
}

func printTableHeader(tw *tabwriter.Writer) {
	fmt.Fprintln(tw, "ID\tRECLAIMABLE\tSIZE\tLAST ACCESSED")
}

func printTableRow(tw *tabwriter.Writer, usageInfo api.UsageInfo) {
	id := usageInfo.ID
	if usageInfo.Mutable {
		id += "*"
	}
	size := fmt.Sprintf("%.2f", units.Bytes(usageInfo.Size))
	if usageInfo.Shared {
		size += "*"
	}
	fmt.Fprintf(tw, "%-71s\t%-11v\t%s\t\n", id, !usageInfo.InUse, size)
}
