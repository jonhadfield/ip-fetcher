package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/googleuta"

	"github.com/urfave/cli/v2"
)

func googleutaCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "googleuta",
		helpName:  "Google User-Triggered Agents prefixes",
		usage:     "Google User-Triggered Agents",
		dataFile:  "googleuta.json",
		linesFile: "googleuta-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_GOOGLEUTA",
		mocks: []mockSource{
			{googleuta.DownloadURL, "../../providers/googleuta/testdata/user-triggered-agents.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := googleuta.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
