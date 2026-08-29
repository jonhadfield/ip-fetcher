package main

import (
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/newrelic"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func newrelicCmd() *cli.Command {
	const (
		providerName  = "newrelic"
		fileNameData  = "newrelic.json"
		fileNameLines = "newrelic-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch New Relic Synthetics ranges",
		Usage:     "New Relic Synthetics (public monitor locations)",
		UsageText: "ip-fetcher newrelic {--stdout | --Path FILE} [--lines]",
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

			p := newrelic.New()

			if isEnvEnabled("IP_FETCHER_MOCK_NEWRELIC") {
				defer gock.Off()

				u, _ := url.Parse(newrelic.DownloadURL)
				gock.New(newrelic.DownloadURL).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/newrelic/testdata/ip-ranges.json")
				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := newrelicData(c, &p)
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

// newrelicData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func newrelicData(c *cli.Context, p *newrelic.NewRelic) ([]byte, error) {
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
