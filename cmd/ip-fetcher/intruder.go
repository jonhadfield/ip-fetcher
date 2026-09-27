package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/intruder"

	"github.com/urfave/cli/v2"
)

func intruderCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "intruder",
		helpName:  "Intruder scanner addresses",
		usage:     "Intruder (vulnerability scanning)",
		dataFile:  "intruder.txt",
		linesFile: "intruder-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_INTRUDER",
		mocks: []mockSource{
			{intruder.DownloadURL, "../../providers/intruder/testdata/allowlist.html"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := intruder.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
