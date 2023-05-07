package main

import (
	"fmt"
	"github.com/jonhadfield/ip-fetcher/providers/abuseipdb"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func abuseipdbCmd() *cli.Command {
	return &cli.Command{
		Name:      "abuseipdb",
		HelpName:  "- fetch AbuseIPDB prefixes",
		Usage:     "AbuseIPDB",
		UsageText: "ip-fetcher abuseipdb --key {--stdout | --path FILE} [--confidence] [--limit]",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "key",
				Usage: "api key", Aliases: []string{"k"}, Required: true,
			},
			&cli.IntFlag{
				Name:  "confidence",
				Usage: "minimum confidence percentage score to return", Value: 75, Aliases: []string{"c"},
			},
			&cli.Int64Flag{
				Name:  "limit",
				Usage: "maximum number of results to return", Value: 1000, Aliases: []string{"l"},
			},
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
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\nerror: must specify at least one of stdout and path")
				os.Exit(1)
			}

			a := abuseipdb.New()
			a.Limit = c.Int64("limit")
			a.APIKey = c.String("key")
			a.ConfidenceMinimum = c.Int("confidence")
			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			if path != "" {
				var out string
				if out, err = saveFile(saveFileInput{
					provider:        "abuseipdb",
					data:            data,
					path:            path,
					defaultFileName: "blacklist",
				}); err != nil {
					return err
				}

				_, _ = os.Stderr.WriteString(fmt.Sprintf("data written to %s\n", out))
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}
}
