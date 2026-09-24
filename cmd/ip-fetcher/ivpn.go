package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/ivpn"

	"github.com/urfave/cli/v2"
)

func ivpnCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "ivpn",
		helpName:  "IVPN server ingress addresses",
		usage:     "IVPN",
		dataFile:  "ivpn.json",
		linesFile: "ivpn-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_IVPN",
		mocks: []mockSource{
			{ivpn.DownloadURL, "../../providers/ivpn/testdata/servers.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := ivpn.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
