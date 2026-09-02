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
		Name:         providerName,
		HelpName:     "- fetch Grafana synthetic monitoring probe prefixes",
		Usage:        "Grafana Synthetic Monitoring (public probes)",
		UsageText:    "ip-fetcher grafana {--stdout | --Path FILE} [--lines]",
		OnUsageError: onUsageError,
		Flags:        providerFlags(),
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

			data, err := providerData(c, p.FetchData, func() (any, error) { return p.Fetch() })
			if err != nil {
				return err
			}

			return writeProviderOutputs(c, path, stdout, providerName, fileNameData, fileNameLines, data)
		},
	}
}
