package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/hetrixtools"

	"github.com/urfave/cli/v2"
)

func hetrixtoolsCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "hetrixtools",
		helpName:  "HetrixTools uptime monitoring addresses",
		usage:     "HetrixTools",
		dataFile:  "hetrixtools.txt",
		linesFile: "hetrixtools-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_HETRIXTOOLS",
		mocks: []mockSource{
			{hetrixtools.DownloadURL, "../../providers/hetrixtools/testdata/ips.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := hetrixtools.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
