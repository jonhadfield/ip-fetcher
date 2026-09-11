package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/m365"

	"github.com/urfave/cli/v2"
)

func m365Cmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "m365",
		helpName:  "Microsoft 365 endpoint prefixes",
		usage:     "Microsoft 365 (Exchange, SharePoint and Teams endpoints)",
		dataFile:  "m365.json",
		linesFile: "m365-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_M365",
		mocks: []mockSource{
			{m365.DownloadURL, "../../providers/m365/testdata/worldwide.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := m365.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
