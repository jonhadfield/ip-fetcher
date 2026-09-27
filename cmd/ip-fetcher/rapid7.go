package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/rapid7"

	"github.com/urfave/cli/v2"
)

func rapid7Cmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "rapid7",
		helpName:  "Rapid7 InsightAppSec scanner addresses",
		usage:     "Rapid7 InsightAppSec (vulnerability scanning)",
		dataFile:  "rapid7.txt",
		linesFile: "rapid7-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_RAPID7",
		mocks: []mockSource{
			{rapid7.DownloadURL, "../../providers/rapid7/testdata/allowlist-cloud-engine-ips.html"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := rapid7.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
