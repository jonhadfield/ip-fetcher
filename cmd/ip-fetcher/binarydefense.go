package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/binarydefense"

	"github.com/urfave/cli/v2"
)

func binarydefenseCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "binarydefense",
		helpName:  "Binary Defense artillery banlist",
		usage:     "Binary Defense (artillery honeypot banlist)",
		dataFile:  "binarydefense.txt",
		linesFile: "binarydefense-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_BINARYDEFENSE",
		mocks: []mockSource{
			{binarydefense.DownloadURL, "../../providers/binarydefense/testdata/banlist.txt"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := binarydefense.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
