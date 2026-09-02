package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/tenable"

	"github.com/urfave/cli/v2"
)

func tenableCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "tenable",
		helpName:  "Tenable cloud scanner prefixes",
		usage:     "Tenable Cloud Scanners (vulnerability scanning)",
		dataFile:  "tenable.json",
		linesFile: "tenable-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_TENABLE",
		mocks: []mockSource{
			{tenable.DownloadURL, "../../providers/tenable/testdata/data.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := tenable.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
