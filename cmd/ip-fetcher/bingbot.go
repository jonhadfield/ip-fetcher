package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/bingbot"

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
		UsageText: "ip-fetcher bingbot {--stdout | --Path FILE}",
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

			a := bingbot.New()

			if isEnvEnabled("IP_FETCHER_MOCK_BINGBOT") {
				defer gock.Off()
				urlBase := bingbot.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/bingbot/testdata/bingbot.json")
				gock.InterceptClient(a.Client.HTTPClient)
			}

			data, _, _, err := a.FetchData()
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
