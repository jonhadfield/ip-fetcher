package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/updown"

	"github.com/urfave/cli/v2"
)

func updownCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "updown",
		helpName:  "updown.io monitoring node addresses",
		usage:     "updown.io (monitoring nodes)",
		dataFile:  "updown.json",
		linesFile: "updown-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_UPDOWN",
		mocks: []mockSource{
			{updown.DownloadURL, "../../providers/updown/testdata/nodes.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := updown.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
