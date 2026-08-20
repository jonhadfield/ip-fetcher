package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/dshield"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func dshieldCmd() *cli.Command {
	const (
		providerName  = "dshield"
		fileNameData  = "dshield.txt"
		fileNameLines = "dshield-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch DShield recommended block list",
		Usage:     "DShield Recommended Block List (SANS Internet Storm Center)",
		UsageText: "ip-fetcher dshield {--stdout | --Path FILE} [--lines]",
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

			p := dshield.New()

			if isEnvEnabled("IP_FETCHER_MOCK_DSHIELD") {
				defer gock.Off()

				u, _ := url.Parse(dshield.DownloadURL)
				gock.New(dshield.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/dshield/testdata/block.txt")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc dshield.Doc
				if doc, err = p.Fetch(); err != nil {
					return err
				}

				if data, err = docToLines(doc); err != nil {
					return err
				}
			} else {
				data, _, _, err = p.FetchData()
				if err != nil {
					return err
				}
			}

			defaultName := fileNameData
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
