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
		providerName  = "googleutf"
		fileName      = "user-triggered-fetchers.json"
		fileNameLines = "googleutf-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Google User Triggered Fetchers prefixes",
		Usage:     "Google User Triggered Fetchers",
		UsageText: "ip-fetcher googleutf {--stdout | --Path FILE} [--lines]",
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

			var data []byte
			if c.Bool(formatLines) {
				var doc googleutf.Doc
				if doc, err = g.Fetch(); err != nil {
					return err
				}

				if data, err = docToLines(doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = g.FetchData()
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
