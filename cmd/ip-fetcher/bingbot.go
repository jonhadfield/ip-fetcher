package main

import (
	"fmt"
	"github.com/jonhadfield/ip-fetcher/providers/bingbot"
	"net/url"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func bingbotCmd() *cli.Command {
	const (
		providerName = "bingbot"
		fileName     = "bingbot.json"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Microsoft Bingbot prefixes",
		Usage:     "Bingbot",
		UsageText: "ip-fetcher bingbot {--stdout | --path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
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
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\nerror: must specify at least one of stdout and path")
				os.Exit(1)
			}

			a := bingbot.New()

			if os.Getenv("IP_FETCHER_MOCK_BINGBOT") == "true" {
				defer gock.Off()
				urlBase := bingbot.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(200).
					File("../../providers/bingbot/testdata/bingbot.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			data, _, _, err := a.FetchData()
			if err != nil {
				return err
			}

			if path != "" {
				var out string
				if out, err = saveFile(saveFileInput{
					provider:        providerName,
					data:            data,
					path:            path,
					defaultFileName: fileName,
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
