package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/xpanse"

	"github.com/urfave/cli/v2"
)

func xpanseCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "xpanse",
		helpName:  "Cortex Xpanse scanner addresses",
		usage:     "Cortex Xpanse (internet scanning)",
		dataFile:  "xpanse.txt",
		linesFile: "xpanse-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_XPANSE",
		mocks: []mockSource{
			{xpanse.DownloadURL, "../../providers/xpanse/testdata/scanning-activity.html"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := xpanse.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
