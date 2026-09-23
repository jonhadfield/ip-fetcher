package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/ipsum"

	"github.com/urfave/cli/v2"
)

func ipsumCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "ipsum",
		helpName:  "IPsum consensus threat list (level 3)",
		usage:     "IPsum (addresses on 3+ blacklists)",
		dataFile:  "ipsum.txt",
		linesFile: "ipsum-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_IPSUM",
		mocks: []mockSource{
			{ipsum.DownloadURL, "../../providers/ipsum/testdata/ipsum.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := ipsum.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
