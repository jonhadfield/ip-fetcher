package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/googleutfg"

	"github.com/urfave/cli/v2"
)

func googleutfgCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "googleutfg",
		helpName:  "Google User-Triggered Fetchers (Google) prefixes",
		usage:     "Google User-Triggered Fetchers (Google)",
		dataFile:  "googleutfg.json",
		linesFile: "googleutfg-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_GOOGLEUTFG",
		mocks: []mockSource{
			{googleutfg.DownloadURL, "../../providers/googleutfg/testdata/user-triggered-fetchers-google.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := googleutfg.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
