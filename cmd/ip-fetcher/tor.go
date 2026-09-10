package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/tor"

	"github.com/urfave/cli/v2"
)

func torCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "tor",
		helpName:  "Tor exit node addresses",
		usage:     "Tor Exit Nodes (addresses traffic leaves the Tor network from)",
		dataFile:  "tor.txt",
		linesFile: "tor-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_TOR",
		mocks: []mockSource{
			{tor.DownloadURL, "../../providers/tor/testdata/tor.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := tor.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
