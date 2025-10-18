package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/icloudpr"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

const (
	SICloudPR = "icloudpr"
)

func iCloudPRCmd() *cli.Command {
	const (
		fileName = "prefixes.csv"
	)

	return &cli.Command{
		Name:      SICloudPR,
		HelpName:  "- fetch iCloud Private Relay prefixes",
		Usage:     "iCloud Private Relay prefixes",
		UsageText: "ip-fetcher icloudpr {--stdout | --Path FILE}",
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

			a := icloudpr.New()

			if isEnvEnabled("IP_FETCHER_MOCK_ICLOUDPR") {
				defer gock.Off()
				urlBase := icloudpr.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/icloudpr/testdata/egress-ip-ranges.csv")
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
