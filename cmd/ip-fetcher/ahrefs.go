package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/ahrefs"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func ahrefsCmd() *cli.Command {
	const (
		providerName = "ahrefs"
		fileName     = "ahrefs.json"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch AhrefsBot prefixes",
		Usage:     "Ahrefs' web crawler",
		UsageText: "ip-fetcher ahrefs {--stdout | --Path FILE}",
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

			a := ahrefs.New()

			if isEnvEnabled("IP_FETCHER_MOCK_AHREFS") {
				defer gock.Off()
				urlBase := ahrefs.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/ahrefs/testdata/ahrefs.json")
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
