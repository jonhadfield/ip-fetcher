package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/google"
	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func googleCmd() *cli.Command {
	const (
		providerName = "google"
		fileName     = "goog.json"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch GOOGLE prefixes",
		Usage:     "Google",
		UsageText: "ip-fetcher google {--stdout | --Path FILE}",
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

			a := google.New()

			if isEnvEnabled("IP_FETCHER_MOCK_GOOGLE") {
				defer gock.Off()
				urlBase := google.DownloadURL
				u, _ := url.Parse(urlBase)
				gock.New(urlBase).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/google/testdata/goog.json")
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
