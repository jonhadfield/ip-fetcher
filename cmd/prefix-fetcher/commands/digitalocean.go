package commands

import (
	"fmt"
	"github.com/jonhadfield/prefix-fetcher/digitalocean"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func DigitalOceanCmd() *cli.Command {
	return &cli.Command{
		Name:      "digitalocean",
		HelpName:  "- fetch DigitalOcean prefixes",
		UsageText: "prefix-fetcher digitalocean {--stdout | --path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			// nolint:errcheck
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("path"))
			if path == "" && !c.Bool("stdout") {
				// nolint:errcheck
				_ = cli.ShowSubcommandHelp(c)

				fmt.Println("\nerror: must specify at least one of stdout and path")
				os.Exit(1)
			}

			a := digitalocean.New()
			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			if path != "" {
				if err = saveFile(saveFileInput{
					provider:        "digitalocean",
					data:            data,
					path:            path,
					defaultFileName: "ServiceTags_Public.json",
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