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
		providerName  = "google"
		fileName      = "goog.json"
		fileNameLines = "google-prefixes.txt"
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

			var data []byte
			if c.Bool(formatLines) {
				var doc google.Doc
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
