package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/googleutf"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func googleutfCmd() *cli.Command {
	const (
		providerName = "googleutf"
		fileName     = "user-triggered-fetchers.json"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Google User Triggered Fetchers prefixes",
		Usage:     "Google User Triggered Fetchers",
		UsageText: "ip-fetcher googleutf {--stdout | --Path FILE}",
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

			g := googleutf.New()

			if isEnvEnabled("IP_FETCHER_MOCK_GOOGLEUTF") {
				defer gock.Off()
				urlBase := googleutf.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/googleutf/testdata/user-triggered-fetchers.json")
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
