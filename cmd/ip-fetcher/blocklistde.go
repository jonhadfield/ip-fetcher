package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/blocklistde"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func blocklistdeCmd() *cli.Command {
	const (
		providerName  = "blocklistde"
		fileNameData  = "blocklistde.txt"
		fileNameLines = "blocklistde-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Blocklist.de reported attacker addresses",
		Usage:     "Blocklist.de (addresses reported for attacks in the last 48 hours)",
		UsageText: "ip-fetcher blocklistde {--stdout | --Path FILE} [--lines]",
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

			p := blocklistde.New()

			if isEnvEnabled("IP_FETCHER_MOCK_BLOCKLISTDE") {
				defer gock.Off()

				u, _ := url.Parse(blocklistde.DownloadURL)
				gock.New(blocklistde.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/blocklistde/testdata/all.txt")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			var data []byte
			if c.Bool(formatLines) {
				var doc blocklistde.Doc
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
