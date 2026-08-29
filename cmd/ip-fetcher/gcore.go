package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/gcore"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func gcoreCmd() *cli.Command {
	const (
		providerName  = "gcore"
		fileNameData  = "gcore.json"
		fileNameLines = "gcore-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Gcore CDN edge addresses",
		Usage:     "Gcore CDN",
		UsageText: "ip-fetcher gcore {--stdout | --Path FILE} [--lines]",
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

			p := gcore.New()

			if isEnvEnabled("IP_FETCHER_MOCK_GCORE") {
				defer gock.Off()

				u, _ := url.Parse(gcore.DownloadURL)
				gock.New(gcore.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/gcore/testdata/public-ip-list.json")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := gcoreData(c, &p)
			if err != nil {
				return err
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

// gcoreData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func gcoreData(c *cli.Context, p *gcore.Gcore) ([]byte, error) {
	if c.Bool(formatLines) {
		doc, err := p.Fetch()
		if err != nil {
			return nil, err
		}

		return docToLines(doc)
	}

	data, _, _, err := p.FetchData()

	return data, err
}
