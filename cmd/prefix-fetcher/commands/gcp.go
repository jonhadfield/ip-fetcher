package commands

import (
	"fmt"
	"github.com/jonhadfield/prefix-fetcher/gcp"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func GCPCmd() *cli.Command {
	return &cli.Command{
		Name:      "gcp",
		HelpName:  "- fetch GCP prefixes",
		UsageText: "prefix-fetcher azure {--stdout | --path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			// nolint:errcheck
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "where to save the file", Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: "write to stdout", Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("path"))
			if path == "" && !c.Bool("stdout") {
				// nolint:errcheck
				_ = cli.ShowSubcommandHelp(c)

				fmt.Println("\nerror: must specify at least one of stdout and path")
				os.Exit(1)
			}

			a := gcp.New()
			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			if path != "" {
				if err = saveFile(saveFileInput{
					provider:        "gcp",
					data:            data,
					path:            path,
					defaultFileName: "cloud.json",
				}); err != nil {
					return err
				}
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}
}