package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/okta"

	"github.com/urfave/cli/v2"
)

func oktaCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "okta",
		helpName:  "Okta identity service prefixes",
		usage:     "Okta (identity service, grouped by cell)",
		dataFile:  "okta.json",
		linesFile: "okta-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_OKTA",
		mocks: []mockSource{
			{okta.DownloadURL, "../../providers/okta/testdata/ip_ranges.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := okta.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
