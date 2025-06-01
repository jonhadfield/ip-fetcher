package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jonhadfield/ip-fetcher/providers/googlesc"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func googlescCmd() *cli.Command {
	const (
		providerName = "googlesc"
		fileName     = "special-crawlers.json"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Google Special Crawlers prefixes",
		Usage:     "Google Special Crawlers",
		UsageText: "ip-fetcher googlesc {--stdout | --Path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "Path",
				Usage: "where to save the file", Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: "write to stdout", Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path := strings.TrimSpace(c.String("Path"))
			if path == "" && !c.Bool("stdout") {
				_ = cli.ShowSubcommandHelp(c)
				fmt.Println("\nerror: must specify at least one of stdout and Path")
				os.Exit(1)
			}

			g := googlesc.New()

			if os.Getenv("IP_FETCHER_MOCK_GOOGLESC") == "true" {
				defer gock.Off()
				urlBase := googlesc.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/googlesc/testdata/special-crawlers.json")
				gock.InterceptClient(g.Client.HTTPClient)
			}

			data, _, _, err := g.FetchData()
			if err != nil {
				return err
			}

			if path != "" {
				var out string
				if out, err = SaveFile(SaveFileInput{
					Provider:        providerName,
					Data:            data,
					Path:            path,
					DefaultFileName: fileName,
				}); err != nil {
					return err
				}

				_, _ = os.Stderr.WriteString(fmt.Sprintf("Data written to %s\n", out))
			}

			if c.Bool("stdout") {
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}
}
