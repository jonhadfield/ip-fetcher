package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/airvpn"

	"github.com/urfave/cli/v2"
)

func airvpnCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "airvpn",
		helpName:  "AirVPN server ingress addresses",
		usage:     "AirVPN",
		dataFile:  "airvpn.json",
		linesFile: "airvpn-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_AIRVPN",
		mocks: []mockSource{
			{airvpn.DownloadURL, "../../providers/airvpn/testdata/status.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := airvpn.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
