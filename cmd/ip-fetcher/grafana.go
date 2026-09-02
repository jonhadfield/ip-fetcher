package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/grafana"

	"github.com/urfave/cli/v2"
)

func grafanaCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "grafana",
		helpName:  "Grafana synthetic monitoring probe prefixes",
		usage:     "Grafana Synthetic Monitoring (public probes)",
		dataFile:  "grafana.json",
		linesFile: "grafana-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_GRAFANA",
		mocks: []mockSource{
			{grafana.DownloadURL, "../../providers/grafana/testdata/synthetics.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := grafana.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
