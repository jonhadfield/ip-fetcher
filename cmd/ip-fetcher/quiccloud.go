package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/quiccloud"

	"github.com/urfave/cli/v2"
)

func quiccloudCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "quiccloud",
		helpName:  "QUIC.cloud CDN prefixes",
		usage:     "QUIC.cloud",
		dataFile:  "quiccloud.txt",
		linesFile: "quiccloud-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_QUICCLOUD",
		mocks: []mockSource{
			{quiccloud.DownloadURL, "../../providers/quiccloud/testdata/quiccloud.html"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := quiccloud.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
