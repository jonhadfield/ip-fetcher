package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/mullvad"

	"github.com/urfave/cli/v2"
)

func mullvadCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "mullvad",
		helpName:  "Mullvad relay addresses",
		usage:     "Mullvad",
		dataFile:  "mullvad.json",
		linesFile: "mullvad-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_MULLVAD",
		mocks: []mockSource{
			{mullvad.DownloadURL, "../../providers/mullvad/testdata/mullvad.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := mullvad.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
