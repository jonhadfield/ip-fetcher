package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/feodo"

	"github.com/urfave/cli/v2"
)

func feodoCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "feodo",
		helpName:  "abuse.ch Feodo Tracker botnet C2 addresses",
		usage:     "abuse.ch Feodo Tracker (botnet command and control servers)",
		dataFile:  "feodo.txt",
		linesFile: "feodo-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_FEODO",
		mocks: []mockSource{
			{feodo.DownloadURL, "../../providers/feodo/testdata/feodo.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := feodo.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
