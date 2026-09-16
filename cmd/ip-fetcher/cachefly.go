package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/cachefly"

	"github.com/urfave/cli/v2"
)

func cacheflyCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "cachefly",
		helpName:  "CacheFly CDN prefixes",
		usage:     "CacheFly",
		dataFile:  "cachefly.txt",
		linesFile: "cachefly-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_CACHEFLY",
		mocks: []mockSource{
			{cachefly.DownloadURL, "../../providers/cachefly/testdata/cachefly.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := cachefly.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
