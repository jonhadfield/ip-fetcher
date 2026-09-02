package main

import (
	"net/http"

	"github.com/jonhadfield/ip-fetcher/providers/uptrends"

	"github.com/urfave/cli/v2"
)

func uptrendsCmd() *cli.Command {
	return providerCommand(providerSpec{
		name:      "uptrends",
		helpName:  "Uptrends checkpoint addresses",
		usage:     "Uptrends (monitoring checkpoints)",
		dataFile:  "uptrends.json",
		linesFile: "uptrends-prefixes.txt",
		mockEnv:   "IP_FETCHER_MOCK_UPTRENDS",
		mocks: []mockSource{
			{uptrends.IPv4URL, "../../providers/uptrends/testdata/ipv4.json"},
			{uptrends.IPv6URL, "../../providers/uptrends/testdata/ipv6.json"},
		},
		newProvider: func() (func() ([]byte, http.Header, int, error), func() (any, error), *http.Client) {
			p := uptrends.New()

			return p.FetchData, func() (any, error) { return p.Fetch() }, p.Client.HTTPClient
		},
	})
}
