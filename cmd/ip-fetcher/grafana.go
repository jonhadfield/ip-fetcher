package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jonhadfield/ip-fetcher/providers/grafana"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gock.v1"
)

func grafanaCmd() *cli.Command {
	const (
		providerName  = "grafana"
		fileNameData  = "grafana.json"
		fileNameLines = "grafana-prefixes.txt"
	)

	return &cli.Command{
		Name:      providerName,
		HelpName:  "- fetch Grafana synthetic monitoring probe prefixes",
		Usage:     "Grafana Synthetic Monitoring (public probes)",
		UsageText: "ip-fetcher grafana {--stdout | --Path FILE} [--lines]",
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

			p := grafana.New()

			if isEnvEnabled("IP_FETCHER_MOCK_GRAFANA") {
				defer gock.Off()

				u, _ := url.Parse(grafana.DownloadURL)
				gock.New(fmt.Sprintf("%s://%s", u.Scheme, u.Host)).
					Get(u.Path).
					Reply(http.StatusOK).
					File("../../providers/grafana/testdata/synthetics.json")

				gock.InterceptClient(p.Client.HTTPClient)
			}

			data, err := grafanaData(c, &p)
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

// grafanaData returns newline separated prefixes when --lines is set, and the
// upstream document otherwise.
func grafanaData(c *cli.Context, p *grafana.Grafana) ([]byte, error) {
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
