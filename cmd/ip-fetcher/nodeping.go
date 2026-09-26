package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/nodeping"

	"github.com/urfave/cli/v2"
)

func nodepingCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "nodeping",
		helpName:  "NodePing uptime monitoring probe addresses",
		usage:     "NodePing",
		dataFile:  "nodeping.txt",
		linesFile: "nodeping-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_NODEPING",
		mocks: []mockSource{
			{nodeping.DownloadURL, "../../providers/nodeping/testdata/pinghosts.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := nodeping.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
