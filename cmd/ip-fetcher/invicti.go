package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/invicti"

	"github.com/urfave/cli/v2"
)

func invictiCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "invicti",
		helpName:  "Invicti scanner addresses",
		usage:     "Invicti (vulnerability scanning)",
		dataFile:  "invicti.txt",
		linesFile: "invicti-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_INVICTI",
		mocks: []mockSource{
			{invicti.USDownloadURL, "../../providers/invicti/testdata/trustlist-us.html"},
			{invicti.EUDownloadURL, "../../providers/invicti/testdata/trustlist-eu.html"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := invicti.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
