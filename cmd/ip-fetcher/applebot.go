package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/applebot"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func applebotCmd() *cli.Command {
	const (
		providerName = "applebot"
		fileName     = "applebot.json"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Applebot prefixes",
		Usage:     "Apple's web crawler",
		UsageText: "ip-fetcher applebot {--stdout | --Path FILE}",
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			_ = cli.ShowSubcommandHelp(cCtx)

			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagPath,
				Usage: usageWhereToSaveFile, Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  flagStdout,
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			path, stdout, err := resolveOutputTargets(c)
			if err != nil {
				return err
			}

			a := applebot.New()

			if isEnvEnabled("IP_FETCHER_MOCK_APPLEBOT") {
				defer gock.Off()
				urlBase := applebot.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/applebot/testdata/applebot.json")
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
