package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/onetrust"

	"github.com/urfave/cli/v2"
)

func onetrustCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "onetrust",
		helpName:  "OneTrust web scanner prefixes",
		usage:     "OneTrust",
		dataFile:  "onetrust.txt",
		linesFile: "onetrust-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_ONETRUST",
		mocks: []mockSource{
			{onetrust.DownloadURL, "../../providers/onetrust/testdata/ips.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := onetrust.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
