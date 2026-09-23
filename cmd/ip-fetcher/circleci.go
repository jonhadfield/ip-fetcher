package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/circleci"

	"github.com/urfave/cli/v2"
)

func circleciCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "circleci",
		helpName:  "CircleCI job IP ranges",
		usage:     "CircleCI",
		dataFile:  "circleci.json",
		linesFile: "circleci-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_CIRCLECI",
		mocks: []mockSource{
			{circleci.DownloadURL, "../../providers/circleci/testdata/circleci.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := circleci.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
