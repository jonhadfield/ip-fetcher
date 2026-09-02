package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/detectify"

	"github.com/urfave/cli/v2"
)

func detectifyCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "detectify",
		helpName:  "Detectify scanner addresses",
		usage:     "Detectify (vulnerability scanning)",
		dataFile:  "detectify.txt",
		linesFile: "detectify-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_DETECTIFY",
		mocks: []mockSource{
			{detectify.DownloadURL, "../../providers/detectify/testdata/scanner-ip-addresses.html"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := detectify.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
