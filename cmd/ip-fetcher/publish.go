package main

import (
	"github.com/jonhadfield/ip-fetcher/publisher"
	"github.com/urfave/cli/v2"
)

func publishCmd() *cli.Command {
	return &cli.Command{
		Name:      "publish",
		Usage:     "publishes the data to a remote location",
		HelpName:  "- fetch and deploy ranges to a git repo",
		UsageText: "ip-fetcher publish",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Action: func(c *cli.Context) error {
			publisher.Publish()

			return nil
		},
	}
}
