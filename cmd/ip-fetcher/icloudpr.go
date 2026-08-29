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
		fileName      = "prefixes.csv"
		fileNameLines = "icloudpr-prefixes.txt"
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
				Name:  flagPath,
				Usage: usageWhereToSaveFile, Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  flagStdout,
				Usage: usageWriteToStdout, Aliases: []string{"s"},
			},
			&cli.BoolFlag{
				Name:  formatLines,
				Usage: usageLinesOutput,
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

			var data []byte
			if c.Bool(formatLines) {
				var doc icloudpr.Doc
				if doc, err = a.Fetch(); err != nil {
					return err
				}

				if data, err = docToLines(doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = a.FetchData()
				if err != nil {
					return err
				}
			}

			defaultName := fileName
			if c.Bool(formatLines) {
				defaultName = fileNameLines
			}

			return writeOutputs(path, stdout, SaveFileInput{
				Provider:        providerName,
				DefaultFileName: defaultName,
				Data:            data,
			})
		},
	}
}
