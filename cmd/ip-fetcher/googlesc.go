package main

import (
	"net/http"
	"net/url"

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
				Usage: usageWhereToSaveFile, Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			g := googlesc.New()

			if isEnvEnabled("IP_FETCHER_MOCK_GOOGLESC") {
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

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: fileName,
				Data:            data,
			})
		},
	}
}
