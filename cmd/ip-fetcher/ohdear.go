package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/ohdear"

	"github.com/urfave/cli/v2"
)

func ohdearCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "ohdear",
		helpName:  "Oh Dear uptime monitoring addresses",
		usage:     "Oh Dear",
		dataFile:  "ohdear.txt",
		linesFile: "ohdear-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_OHDEAR",
		mocks: []mockSource{
			{ohdear.DownloadURL, "../../providers/ohdear/testdata/ips.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := ohdear.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
